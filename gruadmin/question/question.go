package question

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/dgraph-io/gru/gruadmin/tag"
	"github.com/dgraph-io/gru/x"
)

type Question struct {
	Uid      string `json:"_uid_"`
	Text     string
	Positive float64
	Negative float64
	Tags     []tag.Tag
	Options  []Option
}

type Option struct {
	Uid        string `json:"_uid_"`
	Name       string
	Is_correct bool
}

// API for "Adding Question" to Database
func Add(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	var ques Question

	if r.Method == "OPTIONS" {
		return
	}

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}

	x.Debug(ques)

	// // Creating query mutation to save question informations.
	question_info_mutation := "mutation { set { <rootQuestion> <question> <_new_:question> . \n	<_new_:question> <text> \"" + ques.Text +
		"\" . \n <_new_:question> <positive> \"" + strconv.FormatFloat(ques.Positive, 'g', -1, 64) +
		"\" . \n <_new_:question> <negative> \"" + strconv.FormatFloat(ques.Negative, 'g', -1, 64) + "\" . \n "

	// Create and associate Options
	for l := range ques.Options {
		index := strconv.Itoa(l)
		question_info_mutation += "<_new_:option" + index + "> <name> \"" + ques.Options[l].Name +
			"\" .\n <_new_:question> <question.option> <_new_:option" + index + "> . \n "

		// If this option is correct answer
		if ques.Options[l].Is_correct == true {
			question_info_mutation += "<_new_:question> <question.correct> <_new_:option" + index + "> . \n "
		}
	}
	x.Debug(ques.Tags)
	// Create and associate Tags
	for i := range ques.Tags {
		if ques.Tags[i].Uid != "" {
			x.Debug(ques.Tags[i].Uid)
			question_info_mutation += "<_new_:question> <question.tag> <_uid_:" + ques.Tags[i].Uid +
				"> . \n <_uid_:" + ques.Tags[i].Uid + "> <tag.question> <_new_:question> . \n "
		} else {
			index := strconv.Itoa(i)
			question_info_mutation += "<_new_:tag" + index + "> <name> \"" + ques.Tags[i].Name +
				"\" .\n <_new_:question> <question.tag> <_new_:tag" + index +
				"> . \n <_new_:tag" + index + "> <tag.question> <_new_:question> . \n "
		}
	}

	question_info_mutation += " }}"
	x.Debug(question_info_mutation)
	res := dgraph.SendMutation(question_info_mutation)
	if res.Success {
		res.Message = "Question Successfully Saved!"
	}
	question_json_response, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(question_json_response)
}

// FETCH All Questions HANDLER: Incomplete
func Index(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var ques GetQuestion
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}
	x.Debug(ques)
	var question_mutation string
	if ques.Id != "" {
		question_mutation = "{debug(_xid_: rootQuestion) { question (after: " + ques.Id + ", first: 10) { _uid_ text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	} else {
		question_mutation = "{debug(_xid_: rootQuestion) { question (first: 20) { _uid_ text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	}
	x.Debug(question_mutation)
	w.Header().Set("Content-Type", "application/json")

	question_response, err := http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(question_mutation))
	if err != nil {
		panic(err)
	}
	defer question_response.Body.Close()
	question_body, err := ioutil.ReadAll(question_response.Body)
	if err != nil {
		panic(err)
	}
	x.Debug(string(question_body))

	jsonResp, err := json.Marshal(string(question_body))
	if err != nil {
		panic(err)
	}

	w.Write(jsonResp)
}

type QuestionAPIResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	UIDS    struct {
		Question uint64 `json:question`
	}
}

// method to parse question response
func parseQuestionResponse(question_body []byte) (*QuestionAPIResponse, error) {
	var question_response = new(QuestionAPIResponse)
	err := json.Unmarshal(question_body, &question_response)
	if err != nil {
		log.Fatal(err)
	}
	return question_response, err
}

func Edit(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	var ques Question

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}
	x.Debug(ques)
	// Creating query mutation to save question informations.
	question_info_mutation := "mutation { set { <_uid_:" + ques.Uid + "> <text> \"" + ques.Text +
		"\" . \n <_uid_:" + ques.Uid + "> <positive> \"" + strconv.FormatFloat(ques.Positive, 'g', -1, 64) +
		"\" . \n  <_uid_:" + ques.Uid + "> <negative> \"" + strconv.FormatFloat(ques.Negative, 'g', -1, 64) + "\" . \n "

	// Create and associate Options
	for l := range ques.Options {
		// index := strconv.Itoa(l)
		question_info_mutation += "<_uid_:" + ques.Options[l].Uid + "> <name> \"" + ques.Options[l].Name +
			"\" .\n <_uid_:" + ques.Uid + "> <question.option> <_uid_:" + ques.Options[l].Uid + "> . \n "

		// If this option is correct answer
		if ques.Options[l].Is_correct == true {
			x.Debug(ques.Options[l])
			question_info_mutation += "<_uid_:" + ques.Uid + "> <question.correct> <_uid_:" + ques.Options[l].Uid + "> . \n "
		}
		if ques.Options[l].Is_correct == false {
			delete_correct := "mutation { delete { <_uid_:" + ques.Uid + "> <question.correct> <_uid_:" + ques.Options[l].Uid + "> .}}"
			_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(delete_correct))
			if err != nil {
				panic(err)
			}
		}
	}

	// Create and associate Tags
	for i := range ques.Tags {
		if ques.Tags[i].Uid != "" && ques.Tags[i].Is_delete == true {
			query_mutation := "mutation { delete { <_uid_:" + ques.Uid + "> <question.tag> <_uid_:" + ques.Tags[i].Uid + "> .}}"
			x.Debug(query_mutation)
			_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(query_mutation))
			if err != nil {
				panic(err)
			}

		} else if ques.Tags[i].Uid != "" {
			question_info_mutation += "<_uid_:" + ques.Uid + "> <question.tag> <_uid_:" + ques.Tags[i].Uid +
				"> . \n <_uid_:" + ques.Tags[i].Uid + "> <tag.question> <_uid_:" + ques.Uid + "> . \n "
		} else {
			index := strconv.Itoa(i)
			question_info_mutation += "<_new_:tag" + index + "> <name> \"" + ques.Tags[i].Name +
				"\" .\n <_uid_:" + ques.Uid + "> <question.tag> <_new_:tag" + index +
				"> . \n <_new_:tag" + index + "> <tag.question> <_uid_:" + ques.Uid + "> . \n "
		}
	}

	question_info_mutation += " }}"
	x.Debug(question_info_mutation)

	_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(question_info_mutation))
	if err != nil {
		panic(err)
	}

	stats := &server.Response{true, "Question Successfully Updated!"}
	question_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(question_json_response)
}

type TagFilter struct {
	UID string
}

type GetQuestion struct {
	Id string
}

// FILTER QUESTION HANDLER: Fileter By Tags
func Filter(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	var tag TagFilter

	err := json.NewDecoder(r.Body).Decode(&tag)
	if err != nil {
		panic(err)
	}

	filter_query := "{root(_uid_: " + tag.UID + ") { tag.question { text }}"
	filter_response, err := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(filter_query))
	if err != nil {
		panic(err)
	}
	defer filter_response.Body.Close()
	filter_body, err := ioutil.ReadAll(filter_response.Body)
	if err != nil {
		panic(err)
	}
	x.Debug(string(filter_body))
	jsonResp, err := json.Marshal(string(filter_body))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
