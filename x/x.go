package x

import (
	"flag"
	"fmt"
	"math/rand"
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
