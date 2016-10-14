package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Success bool
	// Message to display to the user.
	Message string
	// Actual error.
	Error string
}

func AddCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers",
		"Authorization,Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token,"+
			"X-Auth-Token, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Connection", "close")

	w.Header().Set("Content-Type", "application/json")
}

func ReadBody(r *http.Request, s interface{}) {
	err := json.NewDecoder(r.Body).Decode(s)
	if err != nil {
		log.Fatal(err)
	}
}

func MarshalResponse(r Response) []byte {
	fallbackMsg := "Something went wrong"

	if r.Message == "" {
		r.Message = fallbackMsg
	}
	b, err := json.Marshal(r)
	if err != nil {
		b, _ = json.Marshal(Response{
			Success: false,
			Message: fallbackMsg,
			Error:   err.Error(),
		})
	}
	return b
}

func (r Response) Write(w http.ResponseWriter, err string, msg string, status int) {
	r.Error = err
	r.Message = msg
	w.WriteHeader(status)
	w.Write(MarshalResponse(r))
}

func WriteBody(w http.ResponseWriter, res Response) {
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Fatal(err)
	}
	// r, err := json.Marshal(res)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// w.Write(r)
}
