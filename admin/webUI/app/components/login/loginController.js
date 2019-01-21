angular.module('GruiApp').controller('loginController', [
  "$scope",
  "$state",
  "MainService",
  function loginController($scope, $state, MainService) {
    if (mainVm.isLoggedIn()) {
      $state.transitionTo("root");
    }

    loginVm = this;
    loginVm.authData = {};
    mainVm.pageName = "login-page"

    loginVm.authenticate = function() {
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
              messageType: "success",
            })
            localStorage.setItem("token", data.token);
            $state.transitionTo('root')
          }
        }, function(err) {
          SNACKBAR({
            message: err.data.Message,
            messageType: "error",
          })
        })
    }
  }
]);
