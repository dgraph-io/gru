(function() {

  function candidateService($q, $http, MainService) {

    var services = {}; //Object to return

    services.getQuestion = function() {
      return MainService.post('/quiz/question');
    }

    services.sendFeedback = function(data) {
      var deferred = $q.defer();

      mainVm.showAjaxLoader = true;
      $http({
          method: 'POST',
          url: mainVm.candidate_url + '/quiz/feedback',
          data: $.param(data),
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        })
        .then(function(data) {
            mainVm.showAjaxLoader = false;
            deferred.resolve(data);
          },
          function(response, code) {
            mainVm.showAjaxLoader = false;
            deferred.reject(response);
          }
        );
      return deferred.promise;
    }

    services.submitAnswer = function(requestData) {
      var deferred = $q.defer();

      mainVm.showAjaxLoader = true;
      $http({
          method: 'POST',
          url: mainVm.candidate_url + '/quiz/answer',
          data: $.param(requestData),
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        })
        .then(function(data) {
            mainVm.showAjaxLoader = false;
            deferred.resolve(data);
          },
          function(response, code) {
            mainVm.showAjaxLoader = false;
            deferred.reject(response);
          }
        );
      return deferred.promise;
    }

    services.getTime = function() {
      return MainService.post('/quiz/ping', "", true);
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

  var candidateServiceArray = [
    "$q",
    "$http",
    "MainService",
    candidateService
  ];

  angular.module('GruiApp').service('candidateService', candidateServiceArray);

})();
