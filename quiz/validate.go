package quiz

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func rateLimit() {
	rateTicker := time.NewTicker(rate)
	defer rateTicker.Stop()

	for t := range rateTicker.C {
		select {
		case throttle <- t:
		default:
		}
	}
}

// Successfull Response sent as part of the validate API.
type validateRes struct {
	Token    string `json:"token"`
	Duration string `json:"duration"`
	// Whether quiz was already started by the candidate.
	// If this is true the client can just call the questions API and
	// skip showing the instructions page
	Started bool `json:"quiz_started"`
	Name    string
}

func genToken(uid string) (string, error) {
	// Add the user id as a claim and return the token.
	claims := x.Claims{
		UserId: uid,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	token, err := jwtToken.SignedString([]byte(*auth.Secret))
	if err != nil {
		return "", err
	}
	return token, nil
}

type quizInfo struct {
	Quizzes []quiz `json:"quiz"`
}

func quizQns(quizId string) ([]Question, error) {
	q := `{
		quiz(_uid_: ` + quizId + `) {
			quiz.question {
				_uid_
				text
				positive
				negative
				question.option {
					_uid_
					name
				}
				question.tag {
					name
				}
				multiple
			}
		}
	}`

	var resp quizInfo
	if err := dgraph.QueryAndUnmarshal(q, &resp); err != nil {
		return []Question{}, err
	}
	if len(resp.Quizzes) != 1 {
		return []Question{}, fmt.Errorf("Expected length of quizzes: %v. Got %v",
			1, len(resp.Quizzes))
	}

	return resp.Quizzes[0].Questions, nil
}

func candQuery(cid string) string {
	return `{
        quiz.candidate(_uid_:` + cid + `) {
                name
                email
                token
                validity
                complete
                quiz_start
                candidate.quiz {
                        _uid_
                        duration
                        cut_off
                        threshold
                }
          }
    }`
}

func filter(qns []Question) map[difficulty][]Question {
	qnDiffMap := make(map[difficulty][]Question)
	for _, q := range qns {
		tags := q.Tags
		// TODO - Move the difficulty into a separate field within the
		// question separate from tags. So that we don't have to loop over
		// the tags.
	L:
		for _, t := range tags {
			switch n := t.Name; n {
			case "easy":
				qnDiffMap[EASY] = append(qnDiffMap[EASY], q)
				break L
			case "medium":
				qnDiffMap[MEDIUM] = append(qnDiffMap[MEDIUM], q)
				break L
			case "hard":
				qnDiffMap[HARD] = append(qnDiffMap[HARD], q)
				break L
			default:
				continue
			}

		}
	}
	return qnDiffMap
}

// Checks for candidate in cache, if we find it then we return. Else we load up
// information from the Database into the cache.
func checkAndUpdate(uid string) (int, error) {
	if _, err := readMap(uid); err == nil {
		// Got candidate information in Cache, return.
		return http.StatusOK, nil
	}

	// Candidate doesn't exist in the map. So we get candidate info from database
	// and insert it into map.
	q := candQuery(uid)
	var resp resp
	if err := dgraph.QueryAndUnmarshal(q, &resp); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Something went wrong.")
	}

	if len(resp.Cand) != 1 || len(resp.Cand[0].Quiz) != 1 {
		// No candidiate found with given uid
		return http.StatusUnauthorized, fmt.Errorf("Invalid token.")
	}

	cand := resp.Cand[0]
	quiz := cand.Quiz[0]
	if cand.Complete {
		return http.StatusUnauthorized, fmt.Errorf("You have already completed the quiz.")
	}
	if quiz.Id == "" {
		return http.StatusUnauthorized, fmt.Errorf("Invalid token.")

	}

	c := Candidate{
		name:      cand.Name,
		token:     cand.Token,
		email:     cand.Email,
		quizStart: cand.QuizStart,
		level:     EASY,
	}
	// TODO - Check how can we store this in appropriate format so that explicit parsing isn't
	// required.
	var err error
	if c.validity, err = time.Parse("2006-01-02 15:04:05 +0000 UTC", cand.Validity); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Something went wrong.")
	}
	if c.validity.Before(time.Now()) {
		return http.StatusUnauthorized,
			fmt.Errorf("Your token has already expired. Please mail us at contact@dgraph.io.")
	}

	// We check that quiz duration hasn't elapsed in case the candidate tries
	// to validate again say after a browser crash.
	c.quizDuration = time.Minute * time.Duration(quiz.Duration)

	if timeLeft(c.quizStart, c.quizDuration) < 0 {
		return http.StatusUnauthorized, fmt.Errorf("Your token is no longer valid.")
	}

	c.quizCutoff = quiz.CutOff
	c.quizThreshold = quiz.Threshold

	// Get quiz questions for the quiz id.
	questions, err := quizQns(quiz.Id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Something went wrong.")
	}
	shuffleQuestions(questions)
	c.numQuestions = len(questions)
	c.qns = filter(questions)
	updateMap(uid, c)
	return http.StatusOK, nil
}

func Validate(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}

	// Requests for validation of token are throttled.
	select {
	case <-throttle:
		break
	case <-time.After(rate):
		sr.Write(w, "", "Too many requests. Please try after again.",
			http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	// The id is uid + random string. 33 is the length of the random string.
	if len(id) < 33 {
		sr.Write(w, "", "Invalid token.", http.StatusUnauthorized)
		return
	}

	uid, token := id[:len(id)-33], id[len(id)-33:]
	if status, err := checkAndUpdate(uid); err != nil {
		sr.Write(w, "", err.Error(), status)
		return
	}

	c, err := readMap(uid)
	if err != nil {
		sr.Write(w, "", "Candidate not found.", http.StatusBadRequest)
		return
	}

	if c.token != token {
		sr.Write(w, err.Error(), "Invalid token.", http.StatusUnauthorized)
		return
	}

	// Check for duplicate session.
	if !c.lastExchange.IsZero() && time.Since(c.lastExchange) < 5*time.Second {
		// We update lastExchange time in pings. If we got a ping within
		// the last n second that means there is another active session.
		sr.Write(w, "", "You have another active session. Please try after some time.",
			http.StatusUnauthorized)
		return
	}

	t, err := genToken(uid)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	vr := validateRes{
		Token:    t,
		Duration: timeLeft(c.quizStart, c.quizDuration).String(),
		Started:  !c.quizStart.IsZero(),
		Name:     c.name,
	}
	server.MarshalAndWrite(w, &vr)
}
