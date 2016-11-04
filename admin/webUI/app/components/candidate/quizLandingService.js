(function() {

  function quizLandingService($q, $http, MainService) {

    var services = {}; //Object to return

    services.addName = function(data) {
      var deferred = $q.defer();

      var candidateToken = JSON.parse(localStorage.getItem('candidate_info'));
      $http.defaults.headers.common['Authorization'] = 'Bearer ' + candidateToken.token;
      mainVm.showAjaxLoader = true;
      $http({
          method: 'POST',
          url: mainVm.candidate_url + '/quiz/name',
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

    return services;
  }

  var quizLandingServicesArray = [
    "$q",
    "$http",
    "MainService",
    quizLandingService
  ];

  angular.module('GruiApp').service('quizLandingService', quizLandingServicesArray);
})();
