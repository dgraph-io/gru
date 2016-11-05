(function() {

  function quizService($q, $http, $rootScope, MainService) {

    var services = {}; //Object to return

    services.getAllQuizes = function() {
      return MainService.get('/get-all-quizes');
    }

    services.saveQuiz = function(data) {
      return MainService.post('/add-quiz', data);
    }

    services.editQuiz = function(data) {
      return MainService.put('/quiz/' + data._uid_, data);
    }

    services.getQuiz = function(data) {
      return MainService.get('/quiz/' + data);
    }

    return services;
  }

  var quizServiceArray = [
    "$q",
    "$http",
    "$rootScope",
    "MainService",
    quizService,
  ];

  angular.module('GruiApp').service('quizService', quizServiceArray);

})();
