angular.module('GruiApp').controller('quizController', [
  "$state",
  "quizService",
  "allQuestions",
  function quizController($state, quizService, allQuestions) {
    mainVm.pageName = "quiz";
    quizVm = this;

    quizVm.selectedTags = {}

    quizVm.loadEmptyQuiz = function() {
      quizVm.quiz = {
        questionUids: {},
      };
    }
    quizVm.loadEmptyQuiz();

    quizVm.submitQuiz = function() {
      var quiz = quizVm.quiz;

      var validataionError = quizVm.validateQuiz();
      if (validataionError) {
        SNACKBAR({
          message: validataionError,
          messageType: "error",
        })
        return
      }

      quiz.questions = allQuestions.get().map(function (q) {
        return {
          uid: q.uid,
          is_delete: !quiz.questionUids[q.uid] || undefined,
        }
      });

      var apiCall = quiz.uid
          ? quizService.editQuiz(quiz)
          : quizService.saveQuiz(quiz)

      return apiCall.then(function(data) {
        SNACKBAR({
          message: data.Message,
          messageType: "success",
        })
        $state.transitionTo("quiz.all");
      }, function(err) {
        SNACKBAR({
          message: "Something went wrong: " + err,
          messageType: "error",
        })
      })
    }

    quizVm.validateQuiz = function() {
      var quiz = quizVm.quiz;

      if (!quiz.name) {
        return "Please enter valid Quiz name"
      }
      if (!quiz.duration) {
        return "Please enter valid time"
      }
      if (!quizVm.quizQuestions().length) {
        return "Please add question to the quiz before submitting"
      }
      if (quiz.threshold >= 0) {
        return "Threshold should be less than 0"
      }
      if (quiz.cut_off >= quizVm.getTotalScore(quizVm.quizQuestions())) {
        return "Cutoff should be less than the total possible score"
      }
      return false
    }

    function findByUid(arr, uid) {
      var idx = arr.findIndex(function(el) { return el.uid == uid });
      return {
        index: idx,
        item: idx >= 0 ? arr[idx] : null,
      }
    }

    quizVm.allQuestionTags = function() {
      var allTags = quizVm.quizQuestions().reduce(function(acc, q) {
        return acc.concat(q.tags)
      }, [])
      allTags = allTags.filter(function(tag, index) {
        return index == findByUid(allTags, tag.uid).index;
      })
      allTags.sort(function(a, b) {
        return a.name < b.name ? -1 : (a.name > b.name ? 1 : 0);
      })
      return allTags;
    }

    quizVm.getTagStats = function(tag) {
      var allQuestions = quizVm.quizQuestions();
      var withTag = allQuestions.filter(function(q) {
        return findByUid(q.tags, tag.uid).item
      })
      var score = quizVm.getTotalScore(withTag)
      return {
        count: withTag.length,
        score: score,
        share: score / quizVm.getTotalScore(allQuestions)
      }
    }

    quizVm.showAddQuestionModal = function() {
      mainVm.openModal({
        class: "add-question-modal-template",
        hideClose: true,
        template: "add-question-modal-template",
      });
    }

    quizVm.isOptionCorrect = function(question, option) {
      return findByUid(question.correct, option.uid).item != null
    }

    quizVm.removeQuestion = function(question) {
      quizVm.quiz.questionUids[question.uid] = false;
    }

    quizVm.addQuestion = function(question) {
      quizVm.quiz.questionUids[question.uid] = true;
    }

    quizVm.isQuestionInQuiz = function(question) {
      return quizVm.quiz.questionUids[question.uid];
    }

    quizVm.isQuestionInFilter = function(question) {
      var missingTag = Object.keys(quizVm.selectedTags).find(function(tagUid) {
        return quizVm.selectedTags[tagUid] && findByUid(question.tags,tagUid).item == null
      });
      return !missingTag;
    }

    quizVm.selectedTagsNotEmpty = function() {
      return Object.keys(quizVm.selectedTags).filter(function(tagUid) {
        return quizVm.selectedTags[tagUid];
      }).length;
    }

    // TODO: There's probably a better way but it's not worth my time to google.
    // needed for inverse filter.
    quizVm.isNotInQuiz = function(question) {
      return !quizVm.isQuestionInQuiz(question);
    }

    quizVm.quizQuestions = function() {
      var questionUids = quizVm.quiz.questionUids;
      return allQuestions.get().filter(function (q) {
        return questionUids[q.uid];
      })
    }

    quizVm.allQuestions = function() {
      return allQuestions.get();
    }

    quizVm.getTotalScore = function(questions) {
      return questions.reduce(function(acc, question) {
        return acc + question.correct.length * question.positive;
      }, 0);
    }
  }
]);

angular.module('GruiApp').controller('allQuizController', [
  "quizService",
  function allQuizController(quizService) {
    quizVm.allQuizes = [];

    quizService.getAllQuizzes().then(function(quizzes) {
      quizVm.allQuizes = quizzes;
    }, function(err) {
      console.error(err);
    });
  }
]);

angular.module('GruiApp').controller('editQuizController', [
  "$stateParams",
  "quizService",
  "MainService",
  function editQuizController($stateParams, quizService, mainService) {
    editQuizVm = this;

    quizVm.loadEmptyQuiz();

    // TODO: this is copy-pasted from inviteController
    function parseFatReport(candidates) {
      candidates = candidates.filter(x => x.complete && !x.deleted)

      const qnMap = {}

      function getScore(answers, correct, positive, negative) {
        let res = 0
        for (let ans of answers) {
          res += correct.indexOf(ans) >= 0 ? positive : -negative
        }
        return res;
      }

      for (let k of candidates) {
        for (let qRec of k["candidate.question"]) {
          if (!qRec["candidate.answer"]) {
            continue
          }

          let q = qRec.question[0]
          const correct = q.correct.map(x => x.uid)
          const maxScore = correct.length * q.positive

          const answers = qRec["candidate.answer"].split(',')
          const skipped = !answers.length || answers[0] === "skip"

          const score = skipped ? 0 : getScore(answers, correct, q.positive, q.negative)

          const curQ = qnMap[q.uid] = qnMap[q.uid] || {
            uid: q.uid,
            name: q.name,
            answerCount: 0,
            skippedCount: 0,
            maxScore,
            numOfCorrectChoices: correct.length,
            sumScores: 0,
            sumScoresSquared: 0,
            valMap: getEmptyValMap(),
          }

          curQ.valMap[score]++;

          curQ.answerCount ++;
          curQ.skippedCount += skipped ? 1 : 0;
          curQ.sumScores += score;
          curQ.sumScoresSquared += score * score;
        }
      }

      for (let q of Object.values(qnMap)) {
        let mean = 0;
        let std = q.maxScore / 2;
        const N = q.answerCount;
        if (N > 2) {
          mean = q.sumScores / N
          std = Math.sqrt(q.sumScoresSquared / (N - 1) - q.sumScores * q.sumScores / N / (N - 1))
        }
        q.mean = mean
        q.std = std
        q.difficulty = q.mean / q.maxScore
      }

      return qnMap
    }

    const MAX_SCORE = 300

    function getEmptyValMap() {
      const res = {}
      for (let i = -MAX_SCORE; i <= MAX_SCORE; i++ ) {
        res[i] = 0
      }
      return res
    }

    function buildScoreMatrix(qnMap) {
      let endValMap = getEmptyValMap()
      endValMap[0] = 1

      for (let q of Object.values(qnMap)) {
        const N = q.answerCount;

        if (N <= 0) {
          continue;
        }
          // Calculate new score probability using dynamic programming
          const newEndValMap = getEmptyValMap()
          for (let i = -MAX_SCORE; i <=MAX_SCORE; i++) {
            for (let j = -MAX_SCORE; j <=MAX_SCORE; j++) {
              if (Math.abs(i + j) > MAX_SCORE) {
                continue;
              }
              newEndValMap[i + j] += endValMap[i] * q.valMap[j] / N
            }
          }
          endValMap = newEndValMap
      }
      console.log('End Val Map', endValMap)
      return endValMap
    }

    function getQuizStatsQuery(quizId) {
      return `{
        fatReport(func: uid(${quizId})) {
    			quiz.candidate {
    				uid
    				name
    				email
    				score
    				token
    				validity
    				complete
    				deleted
    				quiz_start
    				invite_sent
    				candidate.question {
    	        question {
    	          uid
    	          name
    	          positive
    	          negative
    	          correct: question.correct {
    	            uid
    	          }
    	        }
    	        question.asked
    	        question.answered
    	        candidate.answer
    	        }
    			}
    		}
      }`;
    }

    // If we are editing an existing quiz - load it.
    if ($stateParams.quizID) {
      // Read by edit-quiz.html to send user back to this quiz after editing a qn.
      editQuizVm.quizId = $stateParams.quizID;

      mainService.proxy(getQuizStatsQuery($stateParams.quizID))
        .then(function(report) {
          quizVm.questionStats = parseFatReport(report.data.fatReport[0]["quiz.candidate"])
          quizVm.percentileMap = buildScoreMatrix(quizVm.questionStats)
        })

      quizService.getQuiz($stateParams.quizID)
        .then(function(quiz) {
          quizVm.quiz = quiz;
          quiz.duration = parseInt(quiz.duration)
          quiz.cut_off = parseFloat(quiz.cut_off)
          quiz.threshold = parseFloat(quiz.threshold)

          quiz.questionUids = {}
          if (quiz['quiz.question']) {
            quiz['quiz.question'].forEach(function (q) {
              quiz.questionUids[q.uid] = true;
            })
          }
        }, function(err) {
          console.error(err);
        });
    }
  }
]);
