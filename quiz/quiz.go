package quiz

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/gru/admin/mail"
	"github.com/dgraph-io/gru/admin/report"
	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// Map of candidate uids to their quiz info which is stored in Candidate
	// struct.
	cmap     map[string]Candidate
	mu       sync.RWMutex
	throttle chan time.Time
)

type difficulty int

const (
	EASY difficulty = 0
	MEDIUM difficulty = 1
	HARD difficulty = 2

	LEVEL_UP difficulty = +1
	LEVEL_DOWN difficulty = -1

	// EASY, MEDIUM and HARD.
	NumLevels = 3

	rate       = time.Second
)

func init() {
	throttle = make(chan time.Time, 3)
	go rateLimit()
	cmap = make(map[string]Candidate)
}

type Answer struct {
	Uid   string `json:"uid"`
	Text string `json:"name"`
}

// Candidate is used to keep track of the state of the quiz for a candidate.
type Candidate struct {
	name          string
	email         string
	token         string
	score         float64
	qns           map[difficulty][]Question
	lastExchange  time.Time
	quizDuration  time.Duration
	quizCutoff    float64
	quizThreshold float64
	quizStart     time.Time
	validity      time.Time
	// number of questions left.
	numQuestions int
	// current question index.
	qnIdx int

	// We use these so that we can show candidate the same question if he
	// refreshes the page.
	lastQnUid  string
	lastQnCuid string
	// To keep track of time spent on current question.
	lastQnAsked time.Time

	// To keep track of if we have already sent mail about a candidates report
	mailSent bool

	// Difficulty level of questions being asked.
	level difficulty
	// No. of consecutive questions correct or wrong. Used to switch the level.
	streak int
}

func updateMap(uid string, c Candidate) {
	mu.Lock()
	defer mu.Unlock()
	cmap[uid] = c
}

func readMap(uid string) (Candidate, error) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := cmap[uid]
	if !ok {
		return Candidate{}, fmt.Errorf("Uid not found in map.")
	}
	return c, nil
}

type QuizInfo struct {
	Uid       string     `json:"uid"`
	Duration  int        `json:"duration"`
	CutOff    float64    `json:"cut_off,string"`
	Threshold float64    `json:"threshold,string"`
	Questions []Question `json:"quiz.question"`
}

// Used to fetch data about a candidate from Dgraph and populate Candidate struct.
type cand struct {
	Uid         string     `json:"uid"`
	Name        string
	Email       string
	Token       string     `json:"token"`
	Validity    string     `json:"validity"`
	Complete    bool       `json:"complete"`
	CompletedAt time.Time  `json:"completed_at,string"`
	Quiz        []QuizInfo `json:"candidate.quiz"`
	QuizStart   time.Time  `json:"quiz_start"`
}

type CandidateResp struct {
	Cand []cand `json:"quiz.candidate"`
}

type QuizCandidatesResp struct {
	Data struct {
		Cand []cand `json:"quiz.candidate"`
	} `json:"data"`
}

type uid struct {
	Uid string `json:"uid"`
}

type qids struct {
	QuestionUid []uid   `json:"question"`
	Score       float64 `json:"candidate.score"`
	Answered    string  `json:"question.answered"`
}

func timeLeft(start time.Time, dur time.Duration) time.Duration {
	if start.IsZero() {
		return dur
	}
	// If start isn't zero we return the time left.
	return start.Add(dur).Sub(time.Now())
}

// Checks the JWT Token and gets the user id from the claims.
func validateToken(r *http.Request) (string, error) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Bearer" {
		return "", fmt.Errorf("Format of authorization header isn't correct")
	}
	token, err := jwt.ParseWithClaims(s[1], &x.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(*auth.Secret), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*x.Claims); ok && claims.UserId != "" {
		return claims.UserId, nil
	}
	return "", fmt.Errorf("Invalid JWT token")
}

func sendReport(cid string) {
	dir, _ := os.Getwd()
	t, err := template.ParseFiles(filepath.Join(dir, "quiz/report.html"))
	if err != nil {
		fmt.Println("Report not sent", err)
		return
	}
	buf := new(bytes.Buffer)
	s, re := report.ReportSummary(cid)
	if re.Err != "" || re.Msg != "" {
		fmt.Printf("Error: %v with msg: %v while generating report.",
			re.Err, re.Msg)
		return
	}
	if err = t.Execute(buf, s); err != nil {
		fmt.Println(err)
	}
	mail.SendReport(s.Name, s.QuizName, s.TotalScore, s.MaxScore, buf.String())
}

// Used to send mail about the candidate when his test ends.
func sendMail(c Candidate, userId string) error {
	if c.mailSent {
		return nil
	}
	go sendReport(userId)
	c.mailSent = true
	updateMap(userId, c)

	if c.score > c.quizCutoff {
		return nil
	}

	m := new(dgraph.Mutation)
	m.SetString(userId, "completed_at", time.Now().Format(time.RFC3339Nano))
	_, err := dgraph.SendMutation(m)
	return err
}
