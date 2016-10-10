package quiz

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

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
	lastQnId string
}

func New(uid string, qns []Question) Candidate {
	fmt.Println(len(qns))
	c := Candidate{}
	c.qns = make([]Question, len(qns))
	copy(c.qns, qns)
	UpdateMap(uid, c)
	return c
}

func init() {
	cmap = make(map[string]Candidate)
}

var (
	cmap map[string]Candidate
	mu   sync.RWMutex
)

func UpdateMap(uid string, c Candidate) {
	fmt.Println("in update map", c)
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

func QuestionHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	// TODO - Validate format
	token, err := jwt.ParseWithClaims(s[1], &x.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("AllYourBase"), nil
	})

	userId := ""
	if claims, ok := token.Claims.(*x.Claims); ok {
		userId = claims.UserId
	} else {
		w.Write([]byte("Unauthorized"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	c, err := ReadMap(userId)
	if err != nil {
		w.Write([]byte("Unauthorized"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	x.Debug(c)
	// TODO - Send END qn if len qns == 0.
	qn := c.qns[0]
	// TODO - Write to DB the qn asked.
	fmt.Println("question", qn)
	c.qns = c.qns[1:]
	c.lastQnId = qn.Id
	qn.Score = c.score
	UpdateMap(userId, c)
	b, err := json.Marshal(qn)
	if err != nil {
		w.Write([]byte("Unauthorized"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fmt.Println("marshalled", string(b))
	w.Write(b)
	return
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

type questionMeta struct {
	Negative float64 `json:"negative,string"`
	Positive float64 `json:"positive,string"`
	// TODO - Maybe store correct later as a comma separated string uids.
	Correct []correct `json:"question.correct"`
}

type qmRes struct {
	QuestionMeta []questionMeta `json:"question"`
}

func qnMeta(qid string) (float64, float64, []string) {
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
		log.Fatal("There should be just question returned")
	}
	question := resp.QuestionMeta[0]
	// TODO - Maybe cache this stuff later.
	correctAnswers := []string{}
	for _, answer := range question.Correct {
		correctAnswers = append(correctAnswers, answer.Uid)
	}
	return question.Positive, question.Negative, correctAnswers
}

func calcScore(qid string, selected []string) float64 {
	pos, neg, actual := qnMeta(qid)
	return isCorrectAnswer(selected, actual, pos, neg)
}

func AnswerHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	// TODO - Validate format
	token, err := jwt.ParseWithClaims(s[1], &x.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("AllYourBase"), nil
	})

	userId := ""
	if claims, ok := token.Claims.(*x.Claims); ok {
		userId = claims.UserId
	} else {
		w.Write([]byte("Unauthorized"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	c, err := ReadMap(userId)
	if err != nil {
		w.Write([]byte("Unauthorized"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	qid := r.PostFormValue("qid")
	aid := r.PostFormValue("aid")
	if qid != c.lastQnId {
		// TODO - Return error
	}
	// TODO - Log response to DB.
	answerIds := strings.Split(aid, ",")
	if len(answerIds) == 0 {
		// TODO - Return error
	}
	// TODO - Get qn with correct options from DB and calculate score based on the
	// response.
	score := calcScore(qid, answerIds)
	c.score = score
	UpdateMap(userId, c)
	// TODO - Update score.
	return
}

// type ServerStatus struct {
// 	TimeLeft string `protobuf:"bytes,1,opt,name=timeLeft,proto3" json:"timeLeft,omitempty"`
// 	Status   string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
// }
//
// type ClientStatus struct {
// 	CurQuestion string `protobuf:"bytes,1,opt,name=curQuestion,proto3" json:"curQuestion,omitempty"`
// 	Token       string `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`

// }
