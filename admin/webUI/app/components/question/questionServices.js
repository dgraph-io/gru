(function() {

  function questionService($q, $http, $rootScope, MainService) {

    var services = {}; //Object to return

    services.saveQuestion = function(data) {
      return MainService.post('/add-question', data);
    }

    services.editQuestion = function(data) {
      return MainService.put('/question/' + data.uid, data);
    }

    services.getAllQuestions = function(hideLoader) {
      return MainService.post('/get-all-questions', hideLoader)
        .then(function(data) {
          return data && data.data && data.data.questions ? data.data.questions : [];
        })
    }

    services.getQuestion = function(questionId) {
      return MainService.get('/question/' + questionId)
        .then(function(data) {
          return data.data.question[0];
        })
    }

    return services;
  }

  var questionServiceArray = [
    "$q",
    "$http",
    "$rootScope",
    "MainService",
    questionService
  ];

  angular.module('GruiApp').service('questionService', questionServiceArray);
})();
