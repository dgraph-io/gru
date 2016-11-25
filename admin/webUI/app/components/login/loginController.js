(function() {

  function loginController($scope, $rootScope, $stateParams, $state, MainService) {
    if (mainVm.isLoggedIn()) {
      $state.transitionTo("root");
    }

    loginVm = this;
    loginVm.authData = {};
    mainVm.pageName = "login-page"

    // Function Declaration
    loginVm.authenticate = authenticate;

    // Functin Deinitions
    function authenticate() {
      if (!loginVm.authData.user || !loginVm.authData.password) {
        SNACKBAR({
          message: "Please fill all the details",
          messageType: "error",
        })
        return;
      }
      MainService.post('/login', loginVm.authData)
        .then(function(data) {
          if (data.token) {
            SNACKBAR({
              message: "Logged In Successfuly",
              messageType: "error",
            })
            localStorage.setItem("token", data.token);
            $state.transitionTo('root')
          }
        }, function(err) {
          SNACKBAR({
            message: "Something went wrong",
            messageType: "error",
          })
        })
    }
  }

  var loginDependency = [
    "$scope",
    "$rootScope",
    "$stateParams",
    "$state",
    "MainService",
    loginController
  ];
  angular.module('GruiApp').controller('loginController', loginDependency);

})();
