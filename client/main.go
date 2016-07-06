package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/dgraph-io/gru/server/interact"
	"github.com/gizak/termui"
)

var (
	token   = flag.String("token", "testtoken", "Authentication token")
	address = flag.String("address", "gru.dgraph.io:443", "Address of the server")
	tls     = flag.Bool("tls", true, "Connection uses TLS if true, else plain TCP")
)

type State int

const (
	// Denotes if user is seeing the options.
	options State = iota
	// User is being asked to confirm an answer.
	confirmAnswer
	// User is being asked to confirm if they want to skip answering the
	// question.
	confirmSkip
	END      = "END"
	DEMOEND  = "DEMOEND"
	PINGDUR  = 5 * time.Second
	NUMRETRY = 12
	TIMEOUT  = 5 * time.Second
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
	startedPing  bool
	testEndCh    chan struct{}
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
		"We will get in touch with you soon",
		"Press Ctrl + Q to exit."}, ". "), score)
}

func fetchAndDisplayQn() {
	client := interact.NewGruQuizClient(conn)

	ctx, _ := context.WithTimeout(context.Background(), TIMEOUT)
	q, err := client.GetQuestion(ctx,
		&interact.Req{Repeat: false, Sid: s.Id, Token: *token})

	try := 0
	for err != nil {
		statusNoConnection()
		log.Printf("Could not get question.Got err: %v", err)
		q, err = client.GetQuestion(ctx,
			&interact.Req{Repeat: false, Sid: s.Id, Token: *token})
		try++
		if try > NUMRETRY {
			showErrorPage()
		}
	}
	statusConnected()
	s.currentQn = q

	// TODO(pawan) - If he has already taken the demo,don't show the screen again.
	if q.Id == DEMOEND {
		clear()
		s.totalScore = 0.0
		s.lastScore = 0.0
		renderInstructionsPage(true)
		return
	}
	if q.Id == END {
		showFinalPage(finalScore(q.Totscore))
		return
	}
	populateQuestionsPage(q)
}

func sendStatus(pingFail *int) {
	status := interact.ClientStatus{
		s.currentQn.Id,
		*token,
	}
	client := interact.NewGruQuizClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), TIMEOUT)
	serverS, err := client.Ping(ctx, &status)
	// TODO - Retry here and don't show error page till X mins.
	if err != nil {
		(*pingFail)++
		log.Println("While sending ping", err)
		if (*pingFail) > NUMRETRY {
			showErrorPage()
			return
		}
		statusNoConnection()
		return
	}
	*pingFail = 0
	statusConnected()

	if serverS.Status == DEMOEND {
		// If its a dummy token, show final screen else instructions box.
		if strings.HasPrefix(*token, "test-") {
			clear()
			showFinalPage(finalScore(s.currentQn.Totscore))
			return
		}
		clear()
		renderInstructionsPage(true)
		return
	}

	if serverS.Status == END {
		clear()
		showFinalPage(finalScore(s.currentQn.Totscore))
		return
	}

	s.servTime.dur, err = time.ParseDuration(serverS.TimeLeft)
	if err != nil {
		log.Printf("Error parsing time from server, %v", err)
	}

	s.leftTime.setTimeLeft(s.servTime.dur)
	termui.Render(termui.Body)
}

func startPing() {
	if s.startedPing {
		return
	}
	pingFail := 0
	ticker := time.NewTicker(PINGDUR)
	go func() {
	L:
		for {
			// If the test ends because of time or questions being
			// over, stop the pings.
			select {
			case <-s.testEndCh:
				break L
			case <-ticker.C:
				sendStatus(&pingFail)
			}
		}
	}()
	s.startedPing = true
}

func initializeTest(tl string) {
	setupQuestionsPage()
	renderQuestionsPage(tl)
	fetchAndDisplayQn()
	startPing()

	if s.currentQn.Id == END {
		return
	}
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
		ctx, _ := context.WithTimeout(context.Background(), TIMEOUT)
		_, err := client.SendAnswer(ctx, &resp)
		try := 0
		for err != nil {
			log.Println("While sending Answer", err)
			statusNoConnection()
			_, err = client.SendAnswer(ctx, &resp)
			try++
			if try > NUMRETRY {
				showErrorPage()
				return
			}
		}
		statusConnected()
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
	if q.IsMultiple {
		qp.scoringInfo.Text = fmt.Sprintf(
			"For every right answer => +%1.1f\n\nFor every wrong answer => -%1.1f\n\nSkip question => %1.1f",
			q.Positive, q.Negative, 0.0)
	} else {
		qp.scoringInfo.Text = fmt.Sprintf(
			"Right answer => +%1.1f\n\nWrong answer => -%1.1f\n\nSkip question => %1.1f",
			q.Positive, q.Negative, 0.0)
	}

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
	startPing()
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	if *token == "testtoken" {
		*token = fmt.Sprintf("test-%s", RandStringBytes(10))
	}

	logFile, _ := os.OpenFile("gru.log", os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666)
	syscall.Dup2(int(logFile.Fd()), 2)
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	s.testEndCh = make(chan struct{}, 2)
	setupErrorPage()
	setupInstructionsPage()
	instructions.BorderLabel = "Connecting"
	instructions.Text = "Connecting to server..."
	termui.Render(instructions)

	var opts []grpc.DialOption
	if *tls {
		var creds credentials.TransportCredentials
		host := strings.Split(*address, ":")
		creds = credentials.NewClientTLSFromCert(nil, host[0])
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithBlock(), grpc.WithTimeout(TIMEOUT))
	conn, err = grpc.Dial(*address, opts...)
	if err != nil {
		log.Println(err)
		showErrorPage()
		return
	}

	client := interact.NewGruQuizClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), TIMEOUT)
	ses, err := client.Authenticate(ctx, &interact.Token{Id: *token})
	if err != nil {
		log.Println(err)
		errorPage.Text = grpc.ErrorDesc(err) + " Press Ctrl+Q to exit and try again."
		termui.Render(errorPage)
	} else {
		setupInitialPage(ses)
	}

	// Pressing Ctrl-q terminates the ui.
	termui.Handle("/sys/kbd/C-q", func(termui.Event) {
		conn.Close()
		termui.StopLoop()
	})

	termui.Loop()

}
