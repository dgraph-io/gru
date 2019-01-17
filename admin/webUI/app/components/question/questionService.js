angular.module('GruiApp').service('questionService', [
  "$q",
  "$http",
  "$rootScope",
  "MainService",
  function questionService($q, $http, $rootScope, MainService) {
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
            return data && data.data && data.data.questions ? data.data.questions : [];
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
