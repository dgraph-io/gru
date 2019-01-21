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

      getAllQuestions: function(hideLoader) {
        return MainService.post('/get-all-questions', {}, hideLoader)
          .then(function(data) {
            if (!data || !data.data) {
              return [];
            }
            data = data.data
            var questions = data.questions || [];
            var answers = data.answers || [];

            var questionUids = questions.reduce(function(acc, q) {
              acc[q.uid] = q;
              q.answerCount = 0;
              q.answerTotalScore = 0;
              q.skipCount = 0;
              return acc
            }, {})

            answers.forEach(function(answer) {
              var question = questionUids[answer.questionUid];
              question.answerCount++;
              question.answerTotalScore += answer.score;
              if (answer.score == 0) {
                question.skipCount++;
              }
            })

            return questions;
          })
      },

      getQuestion: function(questionId) {
        return MainService.get('/question/' + questionId)
          .then(function(data) {
            return data.data.question[0];
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
