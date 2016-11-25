(function() {

  function homeController($scope, $rootScope, $http, $q, $state) {

    // VARIABLE DECLARATION
    homeVm = this;
    mainVm.pageName = "home"

    // FUNCTION DECLARATION
    homeVm.userAuthentication = userAuthentication;

    // FUNCTION DEFINITION

    // Check if user is authorized
    function userAuthentication(testId) {}
  }
  var homeDependency = [
    "$scope",
    "$rootScope",
    "$http",
    "$q",
    "$state",
    homeController
  ];
  angular.module('GruiApp').controller('homeController', homeDependency);

})();
