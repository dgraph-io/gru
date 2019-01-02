angular.module('GruiApp').service('quizService', [
  "$q",
  "$http",
  "$rootScope",
  "MainService",
  function quizService($q, $http, $rootScope, MainService) {
    return {
      getAllQuizzes: function() {
        return MainService.get('/get-all-quizzes').then(function(data) {
          return data.data.quizzes || [];
        })
      },

      saveQuiz: function(data) {
        return MainService.post('/add-quiz', data);
      },

      editQuiz: function(data) {
        return MainService.put('/quiz/' + data.uid, data);
      },

      getQuiz: function(data) {
        return MainService.get('/quiz/' + data).then(function(data) {
          return data.data.quiz[0];
        });
      },
    };
  },
]);
