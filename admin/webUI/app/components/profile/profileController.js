(function() {

  function profileController($scope, $rootScope,profileService) {

    // VARIABLE DECLARATION
    profileVm = this;
    mainVm.pageName = "profile-page"
  }

  function editProfileController($scope, $rootScope, $state, profileService) {
    editProfileVm = this;
    editProfileVm.update = updateProfile;

    profileService.getProfile()
    .then(function(data) {
      editProfileVm.info = {}
      editProfileVm.info.name = data['info'][0]["company.name"]
      editProfileVm.info.email = data['info'][0]["company.email"]
      editProfileVm.info.invite_email = unescape(data['info'][0]["company.invite_email"])
      editProfileVm.info.reject_email = unescape(data['info'][0]["company.reject_email"])
    }, function (err) {
      console.log(err)
    });

    function valid(input) {
      if (!isValidEmail(input.email)) {
        return input.email + " isn't a valid email."
      }
      if (!input.name) {
        return "Name shouldn't be empty.";
      }
      if (!input.invite_email) {
        return "Invite email can't be empty.";
      }
      return true
    }

    function updateProfile() {
      var validateInput = valid(editProfileVm.info);
      if (validateInput != true) {
        SNACKBAR({
          message: validateInput,
          messageType: "error",
        })
        return
      }

      editProfileVm.info.invite_email = escape(editProfileVm.info.invite_email)
      editProfileVm.info.reject_email = escape(editProfileVm.info.reject_email)
      var requestData = angular.copy(editProfileVm.info);

        profileService.updateProfile(requestData)
          .then(function(data) {
            console.log(data)
            SNACKBAR({
              message: "Profile updated successfully.",
              messageType: "success",
            })
            $state.transitionTo("root")
          }, function(err) {
            console.log(err)
          })
    }
  }

  var profileDependency = [
    "$scope",
    "$rootScope",
    "profileService",
    profileController
  ];
  angular.module('GruiApp').controller('profileController', profileDependency);

  var editProfileDependency = [
    "$scope",
    "$rootScope",
    "$state",
    "profileService",
    editProfileController
  ];
  angular.module('GruiApp').controller('editProfileController', editProfileDependency);

})();
