(function(){

    function questionService($q, $http, $rootScope) {

    	var services = {}; //Object to return
      var base_url = "http://localhost:8000";

      services.getQuestion = function(requestData){
          var deferred = $q.defer();

          var req = {
						method: 'GET',
						url: base_url + '/nextquestion?sid=' + requestData.sid,
						headers: {
						  'Authorization': requestData.authorization,
						}
					}

          $http(req)
          .then(function(data) { 
				      deferred.resolve(data.data);
				    },
			     	function(response, code) {
			        deferred.reject(response);
			     	}
				  );

          return deferred.promise;
      }

      services.getStatus = function(requestData) {
      	var deferred = $q.defer();

          $.ajax({
        		method: "POST",
          	url:	base_url + '/status', 
          	data: requestData, 
          	contentType: 'application/x-www-form-urlencoded; charset=utf-8',
          	headers: {
          	'Authorization': mainVm.questionInfo.userId,
          	},
          	success: function(data){
          		deferred.resolve(data);
          	}
          });

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