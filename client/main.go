package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/dgraph-io/gru/server/interact"
	"github.com/gizak/termui"
)

var token = flag.String("token", "testtoken", "Authentication token")

//TODO(Pawan) - Change default address to our server.
var address = flag.String("address", "localhost:8888", "Address of the server")
var curQuestion *interact.Question
var endTT chan *interact.ServerStatus

const (
	//TODO(pawan) - Get from server.
	testDur = 60
)

// Elements for the questions page.
var instructions *termui.Par
var lck sync.Mutex

type QuestionsPage struct {
	timeLeft    *termui.Par
	timeSpent   *termui.Par
	que         *termui.Par
	score       *termui.Par
	lastScore   *termui.Par
	scoringInfo *termui.Par
	answers     *termui.Par
}

type InformationPage struct {
	demo     *termui.Par
	terminal *termui.Par
	general  *termui.Par
	scoring  *termui.Par
	contact  *termui.Par
}

type State int

const (
	// Denotes if user is seeing the options.
	options State = iota
	// User is being asked to confirm an answer.
	confirmAnswer
	// User is being asked to confirm if they want to skip answering the
	// question.
	confirmSkip
)

// To mantain the state of user while he is answering a question.
var status State

var startTime time.Time
var leftTime, servTime time.Duration

// timeTaken per question displayed on the top.
var timeTaken int
var ts float32

// Last score
var ls float32

// Declaring connection as a global as can't reuse the client(its unexported)
var conn *grpc.ClientConn

var sessionId string

var infoPage InformationPage

func resetHandlers() {
	termui.Handle("/sys/kbd", func(e termui.Event) {})
	termui.Handle("/sys/kbd/s", func(e termui.Event) {})
	termui.Handle("/sys/kbd/<enter>", func(e termui.Event) {})
}

func showFinalPage(msg string) {
	instructions = termui.NewPar(msg)
	instructions.BorderLabel = "Thank You"
	instructions.Height = 10
	instructions.Width = termui.TermWidth() / 2
	instructions.Y = termui.TermHeight() / 4
	instructions.X = termui.TermWidth() / 4
	instructions.PaddingTop = 1
	instructions.PaddingLeft = 1

	termui.Clear()
	termui.Body.Rows = termui.Body.Rows[:0]
	termui.Render(instructions)
	resetHandlers()
	conn.Close()
}

func clear() {
	termui.Clear()
	termui.Body.Rows = termui.Body.Rows[:0]
}

func finalScore() string {
	return fmt.Sprintf(strings.Join([]string{"Thank you for taking the test",
		"Your final score was %3.1f",
		"We will get in touch with you soon."}, ". "))
}

func fetchAndDisplayQn() {
	client := interact.NewGruQuizClient(conn)

	q, err := client.GetQuestion(context.Background(),
		&interact.Req{Repeat: false, Sid: sessionId, Token: *token})

	if err != nil {
		log.Fatalf("Could not get question.Got err: %v", err)
	}
	curQuestion = q

	// TODO(pawan) - If he has already taken the demo,don't show the screen again.
	if q.Id == "DEMOEND" {
		clear()
		ts = 0.0
		ls = 0.0
		renderInstructionsPage(true)
		return
	} else if q.Id == "END" {
		showFinalPage(finalScore())
		return
	}
	populateQuestionsPage(q)
}

func streamRecv(stream interact.GruQuiz_StreamChanClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				log.Printf("Error while receiving stream, %v", err)
				endTT <- msg
				return
			}
			endTT <- msg
			log.Println("got end message")
			return
		}

		if msg.Status == "END" {
			clear()
			showFinalPage("dummy")
			// showFinalPage(curQuestion)
			return
		}
		servTime, err = time.ParseDuration(msg.TimeLeft)
		if err != nil {
			log.Printf("Error parsing time from server, %v", err)
		}
		if testDur*time.Second-leftTime-servTime > time.Second ||
			testDur*time.Second-leftTime-servTime < time.Second {
			lck.Lock()
			leftTime = testDur*time.Minute - servTime
			lck.Unlock()
		}
		termui.Render(termui.Body)
	}
}

func streamSend(stream interact.GruQuiz_StreamChanClient) {
	tickChan := time.NewTicker(time.Second * 5).C

	for {
		select {
		case _ = <-endTT:
			return
		case <-tickChan:
			{
				cliStat := &interact.ClientStatus{
					curQuestion.Id,
					*token,
				}
				if err := stream.Send(cliStat); err != nil {
					// TODO: Show error status in a separate box
					//glog.WithField("err", err).Error("Error sending to stream")
				}
			}
		}
	}
}

func initializeTest() {
	setupQuestionsPage()
	renderQuestionsPage()
	fetchAndDisplayQn()

	if curQuestion.Id == "END" {
		return
	}

	client := interact.NewGruQuizClient(conn)
	stream, err := client.StreamChan(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
	}

	cliStat := &interact.ClientStatus{
		"First",
		*token,
	}
	if err := stream.Send(cliStat); err != nil {
		log.Panic(err)
	}
	go streamRecv(stream)
	go streamSend(stream)
}

func renderSelectedAnswers(selected []string, m map[string]*interact.Answer) {
	check := "Selected:\n\n"
	for _, k := range selected {
		check += m[string(k)].Str + "\n"
	}
	check += "\nPress ENTER to confirm. Press any other key to cancel."
	qp.answers.Text = check
	status = confirmAnswer
	termui.Render(termui.Body)
}

func optionHandler(e termui.Event, q *interact.Question, selected []string,
	m map[string]*interact.Answer, ansBody string) []string {
	k := e.Data.(termui.EvtKbd).KeyStr

	// For single correct answer qn we just render
	// the selected answer.
	if !q.IsMultiple {
		if status != options {
			return []string{}
		}
		// We append the selected answer and render it.
		selected = append(selected, k)
		renderSelectedAnswers(selected, m)
		return selected
	}
	// For multiple choice questions we check
	// if the user has already selected the answer before.
	exists := false
	for _, key := range selected {
		if key == k {
			exists = true
		}
	}
	// If he hasn't selected the answer before, we display
	// it below the options now.
	if !exists && status == options {
		selected = append(selected, k)
		sort.StringSlice(selected).Sort()
		qp.answers.Text = ansBody + strings.Join(selected, ", ")
		qp.answers.Text += "\nPress Enter to see chosen options."
		termui.Render(termui.Body)
	}
	return selected
}

func enterHandler(e termui.Event, q *interact.Question, selected []string,
	m map[string]*interact.Answer) {
	// If the user presses enter after selecting options for a
	// multiple choice question.
	if q.IsMultiple && len(selected) > 0 && status == options {
		renderSelectedAnswers(selected, m)
	} else if status == confirmAnswer || status == confirmSkip {
		var answerIds []string
		for _, s := range selected {
			if s == "skip" {
				answerIds = []string{"skip"}
				break
			}
			answerIds = append(answerIds, m[s].Id)
		}
		resp := interact.Response{Qid: q.Id, Aid: answerIds,
			Sid: sessionId, Token: *token}
		client := interact.NewGruQuizClient(conn)
		client.SendAnswer(context.Background(), &resp)
		fetchAndDisplayQn()
	}
}

func keyHandler(ansBody string, selected []string) []string {
	qp.answers.Text = ansBody
	selected = selected[:0]
	status = options
	termui.Render(termui.Body)
	return selected
}

func populateQuestionsPage(q *interact.Question) {
	timeTaken = 0
	qp.que.Text = q.Str
	qp.scoringInfo.Text = fmt.Sprintf("Right answer => +%1.1f\n\nWrong answer => -%1.1f",
		q.Positive, q.Negative)

	// Selected contains the options user has already selected.
	selected := []string{}
	// This is the body of the answer which has all the options.
	ansBody := ""
	// Map m contains a map of the key to select an answer and the answer
	// corresponding to it.
	m := make(map[string]*interact.Answer)
	var buf bytes.Buffer

	status = options
	if q.IsMultiple {
		buf.WriteString("This question could have multiple correct answers.\n\n")
	} else {
		buf.WriteString("This question only has a single correct answer.\n\n")
	}
	opt := 'a'
	for _, o := range q.Options {
		buf.WriteRune(opt)
		buf.WriteRune(')')
		buf.WriteRune(' ')
		buf.WriteString(o.Str)
		buf.WriteRune('\n')
		m[string(opt)] = o
		opt++
	}
	buf.WriteString("\ns) Skip question\n\n")
	qp.score.Text = fmt.Sprintf("%3.1f", q.Totscore)
	ls = q.Totscore - ts
	qp.lastScore.Text = fmt.Sprintf("%2.1f", ls)
	ts = q.Totscore
	// We store this so that this can be rendered later based on different
	// key press.
	ansBody = buf.String()
	qp.answers.Text = ansBody
	termui.Render(termui.Body)

	// Attaching event handlers on the options for a answer.
	for i := 'a'; i < opt; i++ {
		termui.Handle(fmt.Sprintf("/sys/kbd/%c", i),
			func(e termui.Event) {
				selected = optionHandler(e, q, selected, m,
					ansBody)
			})
	}

	termui.Handle("/sys/kbd/s", func(e termui.Event) {
		qp.answers.Text = "Are you sure you want to skip the question? \n\nPress ENTER to confirm. Press any other key to cancel."
		selected = selected[:0]
		selected = append(selected, "skip")
		termui.Render(termui.Body)
		status = confirmSkip
	})

	termui.Handle("/sys/kbd/<enter>", func(e termui.Event) {
		if len(selected) == 0 {
			return
		}
		enterHandler(e, q, selected, m)
	})

	// On any other keypress we reset the answer text and the selected answers.
	termui.Handle("/sys/kbd", func(e termui.Event) {
		selected = keyHandler(ansBody, selected)
	})
}

func initializeDemo() {
	clear()
	setupQuestionsPage()
	renderQuestionsPage()
	fetchAndDisplayQn()
}

func setupInitialPage(s *interact.Session) {
	state := s.State
	if state == interact.Quiz_TEST_FINISHED {
		//show final page saying test already taken and return
		showFinalPage("You have already taken the test.")
	}
	sessionId = s.Id
	if state == interact.Quiz_TEST_STARTED {
		//call get questions and return
	}
	setupInfoPage(termui.TermHeight(), termui.TermWidth())
	if state == interact.Quiz_DEMO_NOT_TAKEN {
		//call instructions screen with demo taken to be false
		renderInstructionsPage(false)
	}
	if state == interact.Quiz_TEST_NOT_TAKEN {
		//call instructions screen with demo taken to be true
		renderInstructionsPage(true)
	}

}

func main() {
	flag.Parse()
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	conn, err = grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
		return
	}

	instructions = termui.NewPar(
		fmt.Sprintf("Wait a second, we are getting you started..."))
	instructions.BorderLabel = "Authenticating"
	instructions.Height = 10
	instructions.Width = termui.TermWidth() / 2
	instructions.Y = termui.TermHeight() / 4
	instructions.X = termui.TermWidth() / 4
	instructions.PaddingTop = 1
	instructions.PaddingLeft = 1
	termui.Render(instructions)

	// Set up a connection to the server.

	client := interact.NewGruQuizClient(conn)
	s, err := client.Authenticate(context.Background(), &interact.Token{Id: *token})
	if err != nil {
		instructions.Text = grpc.ErrorDesc(err) + " Press Ctrl+Q to exit and try again."
		instructions.TextFgColor = termui.ColorRed
		termui.Render(instructions)
		return
	}
	setupInitialPage(s)

	// Pressing Ctrl-q terminates the ui.
	termui.Handle("/sys/kbd/C-q", func(termui.Event) {
		conn.Close()
		termui.StopLoop()
	})
	termui.Loop()
}
