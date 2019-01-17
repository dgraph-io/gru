angular.module('GruiApp').controller('quizController', [
  "$state",
  "quizService",
  "allQuestions",
  function quizController($state, quizService, allQuestions) {
    mainVm.pageName = "quiz";
    quizVm = this;
    quizVm.quiz = {};

    quizVm.submitQuiz = function() {
      var validataionError = quizVm.validateQuiz(quizVm.quiz);
      if (validataionError) {
        SNACKBAR({
          message: validataionError,
          messageType: "error",
        })
        return
      }

      quizService.saveQuiz(quizVm.quiz)
        .then(function(data) {
          quizVm.quiz = {}
          SNACKBAR({
            message: data.Message,
            messageType: "success",
          })
          $state.transitionTo("quiz.all");
        }, function(err) {
          console.error(err);
        })
    }

    function validateQuiz(quiz) {
      if (!quiz.name) {
        return "Please enter valid Quiz name"
      }
      if (!quiz.duration) {
        return "Please enter valid time"
      }
      if (!quiz.questions) {
        return "Please add question to the quiz before submitting"
      }
      if (quiz.threshold >= 0) {
        return "Threshold should be less than 0"
      }
      if (quiz.cut_off >= quizVm.getTotalScore(quiz.questions)) {
        return "Cutoff should be less than the total possible score"
      }
      return false
    }

    quizVm.getTotalScore = function() {
      // TODO
      return 666;
      var totalScore = 0;
      questions.forEach(function(question) {
        if (!question.is_delete) {
          totalScore += question.correct.length * question.positive;
        }
      });
      return totalScore;
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

angular.module('GruiApp').controller('addQuizController', [
  'allQuestions',
  function addQuizController(allQuestions) {
    addQuizVm = this;
    quizVm.quiz = {};
  }
]);

angular.module('GruiApp').controller('editQuizController', [
  "$stateParams",
  "quizService",
  "allQuestions",
  function editQuizController($stateParams, quizService, allQuestions) {
    editQuizVm = this;
    quizVm.quiz = {};

    // Read by edit-quiz.html to send user back to this quiz after editing a qn.
    editQuizVm.quizId = $stateParams.quizID;

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
        console.log('q uids = ', quiz.questionUids);
      }, function(err) {
        console.error(err);
      });

    editQuizVm.getQuestionCount = function getQuestionCount() {
      // TODO
      return 333;
    }
  }
]);
