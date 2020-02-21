angular.module('GruiApp').controller('quizController', [
  "$state",
  "quizService",
  "allQuestions",
  "questionService",
  function quizController($state, quizService, allQuestions, questionService) {
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

    quizVm.allQuestions = function() {
      return allQuestions.get();
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

    quizVm.clickSetPoints = function(question, positive, negative) {
      questionService.editScore({
        uid: question.uid,
        negative,
        positive,
      }).then(function(data) {
          allQuestions.refresh();
          SNACKBAR({
            message: data.Message,
            messageType: "success",
          })
          questionService.updateAllScores();
        }, function(err) {
          SNACKBAR({
            message: "Something went wrong: " + err,
            messageType: "error",
          })
        })
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
  function editQuizController($stateParams, quizService) {
    editQuizVm = this;

    quizVm.loadEmptyQuiz();

    // If we are editing an existing quiz - load it.
    if ($stateParams.quizId) {
      // Read by edit-quiz.html to send user back to this quiz after editing a qn.
      editQuizVm.quizId = $stateParams.quizId;

      quizService.getQuiz($stateParams.quizId)
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
