angular.module('GruiApp').service('questionService', [
  "MainService",
  function questionService(MainService) {
    return {
      saveQuestion: function(data) {
        return MainService.post('/add-question', data);
      },

      editQuestion: function(data) {
        return MainService.put('/question/' + data.uid, data);
      },

      editScore: function(data) {
        return MainService.post('/question/editScore', data);
      },

      getAllQuestions: function(hideLoader) {
        return MainService.post('/get-all-questions', {}, hideLoader)
          .then(function(data) {
            if (!data || !data.data) {
              return [];
            }
            data = data.data
            var questions = data.questions || [];
            var answers = data.answers || [];

            questions.forEach(mainVm.fixQuestionUnescape);

            var questionUids = questions.reduce(function(acc, q) {
              acc[q.uid] = q;
              q.answerCount = 0;
              q.answerTotalScore = 0;
              q.skipCount = 0;
              return acc
            }, {})

            answers.forEach(function(answer) {
              var question = questionUids[answer.questionUid];
              if (!question) {
                console.error('Uknown question for answer ', answer);
                return
              }
              question.answerCount = answer.totalCount;
              question.answerTotalScore = answer.totalScore;
              question.skipCount = answer.skippedCount || 0;
              question.difficulty = answer.correctCount / answer.totalCount;
            })

            return questions;
          })
      },

      getQuestion: function(questionId) {
        return MainService.get('/question/' + questionId)
          .then(function(data) {
            return mainVm.fixQuestionUnescape(data.data.question[0]);
          })
      },
    }
  }
]);

angular.module('GruiApp').service('allQuestions', [
  'questionService',
  '$rootScope',
  function(questionService, $rootScope) {
    var allQuestions = [];

    function fetchQuestions() {
      questionService.getAllQuestions(true).then(
        function(questions) {
          setTimeout(function() {
            $rootScope.$apply(function() {
              allQuestions = questions;
            });
          }, 1);
        },
        function(err) {
          console.error(err);
        });
    }
    fetchQuestions();

    setInterval(fetchQuestions, 60000);

    return {
      get: function() {
        return allQuestions;
      },
      refresh: fetchQuestions,
    }
  }
]);
