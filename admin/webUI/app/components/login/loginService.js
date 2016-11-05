(function() {

  function loginService($q, $http, $rootScope, MainService) {

  }
  var loginServiceArray = [
    "$q",
    "$http",
    "$rootScope",
    "MainService",
    loginService
  ];

  angular.module('GruiApp').service('loginService', loginServiceArray);

})();
