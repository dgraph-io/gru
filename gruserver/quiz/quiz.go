package quiz

type Answer struct {
	Id  string `json:"id"`
	Str string `json:"str"`
}

type Question struct {
	Id         string    `json:"id"`
	Str        string    `json:"str"`
	Options    []*Answer `json:"options"`
	IsMultiple bool      `json:"isMultiple"`
	Positive   float32   `json:"positive"`
	Negative   float32   `json:"negative"`
	Score      float32   `json:"score"`
}

type Response struct {
	Qid   string   `json:"qid"`
	Aid   []string `json:"aid"`
	Sid   string   `json:"sid"`
	Token string   `json:"token"`
}

type AnswerStatus struct {
	Status int64 `json:"status"`
}

type ServerStatus struct {
	TimeLeft string `protobuf:"bytes,1,opt,name=timeLeft,proto3" json:"timeLeft,omitempty"`
	Status   string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
}

type ClientStatus struct {
	CurQuestion string `protobuf:"bytes,1,opt,name=curQuestion,proto3" json:"curQuestion,omitempty"`
	Token       string `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
}
