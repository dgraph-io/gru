package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Success bool
	Message string
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
