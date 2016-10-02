(function(){

    function questionService($q, $http, $rootScope, MainService) {

    	var services = {}; //Object to return
      var base_url = "http://localhost:8082";

      services.saveQuestion = function(data){
        return MainService.post('/add-question', data)
      }

      services.editQuestion = function(data){
        var deferred = $q.defer();

        var req = {
          method: 'POST',
          url: base_url + '/edit-question',
          data: data,
          dataType: 'json',
        }
        mainVm.showAjaxLoader = true;
        $http(req)
        .then(function(data) { 
            mainVm.showAjaxLoader = false;
            deferred.resolve(data.data);
          },
          function(response, code) {
            mainVm.showAjaxLoader = false;
            deferred.reject(response);
          }
        );

        return deferred.promise;
      }

      services.getAllQuestions = function(requestData){
        return MainService.post('/get-all-questions', {"id": ""})
      }

      // private functions
      function handleSuccess(data) { //SUCCESS API HIT
          deferred.resolve(data);
      }
      function handleError(error) { //ERROR ON API HIT
          deferred.reject(error);
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