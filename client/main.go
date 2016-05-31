package main

import (
	"bytes"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/gizak/termui"
)

var token = flag.String("token", "", "Authentication token")

const (
	// Test duration in minutes
	testDur = 60
)

var th int
var tw int

type question struct {
	str      string
	multiple bool
}

type answer struct {
	id  string
	str string
}

// for the questions screen.
var instructions *termui.Par
var timeLeft *termui.Par
var que *termui.Par
var score *termui.Par
var s *termui.Par
var a *termui.Par

// For the home screen
var demo *termui.Par
var terminal *termui.Par
var general *termui.Par
var scoring *termui.Par
var contact *termui.Par

var q1 = question{`What is the capital of France?`, false}

var q2 = question{`Which among the following were originally developed at Google?`, true}

var q3 = question{`Which one is the largest ocean in the world?`, false}

var a1 = []answer{
	{"1", "Berlin"},
	{"2", "Paris"},
	{"3", "Rome"},
	{"4", "London"},
}

var a2 = []answer{
	{"1", "Go programming language"},
	{"2", "Ruby"},
	{"3", "Angular"},
	{"4", "Rust"},
}

var a3 = []answer{
	{"1", "Indian"},
	{"2", "Pacific"},
	{"3", "Atlantic"},
	{"4", "Arctic"},
}

var currentQn = 1

type State int

const (
	options State = iota
	confirmAnswer
	confirmSkip
)

var state State

func renderHomeScreen() {
	termui.Render(instructions)
	// Adding an offset so that all these boxes come inside the instructions box.
	termui.Body.Y = 2
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(10, 1, terminal)),
		termui.NewRow(
			termui.NewCol(10, 1, general)),
		termui.NewRow(
			termui.NewCol(10, 1, scoring)),
		termui.NewRow(
			termui.NewCol(10, 1, contact)),
		termui.NewRow(
			termui.NewCol(10, 1, demo)))

	termui.Body.Align()
	termui.Render(termui.Body)
}

func renderSelectedAnswers(selected []string, m map[string]answer) {
	check := "Selected:\n\n"
	for _, k := range selected {
		check += m[string(k)].str + "\n"
	}
	check += "\nPress ENTER to confirm. Press any other key to cancel."
	a.Text = check
	state = confirmAnswer
	termui.Render(termui.Body)
}

func populate(q question, answers []answer) {
	que.Text = q.str

	// Selected contains the options user has already selected.
	selected := []string{}
	aBody := ""
	// State stores the state of the user, whether he is shown the options
	// for a question, or is being asked to confirm his answer/skip.
	state = options
	m := make(map[string]answer)
	start := 'a'
	var buf bytes.Buffer

	if q.multiple {
		buf.WriteString("This question could have multiple correct answers.\n\n")
	} else {
		buf.WriteString("This question only has a single correct answer.\n\n")
	}

	for _, ans := range answers {
		buf.WriteRune(start)
		buf.WriteRune(')')
		buf.WriteRune(' ')
		buf.WriteString(ans.str)
		buf.WriteRune('\n')

		// This is the handler for pressing keys which represent an answer.
		termui.Handle(fmt.Sprintf("/sys/kbd/%c", start),
			func(e termui.Event) {
				k := e.Data.(termui.EvtKbd).KeyStr

				// For single correct answer qn we just render
				// the selected answer.
				if !q.multiple {
					if state != options {
						return
					}
					selected = append(selected, k)
					renderSelectedAnswers(selected, m)
					return
				} else {
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
					if !exists && state == options {
						selected = append(selected, k)
						sort.StringSlice(selected).Sort()
						a.Text = aBody + strings.Join(selected, ", ")
						a.Text += "\nPress Enter to see chosen options."
						// fmt.Println(a.Text)
						termui.Render(termui.Body)
					}
				}
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
		state = confirmSkip
	})

	termui.Handle("/sys/kbd/<enter>", func(e termui.Event) {
		// If the user presses enter after selecting options
		// for a multiple choice question.
		// fmt.Println(state)
		if len(selected) > 0 && state == options {
			renderSelectedAnswers(selected, m)
		} else if state == confirmAnswer || state == confirmSkip {
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
		state = options
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

	que = termui.NewPar("")
	que.BorderLabel = "Question"
	que.PaddingLeft = 1
	que.PaddingRight = 1
	que.PaddingBottom = 1
	que.Height = 33

	s = termui.NewPar("Right answer => +5.0\n\nWrong answer => -2.5")
	s.BorderLabel = "Scoring"
	s.PaddingTop = 1
	s.PaddingLeft = 1
	s.Height = 33

	a = termui.NewPar("")
	a.TextFgColor = termui.ColorCyan
	a.BorderLabel = "Answers"
	a.PaddingLeft = 1
	a.PaddingRight = 1
	a.PaddingBottom = 1
	a.Height = 14

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(6, 0, timeLeft),
			termui.NewCol(6, 0, score)),
		termui.NewRow(
			termui.NewCol(10, 0, que),
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

	instructions = termui.NewPar("")
	instructions.BorderLabel = "Instructions"
	instructions.Height = 50
	instructions.Width = tw
	instructions.PaddingTop = 2

	// TODO - Take duration from constant.
	terminal = termui.NewPar(`
		- Please ensure that you can see all the 4 borders of the Instructions box.
		- If you can't see them, you need to increase the size of your terminal or adjust the font-size to a smaller value.
		- DO NOT proceed with the test, until you are able to see all 4 outer borders of the Instructions box.`)
	terminal.BorderLabel = "Terminal"
	terminal.Height = 8
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
		- At point you can press Ctrl-q to end the test.`)
	general.BorderLabel = "General"
	general.Height = 15
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
		- If there are any problems with the setup, or something is unclear, please DO NOT start the test.
		- Send email to contact@dgraph.io and tell us the problem. So we can solve it before you take the test.`)
	contact.BorderLabel = "Contact"
	contact.Height = 10
	contact.Width = tw
	contact.PaddingLeft = 2

	demo = termui.NewPar("We have a demo of the how the test would look like. Press s to start the demo.")
	demo.Border = false
	demo.Height = 3
	demo.Width = tw
	demo.TextFgColor = termui.ColorCyan
	demo.PaddingLeft = 2
	demo.PaddingTop = 1

	renderHomeScreen()

	draw := func(t int) {
	}
	// Pressing q terminates the ui.
	termui.Handle("/sys/kbd/C-q", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Handle("/sys/kbd/s", func(e termui.Event) {
		termui.Clear()
		termui.Body.Y = 0
		termui.Body.Rows = termui.Body.Rows[:0]
		demoInit()
	})
	termui.Handle("/timer/1s", func(e termui.Event) {
		t := e.Data.(termui.EvtTimer)
		draw(int(t.Count))
	})
	termui.Loop()
}
