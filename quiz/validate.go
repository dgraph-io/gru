package quiz

import (
	"net/http"
	"time"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/auth"
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
	if !c.lastExchange.IsZero() && time.Now().Sub(c.lastExchange) < 5*time.Second {
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
