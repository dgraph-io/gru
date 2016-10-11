package quiz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
)

type Answer struct {
	Id   string `json:"_uid_"`
	Text string `json:"name"`
}

type Question struct {
	Id      string   `json:"_uid_"`
	Text    string   `json:"text"`
	Options []Answer `json:"question.option"`
	// TODO - Remove the ,string after we incorporate Dgraph schema here.
	IsMultiple bool    `json:"multiple,string"`
	Positive   float64 `json:"positive,string"`
	Negative   float64 `json:"negative,string"`
	// Score of the candidate is sent as part of the questions API.
	Score float64 `json:"score"`
}

type Candidate struct {
	score float64
	qns   []Question
	// Used to check the order of answers.
	lastQnId     string
	lastExchange time.Time
}

func (c Candidate) LastExchange() time.Time {
	return c.lastExchange
}

func New(uid string, qns []Question) Candidate {
	c := Candidate{}
	c.qns = make([]Question, len(qns))
	copy(c.qns, qns)
	UpdateMap(uid, c)
	return c
}

func init() {
	// TODO - Handler server crashes and restarts. That would mean reload cmap from DB.
	cmap = make(map[string]Candidate)
}

var (
	cmap map[string]Candidate
	mu   sync.RWMutex
)

func UpdateMap(uid string, c Candidate) {
	mu.Lock()
	defer mu.Unlock()
	cmap[uid] = c
}

func ReadMap(uid string) (Candidate, error) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := cmap[uid]
	if !ok {
		return Candidate{}, fmt.Errorf("Uid not found in map.")
	}
	return c, nil
}

func validate(r *http.Request) (string, error) {
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

	if claims, ok := token.Claims.(*x.Claims); ok {
		return claims.UserId, nil
	}
	return "", fmt.Errorf("Cannot parse claims.")
}

func QuestionHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)

	var userId string
	var err error
	if userId, err = validate(r); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	c, err := ReadMap(userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("User not found."))
		return
	}

	x.Debug(c)
	// TODO - Send END qn if len qns == 0.
	if len(c.qns) == 0 {
		q := Question{
			Id:    "END",
			Score: c.score,
		}
		b, err := json.Marshal(q)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Unauthorized"))
			return
		}
		w.Write(b)
		return
	}

	qn := c.qns[0]
	// TODO - Write to DB the qn asked.
	c.qns = c.qns[1:]
	c.lastQnId = qn.Id
	UpdateMap(userId, c)
	qn.Score = c.score
	b, err := json.Marshal(qn)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
	w.Write(b)
}

func isCorrectAnswer(selected []string, actual []string, pos, neg float64) float64 {
	if selected[0] == "skip" {
		return 0
	}
	// For multiple choice qnstions, we have partial scoring.
	if len(actual) == 1 {
		if selected[0] == actual[0] {
			return pos
		}
		return -neg
	}
	var score float64
	for _, aid := range selected {
		correct := false
		for _, caid := range actual {
			if caid == aid {
				correct = true
				break
			}
		}
		if correct {
			score += pos
		} else {
			score -= neg
		}
	}
	return score
}

type correct struct {
	Uid string `json:"_uid_"`
}

// Used to marshal response from Dgraph.
type questionMeta struct {
	Negative float64 `json:"negative,string"`
	Positive float64 `json:"positive,string"`
	// TODO - Maybe store correct later as a comma separated string uids so that
	// processing isn't required.
	Correct []correct `json:"question.correct"`
}

type qmRes struct {
	QuestionMeta []questionMeta `json:"question"`
}

type questionCorrectMeta struct {
	negative float64
	positive float64
	correct  []string
}

func qnMeta(qid string) (questionCorrectMeta, error) {
	q := `{
        question(_uid_: ` + qid + `) {
                question.correct {
                _uid_
        }
        positive
        negative
        }
}`
	res := dgraph.Query(q)
	var resp qmRes
	json.Unmarshal(res, &resp)

	if len(resp.QuestionMeta) != 1 {
		return questionCorrectMeta{},
			fmt.Errorf("There should be just one question returned")
	}
	question := resp.QuestionMeta[0]
	// TODO - Maybe cache this stuff later.
	correctAnswers := []string{}
	for _, answer := range question.Correct {
		correctAnswers = append(correctAnswers, answer.Uid)
	}

	return questionCorrectMeta{
		negative: question.Negative,
		positive: question.Positive,
		correct:  correctAnswers,
	}, nil
}

func AnswerHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)

	var userId string
	var err error
	if userId, err = validate(r); err != nil {
		w.Write([]byte("Unauthorized"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	c, err := ReadMap(userId)
	if err != nil {
		w.Write([]byte("User not found."))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	qid := r.PostFormValue("qid")
	aid := r.PostFormValue("aid")
	if qid != c.lastQnId {
		// TODO - Return error
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO - Log response to DB.
	answerIds := strings.Split(aid, ",")
	if len(answerIds) == 0 {
		// TODO - Return error
		w.Write([]byte("Answer ids can't be empty"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m, err := qnMeta(qid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	score := isCorrectAnswer(answerIds, m.correct, m.positive, m.negative)
	c.score = score
	UpdateMap(userId, c)
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)

	var userId string
	var err error
	if userId, err = validate(r); err != nil {
		w.Write([]byte("Unauthorized"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	c, err := ReadMap(userId)
	if err != nil {
		w.Write([]byte("User not found."))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c.lastExchange = time.Now()
	UpdateMap(userId, c)
}

// type ServerStatus struct {
// 	TimeLeft string `protobuf:"bytes,1,opt,name=timeLeft,proto3" json:"timeLeft,omitempty"`
// 	Status   string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
// }
