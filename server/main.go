package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var data string

type Option struct {
	Uid string
	Str string
}

type T struct {
	Qid      string
	Question string
	Correct  []string
	Opt      []map[string]string
	Score    int
	Tag      string
}

var filename = flag.String("file", "", "Input question file")

func main() {
	flag.Parse()

	buf := bytes.NewBuffer(nil)
	f, _ := os.Open(*filename)
	io.Copy(buf, f)
	data := buf.Bytes()

	var m map[interface{}][]T

	err := yaml.Unmarshal(data, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)

	d, err := yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))
}
