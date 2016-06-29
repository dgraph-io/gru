package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"sort"
	"strconv"
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
var endTT chan *interact.ServerStatus

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

type clock struct {
	sync.Mutex
	dur time.Duration
}

type Session struct {
	Id           string
	currentQn    *interact.Question
	status       State
	leftTime     clock
	servTime     clock
	testDuration string
	timeTaken    int
	totalScore   float32
	lastScore    float32
	showingAns   bool
}

var s Session

func (c *clock) setTimeLeft(serverDur time.Duration) {
	c.Lock()
	if c.dur-serverDur >= time.Second ||
		serverDur-c.dur >= time.Second {
		c.dur = serverDur
	}
	c.Unlock()
}

// Declaring connection as a global as can't reuse the client(its unexported)
var conn *grpc.ClientConn

func finalScore(score float32) string {
	return fmt.Sprintf(strings.Join([]string{"Thank you for taking the test",
		"Your final score was %4.1f",
		"We will get in touch with you soon."}, ". "), score)
}

func fetchAndDisplayQn() {
	client := interact.NewGruQuizClient(conn)

	q, err := client.GetQuestion(context.Background(),
		&interact.Req{Repeat: false, Sid: s.Id, Token: *token})

	if err != nil {
		log.Fatalf("Could not get question.Got err: %v", err)
	}
	s.currentQn = q

	// TODO(pawan) - If he has already taken the demo,don't show the screen again.
	if q.Id == "DEMOEND" {
		clear()
		s.totalScore = 0.0
		s.lastScore = 0.0
		renderInstructionsPage(true)
		return
	} else if q.Id == "END" {
		showFinalPage(finalScore(q.Totscore))
		return
	}
	populateQuestionsPage(q)
}

func streamRecv(stream interact.GruQuiz_StreamChanClient) {
	for {
		msg, err := stream.Recv()

		if err != nil {
			if err != io.EOF {
				//log.Printf("Error while receiving stream, %v", err)
				endTT <- msg
				return
			}
			endTT <- msg
			log.Println("got end message")
			return
		}

		if msg.Status == "END" {
			clear()
			showFinalPage(finalScore(s.currentQn.Totscore))
			return
		}

		s.servTime.dur, err = time.ParseDuration(msg.TimeLeft)
		if err != nil {
			log.Printf("Error parsing time from server, %v", err)
		}

		s.leftTime.setTimeLeft(s.servTime.dur)
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
					s.currentQn.Id,
					*token,
				}
				if err := stream.Send(cliStat); err != nil {
					// TODO: Show error s.status in a separate box
					//glog.WithField("err", err).Error("Error sending to stream")
				}
			}
		}
	}
}

func initializeTest(tl string) {
	setupQuestionsPage()
	renderQuestionsPage(tl)
	fetchAndDisplayQn()

	if s.currentQn.Id == "END" {
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
	s.showingAns = true
	check := "Selected:\n\n"
	for _, k := range selected {
		check += m[string(k)].Str + "\n"
	}
	check += "\nPress ENTER to confirm. Press ESC to cancel."
	qp.answers.Text = check
	s.status = confirmAnswer
	termui.Render(termui.Body)
}

func optionHandler(e termui.Event, q *interact.Question, selected []string,
	m map[string]*interact.Answer, ansBody string) []string {
	k := e.Data.(termui.EvtKbd).KeyStr

	// For single correct answer qn we just render
	// the selected answer.
	if !q.IsMultiple {
		if s.status != options {
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
	if !exists && s.status == options {
		selected = append(selected, k)
		sort.StringSlice(selected).Sort()
		qp.answers.Text = ansBody + strings.Join(selected, ", ")
		qp.answers.Text += "\nPress Enter to see chosen options. Press ESC to cancel."
		termui.Render(termui.Body)
	}
	return selected
}

func enterHandler(e termui.Event, q *interact.Question, selected []string,
	m map[string]*interact.Answer) {
	// If the user presses enter after selecting options for a
	// multiple choice question.
	if q.IsMultiple && len(selected) > 0 && s.status == options {
		renderSelectedAnswers(selected, m)
		return
	}
	if s.status == confirmAnswer || s.status == confirmSkip {
		var answerIds []string
		for _, s := range selected {
			if s == "skip" {
				answerIds = []string{"skip"}
				break
			}
			answerIds = append(answerIds, m[s].Id)
		}
		resp := interact.Response{Qid: q.Id, Aid: answerIds,
			Sid: s.Id, Token: *token}
		client := interact.NewGruQuizClient(conn)
		client.SendAnswer(context.Background(), &resp)
		fetchAndDisplayQn()
	}
}

func escapeHandler(ansBody string, selected []string) []string {
	s.showingAns = false
	qp.answers.Text = ansBody
	selected = selected[:0]
	s.status = options
	termui.Render(termui.Body)
	return selected
}

func populateQuestionsPage(q *interact.Question) {
	resetHandlers()
	s.timeTaken = 0
	qp.que.Text = q.Str
	qp.scoringInfo.Text = fmt.Sprintf(
		"Right answer => +%1.1f\n\nWrong answer => -%1.1f\n\nSkip question Aut=> %1.1f",
		q.Positive, q.Negative, 0.0)

	// Selected contains the options user has already selected.
	selected := []string{}
	s.showingAns = false
	// This is the body of the answer which has all the options.
	ansBody := ""
	// Map m contains a map of the key to select an answer and the answer
	// corresponding to it.
	m := make(map[string]*interact.Answer)
	var buf bytes.Buffer

	s.status = options
	if q.IsMultiple {
		buf.WriteString("This question could have MULTIPLE correct answers.\n\n")
	} else {
		buf.WriteString("This question has a SINGLE correct answer.\n\n")
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
	s.lastScore = q.Totscore - s.totalScore
	qp.lastScore.Text = fmt.Sprintf("%2.1f", s.lastScore)
	s.totalScore = q.Totscore
	// We store this so that this can be rendered later based on different
	// key press.
	ansBody = buf.String()
	qp.answers.Text = ansBody
	termui.Render(termui.Body)

	// Attaching event handlers on the options for a answer.
	for i := 'a'; i < opt; i++ {
		termui.Handle(fmt.Sprintf("/sys/kbd/%c", i),
			func(e termui.Event) {
				if s.showingAns {
					return
				}
				selected = optionHandler(e, q, selected, m,
					ansBody)
			})
	}

	termui.Handle("/sys/kbd/s", func(e termui.Event) {
		if s.showingAns {
			return
		}
		s.showingAns = true
		qp.answers.Text = "Are you sure you want to skip the question? \n\nPress ENTER to confirm. Press ESC to cancel."
		selected = selected[:0]
		selected = append(selected, "skip")
		termui.Render(termui.Body)
		s.status = confirmSkip
	})

	termui.Handle("/sys/kbd/<enter>", func(e termui.Event) {
		if len(selected) == 0 {
			return
		}
		enterHandler(e, q, selected, m)
	})

	// On any other keypress we reset the answer text and the selected answers.
	termui.Handle("/sys/kbd/<escape>", func(e termui.Event) {
		selected = escapeHandler(ansBody, selected)
	})
}

func initializeDemo(tl string) {
	clear()
	setupQuestionsPage()
	renderQuestionsPage(tl)
	fetchAndDisplayQn()
}

func setupInitialPage(ses *interact.Session) {
	state := ses.State
	s.testDuration = ses.TestDuration
	d, _ := time.ParseDuration(s.testDuration)
	dm := strconv.FormatFloat(d.Minutes(), 'f', 0, 64)
	// TODO(pawan) - Handle error and take to final page.
	s.Id = ses.Id
	if state == interact.Quiz_TEST_FINISHED {
		//show final page saying test already taken and return
		showFinalPage("You have already taken the test.")
	}
	if state == interact.Quiz_TEST_STARTED {
		initializeTest(ses.TimeLeft)
	}
	setupInfoPage(termui.TermHeight(), termui.TermWidth(), dm)
	if state == interact.Quiz_DEMO_NOT_TAKEN {
		//call instructions screen with demo taken to be false
		renderInstructionsPage(false)
	}
	if state == interact.Quiz_TEST_NOT_TAKEN {
		//call instructions screen with demo taken to be true
		renderInstructionsPage(true)
	}
	if state == interact.Quiz_DEMO_STARTED {
		initializeDemo(ses.TestDuration)
	}
}

func main() {
	rand.Seed(42)
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

	client := interact.NewGruQuizClient(conn)
	s, err := client.Authenticate(context.Background(), &interact.Token{Id: *token})
	if err != nil {
		instructions.Text = grpc.ErrorDesc(err) + " Press Ctrl+Q to exit and try again."
		instructions.TextFgColor = termui.ColorRed
		termui.Render(instructions)
	} else {
		setupInitialPage(s)
	}

	// Pressing Ctrl-q terminates the ui.
	termui.Handle("/sys/kbd/C-q", func(termui.Event) {
		conn.Close()
		termui.StopLoop()
	})
	termui.Loop()
}
