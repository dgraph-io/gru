(function(){

  function questionService($q, $http, $rootScope, MainService) {

  	var services = {}; //Object to return

    services.saveQuestion = function(data){
      return MainService.post('/add-question', data)
    }

    services.editQuestion = function(data){
      return MainService.post('/edit-question', data)
    }

    services.getAllQuestions = function(requestData){
      return MainService.post('/get-all-questions', {"id": ""})
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