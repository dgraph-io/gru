(function(){

    function questionService($q, $http, $rootScope) {

    	var services = {}; //Object to return
      var base_url = "http://localhost:8082";

      services.saveQuestion = function(data){
        var deferred = $q.defer();

        var req = {
          method: 'POST',
          url: base_url + '/add-question',
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
        var deferred = $q.defer();

        var req = {
					method: 'POST',
					url: base_url + '/get-all-questions',
          data: {
            "id": ""
          },
				}
        mainVm.showAjaxLoader = true;
        $http(req)
        .then(function(data) {
            mainVm.showAjaxLoader = false; 
			      deferred.resolve(data.data);
			    },
		     	function(response, code) {
		        deferred.reject(response);
		     	}
			  );

        return deferred.promise;
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
        questionService
    ];

    angular.module('GruiApp').service('questionService', questionServiceArray); 

})();