(function() {

  function homeController($scope, $rootScope, $http, $q, $state, Upload) {

    // VARIABLE DECLARATION
    homeVm = this;
    mainVm.pageName = "home"

    // FUNCTION DECLARATION
    // homeVm.userAuthentication = userAuthentication;

    // FUNCTION DEFINITION

    // Check if user is authorized
    // upload later on form submit or something similar
    $scope.submit = function() {
      if (homeVm.file) {
        $scope.upload(homeVm.file);
      }
    };

    // upload on file select or drop
    $scope.upload = function(file) {
      Upload.upload({
        url: 'upload/url',
        data: { file: file, 'username': $scope.username }
      }).then(function(resp) {
        console.log('Success ' + resp.config.data.file.name + 'uploaded. Response: ' + resp.data);
      }, function(resp) {
        console.log('Error status: ' + resp.status);
      }, function(evt) {
        var progressPercentage = parseInt(100.0 * evt.loaded / evt.total);
        console.log('progress: ' + progressPercentage + '% ' + evt.config.data.file.name);
      });
    };
  }
  var homeDependency = [
    "$scope",
    "$rootScope",
    "$http",
    "$q",
    "$state",
    'Upload',
    homeController
  ];
  angular.module('GruiApp').controller('homeController', homeDependency);

})();
