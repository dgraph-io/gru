package x

import (
	"flag"
	"fmt"
	"math"

	jwt "github.com/dgrijalva/jwt-go"
)

var debug = flag.Bool("debug", false, "Whether to print debug info")

func Debug(log interface{}) {
	if *debug {
		fmt.Println(log)
	}
}

func StringInSlice(a string, list []string) int {
	for idx, b := range list {
		if b == a {
			return idx
		}
	}
	return -1
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

type Claims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}
