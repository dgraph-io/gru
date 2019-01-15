angular.module('GruiApp').controller('quizController', [
  "$scope",
  "$rootScope",
  "$stateParams",
  "$http",
  "$state",
  "quizService",
  "questionService",
  function quizController($scope, $rootScope, $stateParams, $http, $state, quizService, questionService) {
    mainVm.pageName = "quiz";
    quizVm = this;
    quizVm.newQuiz = {};

    quizVm.removeSelectedQuestion = removeSelectedQuestion;
    quizVm.removeCheckedQuestion = removeCheckedQuestion;
    quizVm.addQuizForm = addQuizForm;
    quizVm.validateInput = validateInput;
    quizVm.getTotalScore = getTotalScore;
    quizVm.resetForm = resetForm;

    quizVm.getAllQuestions = function() {
      questionService.getAllQuestions().then(function(questions) {
        mainVm.allQuestions = questions;
      });

      $rootScope.upgradeMDL();
    }

    function removeSelectedQuestion(key) {
      delete quizVm.newQuiz.questions[key];
      if (!Object.keys(quizVm.newQuiz.questions).length) {
        delete quizVm.newQuiz.questions;
      }
    }

    function removeCheckedQuestion(index) {
      var questionIndex = 'question-' + index;
      if (quizVm.newQuiz.questions[questionIndex] === false) {
        quizVm.removeSelectedQuestion(questionIndex);
      }
    }

    function addQuizForm() {
      var questions = []
      var requestData = {};
      requestData = angular.copy(quizVm.newQuiz);

      validataionError = quizVm.validateInput(requestData);
      if (validataionError) {
        SNACKBAR({
          message: validataionError,
          messageType: "error",
        })
        return
      }
      var qustionsClone = angular.copy(quizVm.newQuiz.questions)
      angular.forEach(qustionsClone, function(value, key) {
        if (qustionsClone[key]) {
          questions.push({
            uid: value.uid
          });
        }
      });

      requestData.questions = questions;

      quizService.saveQuiz(requestData)
        .then(function(data) {
          quizVm.newQuiz = {}
          SNACKBAR({
            message: data.Message,
            messageType: "success",
          })
          $state.transitionTo("quiz.all");
        }, function(err) {
          console.error(err);
        })
    }

    function validateInput(inputs) {
      if (!inputs.name) {
        return "Please enter valid Quiz name"
      }
      if (!inputs.duration) {
        return "Please enter valid time"
      }
      if (!inputs.questions) {
        return "Please add question to the quiz before submitting"
      }
      if (inputs.threshold >= 0) {
        return "Threshold should be less than 0"
      }
      if (inputs.cut_off >= getTotalScore(inputs.questions)) {
        return "Cutoff should be less than the total possible score"
      }
      return false
    }

    function getTotalScore(questions) {
      var totalScore = 0;
      angular.forEach(questions, function(question, key) {
        if (!question.is_delete) {
          totalScore += question.correct.length * question.positive;
        }
      });
      if (quizVm.newQuiz.newQuestions && quizVm.newQuiz.newQuestions.length) {
        for (var i = 0; i < quizVm.newQuiz.newQuestions.length; i++) {
          var commulativeScore = 0;
          var thisQues = quizVm.newQuiz.newQuestions[i];
          commulativeScore = thisQues.correct.length * thisQues.positive;
          totalScore += commulativeScore;
        }
      }
      return totalScore;
    }

    function resetForm() {
      quizVm.newQuiz = {};
    }
  }
]);

angular.module('GruiApp').controller('allQuizController', [
  "quizService",
  "questionService",
  function allQuizController(quizService, questionService) {
    quizVm.allQuizes = [];

    quizService.getAllQuizzes().then(function(quizzes) {
      quizVm.allQuizes = quizzes;
    }, function(err) {
      console.error(err);
    });
    quizVm.getAllQuestions();
  }
]);

angular.module('GruiApp').controller('addQuizController', [
  function addQuizController() {
    quizVm.getAllQuestions();
  }
]);

angular.module('GruiApp').controller('editQuizController', [
  "$rootScope",
  "$stateParams",
  "$state",
  "quizService",
  function editQuizController($rootScope, $stateParams, $state, quizService) {
    editQuizVm = this;
    quizVm.newQuiz = {};
    editQuizVm.selectedQuestion;

    mainVm.allQuestions = [];

    editQuizVm.editQuiz = editQuiz;
    editQuizVm.isExisting = isExisting;

    editQuizVm.quizId = $stateParams.quizID;

    quizService.getQuiz($stateParams.quizID)
      .then(function(quiz) {
        quizVm.newQuiz = quiz;
        quizVm.newQuiz.duration = parseInt(quizVm.newQuiz.duration)
        quizVm.newQuiz.cut_off = parseFloat(quizVm.newQuiz.cut_off)
        quizVm.newQuiz.threshold = parseFloat(quizVm.newQuiz.threshold)

        editQuizVm.selectedQuestion = quiz['quiz.question'];
        quizVm.newQuiz.newQuestions = [];

        quizVm.getAllQuestions();
      }, function(err) {
        console.error(err);
      });

    function editQuiz() {
      quizVm.newQuiz.questions = angular.copy(quizVm.newQuiz['quiz.question']);
      validataionError = quizVm.validateInput(quizVm.newQuiz);
      if (validataionError) {
        SNACKBAR({
          message: validataionError,
          messageType: "error",
        })
        return
      }

      var newQues = quizVm.newQuiz.newQuestions;

      var deletedAllExisting = true;
      for (var j = 0; j < quizVm.newQuiz.questions.length; j++) {
        if (!quizVm.newQuiz.questions[j].is_delete) {
          deletedAllExisting = false;
        }
      }

      if (deletedAllExisting && !newQues.length) {
        SNACKBAR({
          message: "You must add at least one question",
          messageType: "error",
        });
        return;
      }

      if (newQues) {
        for (var i = 0; i < newQues.length; i++) {
          quizVm.newQuiz.questions.push({
            uid: newQues[i].uid,
            text: newQues[i].text,
          });
        }
      }

      // API CALL
      quizService.editQuiz(quizVm.newQuiz)
        .then(function(data) {
          SNACKBAR({
            message: data.Message,
            messageType: "error",
          });
          quizVm.newQuiz = {};
          $state.transitionTo("quiz.all");
        }, function(err) {
          console.error(err);
        })
    }

    editQuizVm.addNewQuestion = function addNewQuestion(question, index) {
      var questionLength = editQuizVm.selectedQuestion.length;

      if (question.is_checked) {
        for (var i = 0; i < questionLength; i++) {
          var currentQues = editQuizVm.selectedQuestion[i];
          if (currentQues.uid == question.uid) {
            if (currentQues.is_delete === true) {
              currentQues.is_delete = false;
              return;
            } else {
              SNACKBAR({
                message: "Already selected, uncheck to remove it",
                messageType: "error",
              });
              return;
            }
          }
        }

        question.index = index;
        quizVm.newQuiz.newQuestions.push(question);
      } else {
        for (var i = 0; i < questionLength; i++) {
          var currentQues = editQuizVm.selectedQuestion[i];
          if (currentQues.uid == question.uid) {
            if (!currentQues.is_delete) {
              currentQues.is_delete = true;
              return;
            }
          }
        }
        var idx = mainVm.indexOfObject(quizVm.newQuiz.newQuestions, question);
        if (idx >= 0) {
          quizVm.newQuiz.newQuestions.splice(idx, 1)
        }
      }
    }

    editQuizVm.getQuestionCount = function getQuestionCount() {
      var existingCount = 0;
      var existingQues = quizVm.newQuiz['quiz.question'];
      for (var i = 0; i < existingQues.length; i++) {
        if (!existingQues[i].is_delete) {
          existingCount += 1;
        }
      }
      return existingCount + quizVm.newQuiz.newQuestions.length;
    }

    function isExisting(question) {
      var existingQues = editQuizVm.selectedQuestion;
      var existingQuesLen = existingQues.length;
      for (var i = 0; i < existingQuesLen; i++) {
        if (!existingQues[i].is_delete && existingQues[i].uid == question.uid) {
          question.is_checked = true;
          return;
        }
      }
      return false;
    }
  }
]);
