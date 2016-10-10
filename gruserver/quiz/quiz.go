package quiz

import (
	"fmt"
	"sync"
)

type Answer struct {
	id   string
	text string
}

type Question struct {
	id         string
	text       string
	options    []*Answer
	isMultiple bool
	positive   float32
	negative   float32
}

type Candidate struct {
	score float32
	qns   []string
}

func init() {
	cmap = make(map[string]Candidate)
}

var (
	cmap map[string]Candidate
	mu   sync.RWMutex
)

func UpdateMap(uid string, ids []string) {
	mu.Lock()
	c, ok := cmap[uid]
	if !ok {
		return
	}
	c.qns = ids
	cmap[uid] = c
	mu.Unlock()
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

// type Response struct {
// 	Qid   string   `json:"qid"`
// 	Aid   []string `json:"aid"`
// 	Sid   string   `json:"sid"`
// 	Token string   `json:"token"`
// }
//
// type AnswerStatus struct {
// 	Status int64 `json:"status"`
// }
//
// type ServerStatus struct {
// 	TimeLeft string `protobuf:"bytes,1,opt,name=timeLeft,proto3" json:"timeLeft,omitempty"`
// 	Status   string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
// }
//
// type ClientStatus struct {
// 	CurQuestion string `protobuf:"bytes,1,opt,name=curQuestion,proto3" json:"curQuestion,omitempty"`
// 	Token       string `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
// }
