package main

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/gizak/termui"
)

var token = flag.String("token", "", "Authentication token")

const (
	// Test duration in minutes
	testDur = 60
)

var th int
var tw int

type answer struct {
	id  string
	str string
}

// for the questions screen.
var timeLeft *termui.Par
var question *termui.Par
var score *termui.Par
var s *termui.Par
var a *termui.Par

// For the home screen
var terminal *termui.Par
var general *termui.Par
var scoring *termui.Par
var contact *termui.Par

var q1 = `
	What is the capital of France?
	`

var q2 = `Which is the largest country in the world in terms of land area?`

var q3 = `Which one is the largest ocean in the world?`

var a1 = []answer{
	{"1", "Berlin"},
	{"2", "Paris"},
	{"3", "Rome"},
	{"4", "London"},
}

var a2 = []answer{
	{"1", "Russia"},
	{"2", "Canada"},
	{"3", "USA"},
	{"4", "China"},
}

var a3 = []answer{
	{"1", "Indian"},
	{"2", "Pacific"},
	{"3", "Atlantic"},
	{"4", "Arctic"},
}

var currentQn = 1

func renderHomeScreen() {
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(12, 0, terminal)),
		termui.NewRow(
			termui.NewCol(12, 0, general)),
		termui.NewRow(
			termui.NewCol(12, 0, scoring)),
		termui.NewRow(
			termui.NewCol(12, 0, contact)))

	termui.Body.Align()
	termui.Render(termui.Body)
}

func populate(q string, answers []answer) {
	question.Text = q

	selected := []string{}
	state := "options"
	skip := false
	aBody := ""

	var buf bytes.Buffer
	m := make(map[string]answer)
	start := 'a'
	for _, ans := range answers {
		buf.WriteRune(start)
		buf.WriteRune(')')
		buf.WriteRune(' ')
		buf.WriteString(ans.str)
		buf.WriteRune('\n')

		termui.Handle(fmt.Sprintf("/sys/kbd/%c", start),
			func(e termui.Event) {
				k := e.Data.(termui.EvtKbd).KeyStr
				// check := "Selected:\n\n" + m[string(k)].str + "\n\nPress ENTER to confirm. Press any other key to cancel."
				// a.Text = check
				exists := false
				for _, key := range selected {
					if key == k {
						exists = true
					}
				}
				if !exists && state == "options" {
					selected = append(selected, k)
					a.Text += k + " "
					termui.Render(termui.Body)
				}

				// fmt.Println(a.Text)
			})

		m[string(start)] = ans
		start++
	}
	buf.WriteString("\ns) Skip question\n\n")

	aBody = buf.String()
	a.Text = aBody

	termui.Handle("/sys/kbd/s", func(e termui.Event) {
		a.Text = "Are you sure you want to skip the question? \n\nPress ENTER to confirm. Press any other key to cancel."
		termui.Render(termui.Body)
		skip = true
	})

	termui.Handle("/sys/kbd/<enter>", func(e termui.Event) {
		if len(selected) > 0 && state == "options" {
			check := "Selected:\n\n"
			for _, k := range selected {
				check += m[string(k)].str + "\n"
			}
			check += "\nPress ENTER to confirm. Press any other key to cancel."
			a.Text = check
			state = "confirm"
			termui.Render(termui.Body)
		} else if state == "confirm" || skip {
			if currentQn == 0 {
				populate(q1, a1)
			}
			if currentQn == 1 {
				populate(q2, a2)
			} else if currentQn == 2 {
				populate(q3, a3)
			} else if currentQn == 3 {
				termui.Clear()
				termui.Body.Rows = termui.Body.Rows[:0]
				renderHomeScreen()
				currentQn = -1
			}
			currentQn += 1
		}
	})

	termui.Handle("/sys/kbd", func(e termui.Event) {
		a.Text = aBody
		selected = selected[:0]
		state = "options"
		skip = false
		termui.Render(termui.Body)
	})

	termui.Render(termui.Body)
}

func demoInit() {
	timeLeft = termui.NewPar(fmt.Sprintf("%d mins", testDur))
	timeLeft.Height = 3
	timeLeft.BorderLabel = "Time Left"

	ts := 0.0
	score = termui.NewPar(fmt.Sprintf("%.1f", ts))
	score.BorderLabel = "Total Score"
	score.Height = 3

	question = termui.NewPar("")
	question.BorderLabel = "Question"
	question.PaddingLeft = 1
	question.PaddingRight = 1
	question.PaddingBottom = 1
	question.Height = 35

	s = termui.NewPar("Right answer => +5.0\n\nWrong answer => -2.5")
	s.BorderLabel = "Scoring"
	s.BorderBottom = false
	s.PaddingTop = 1
	s.PaddingLeft = 1
	s.Height = 35

	a = termui.NewPar("")
	a.TextFgColor = termui.ColorCyan
	a.BorderLabel = "Answers"
	a.PaddingLeft = 1
	a.PaddingRight = 1
	a.PaddingBottom = 1
	a.Height = 11

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(6, 0, timeLeft),
			termui.NewCol(6, 0, score)),
		termui.NewRow(
			termui.NewCol(10, 0, question),
			termui.NewCol(2, 0, s)),
		termui.NewRow(
			termui.NewCol(12, 0, a)))

	termui.Body.Align()
	termui.Render(termui.Body)
	populate(q1, a1)
}

func main() {
	flag.Parse()
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	th = termui.TermHeight()
	tw = termui.TermWidth()

	// TODO - Take duration from constant.
	terminal = termui.NewPar(`
		- Please ensure that you can see all the 4 borders of the Instructions box.
		- If you can't see them, you need to increase the size of your terminal or adjust the font-size to a smaller value.
		- DO NOT proceed with the test, until you are able to see all 4 outer borders of the boxes titled Terminal, General, Scoring and Contact.`)
	terminal.BorderLabel = "Terminal"
	terminal.Height = 10
	terminal.Width = tw
	terminal.PaddingLeft = 2

	// TODO - Take duration from constant.
	general = termui.NewPar(`
		- By taking this test, you agree not to discuss/post the questions shown here.
		- The duration of the test is 60 mins. Timing would be clearly shown.
		- Once you start the test, the timer would not stop, irrespective of any client side issues.
		- Questions can have single or multiple correct answers. They will be shown accordingly.
		- Your total score and the time left at any point in the test would be displayed on the top.
		- You would be given the option to have a second attempt at a question if your first answer is wrong.
		- The scoring for each attempt of a question, would be visible to you in a separate section.


		We have a demo of the how the test would look like. Press s to start the demo.
		At point you can press Ctrl-q to end the test.`)
	general.BorderLabel = "General"
	general.Height = 20
	general.Width = tw
	general.PaddingLeft = 2

	scoring = termui.NewPar(`
		- There is NEGATIVE scoring for wrong answers. So, please DO NOT GUESS.
		- If you skip a question, the score awarded is always ZERO.
		- You might be given the option to recover from negative score with a second attempt.
		- In the above case, please note that another wrong answer would have further negative score.
		- Scoring would be clearly marked in the question on the right hand side box.`)
	scoring.BorderLabel = "Scoring"
	scoring.Height = 10
	scoring.Width = tw
	scoring.PaddingLeft = 2

	contact = termui.NewPar(`
		- If there are any problems with the setup, or something is unclear, please do not start the test.
		- Send email to contact@dgraph.io and tell us the problem. So we can solve it before you take the test.`)
	contact.BorderLabel = "Contact"
	contact.Height = 9
	contact.Width = tw
	contact.PaddingLeft = 2

	renderHomeScreen()

	draw := func(t int) {
	}
	// Pressing q terminates the ui.
	termui.Handle("/sys/kbd/C-q", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Handle("/sys/kbd/s", func(e termui.Event) {
		termui.Body.Rows = termui.Body.Rows[:0]
		demoInit()
	})
	termui.Handle("/timer/1s", func(e termui.Event) {
		t := e.Data.(termui.EvtTimer)
		draw(int(t.Count))
	})
	termui.Loop()
}
