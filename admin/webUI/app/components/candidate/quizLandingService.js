(function() {
  function quizLandingService($q, $http, MainService) {
    var services = {}; //Object to return

    return services;
  }

  var quizLandingServicesArray = [
    "$q",
    "$http",
    "MainService",
    quizLandingService
  ];

  angular
    .module("GruiApp")
    .service("quizLandingService", quizLandingServicesArray);
})();
