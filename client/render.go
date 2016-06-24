package main

import (
	"fmt"
	"time"

	"github.com/gizak/termui"
)

func setupInfoPage(th, tw int) {
	instructions = termui.NewPar("")
	instructions.BorderLabel = "Instructions"
	instructions.Height = 50
	instructions.Width = tw
	instructions.PaddingTop = 2

	infoPage.terminal = termui.NewPar(`
                - Please ensure that you can see all the 4 borders of the Instructions box.
                - If you can't see them, you need to increase the size of your terminal or adjust the font-size to a smaller value.
                - DO NOT proceed with the test, until you are able to see all 4 outer borders of the Instructions box.`)
	infoPage.terminal.BorderLabel = "Terminal"
	infoPage.terminal.Height = 8
	infoPage.terminal.Width = tw
	infoPage.terminal.PaddingLeft = 2

	// TODO - Take duration from constant.
	infoPage.general = termui.NewPar(`
                - By taking this test, you agree not to discuss/post the questions shown here.
                - The duration of the test is 60 mins. Timing would be clearly shown.
                - Once you start the test, the timer would not stop, irrespective of any client side issues.
                - Questions can have single or multiple correct answers. They will be shown accordingly.
                - Your total score and the time left at any point in the test would be displayed on the top.
                - You would be given the option to have a second attempt at a question if your first answer is wrong.
                - The scoring for each attempt of a question, would be visible to you in a separate section.
                - At point you can press Ctrl-q to end the test.`)
	infoPage.general.BorderLabel = "General"
	infoPage.general.Height = 15
	infoPage.general.Width = tw
	infoPage.general.PaddingLeft = 2

	infoPage.scoring = termui.NewPar(`
                - There is NEGATIVE scoring for wrong answers. So, please DO NOT GUESS.
                - If you skip a question, the score awarded is always ZERO.
                - You might be given the option to recover from negative score with a second attempt.
                - In the above case, please note that another wrong answer would have further negative score.
                - scoring would be clearly marked in the question on the right hand side box.`)
	infoPage.scoring.BorderLabel = "Scoring"
	infoPage.scoring.Height = 10
	infoPage.scoring.Width = tw
	infoPage.scoring.PaddingLeft = 2

	infoPage.contact = termui.NewPar(`
                - If there are any problems with the setup, or something is unclear, please DO NOT start the test.
                - Send email to contact@dgraph.io and tell us the problem. So we can solve it before you take the test.`)
	infoPage.contact.BorderLabel = "Contact"
	infoPage.contact.Height = 10
	infoPage.contact.Width = tw
	infoPage.contact.PaddingLeft = 2

	infoPage.demo = termui.NewPar("We have a demo of the how the test would look like. Press s to start the demo.")
	infoPage.demo.Border = false
	infoPage.demo.Height = 3
	infoPage.demo.Width = tw
	infoPage.demo.TextFgColor = termui.ColorCyan
	infoPage.demo.PaddingLeft = 2
	infoPage.demo.PaddingTop = 1
}

var qp QuestionsPage

func setupQuestionsPage() {
	qp.timeLeft = termui.NewPar(fmt.Sprintf("%02d:00", testDur))
	qp.timeLeft.Height = 3
	qp.timeLeft.BorderLabel = "Time Left"

	qp.timeSpent = termui.NewPar("00:00")
	qp.timeSpent.Height = 3
	qp.timeSpent.BorderLabel = "Time spent"

	ts := 00.0
	qp.score = termui.NewPar(fmt.Sprintf("%2.1f", ts))
	qp.score.BorderLabel = "Total Score"
	qp.score.Height = 3

	qp.lastScore = termui.NewPar("0.0")
	qp.lastScore.BorderLabel = "Last Score"
	qp.lastScore.Height = 3

	qp.que = termui.NewPar("")
	qp.que.BorderLabel = "Question"
	qp.que.PaddingLeft = 1
	qp.que.PaddingRight = 1
	qp.que.PaddingBottom = 1
	qp.que.Height = 33

	qp.scoringInfo = termui.NewPar("")
	qp.scoringInfo.BorderLabel = "Scoring"
	qp.scoringInfo.PaddingTop = 1
	qp.scoringInfo.PaddingLeft = 1
	qp.scoringInfo.Height = 33

	qp.answers = termui.NewPar("")
	qp.answers.TextFgColor = termui.ColorCyan
	qp.answers.BorderLabel = "Answers"
	qp.answers.PaddingLeft = 1
	qp.answers.PaddingRight = 1
	qp.answers.PaddingBottom = 1
	qp.answers.Height = 14
}

func renderInstructionsPage(demoTaken bool) {
	resetHandlers()
	termui.Render(instructions)
	// Adding an offset so that all these boxes come inside the instructions box.
	termui.Body.Y = 2
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(10, 1, infoPage.terminal)),
		termui.NewRow(
			termui.NewCol(10, 1, infoPage.general)),
		termui.NewRow(
			termui.NewCol(10, 1, infoPage.scoring)),
		termui.NewRow(
			termui.NewCol(10, 1, infoPage.contact)),
		termui.NewRow(
			termui.NewCol(10, 1, infoPage.demo)))

	if demoTaken {
		infoPage.demo.Text = "Press s to start the test."
	}
	termui.Body.Align()
	termui.Render(termui.Body)

	termui.Handle("/sys/kbd/s", func(e termui.Event) {
		if !demoTaken {
			initializeDemo()
			return
		}
		clear()
		initializeTest()
	})
}

func renderQuestionsPage() {
	termui.Body.Y = 0
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(3, 0, qp.timeLeft),
			termui.NewCol(3, 0, qp.timeSpent),
			termui.NewCol(3, 0, qp.score),
			termui.NewCol(3, 0, qp.lastScore)),
		termui.NewRow(
			termui.NewCol(10, 0, qp.que),
			termui.NewCol(2, 0, qp.scoringInfo)),
		termui.NewRow(
			termui.NewCol(12, 0, qp.answers)))

	termui.Body.Align()
	termui.Render(termui.Body)

	secondsCount := 0
	leftTime.setTimeLeft(testDur * time.Minute)

	termui.Handle("/timer/1s", func(e termui.Event) {
		secondsCount += 1
		timeTaken += 1
		leftTime.setTimeLeft(leftTime.left - time.Second)
		qp.timeSpent.Text = fmt.Sprintf("%02d:%02d", timeTaken/60,
			timeTaken%60)
		qp.timeLeft.Text = fmt.Sprintf("%02d:%02d", leftTime.left/time.Minute,
			(leftTime.left%time.Minute)/time.Second)
		termui.Render(termui.Body)
	})
}
