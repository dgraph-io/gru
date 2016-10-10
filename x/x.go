package x

import (
	"flag"
	"fmt"
	"math/rand"

	jwt "github.com/dgrijalva/jwt-go"
)

var debug = flag.Bool("debug", false, "Whether to print debug info")

func Debug(log interface{}) {
	if *debug {
		fmt.Println(log)
	}
}

func Shuffle(ids []string) {
	for i := range ids {
		j := rand.Intn(i + 1)
		ids[i], ids[j] = ids[j], ids[i]
	}
}

type Claims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}
