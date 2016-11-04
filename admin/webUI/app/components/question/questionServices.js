(function() {

  function questionService($q, $http, $rootScope, MainService) {

    var services = {}; //Object to return

    services.saveQuestion = function(data) {
      return MainService.post('/add-question', data);
    }

    services.editQuestion = function(data) {
      return MainService.put('/question/' + data._uid_, data);
    }

    services.getAllQuestions = function(requestData, hideLoader) {
      return MainService.post('/get-all-questions', requestData, hideLoader);
    }

    services.getQuestion = function(questionID) {
      return MainService.get('/question/' + questionID);
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
