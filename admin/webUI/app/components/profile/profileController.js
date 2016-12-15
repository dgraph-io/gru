(function() {

  function profileController($scope, $rootScope, profileService) {

    // VARIABLE DECLARATION
    profileVm = this;
    mainVm.pageName = "profile-page"

    marked.setOptions({
      renderer: new marked.Renderer(),
      gfm: true,
      tables: true,
      breaks: false,
      pedantic: false,
      sanitize: false, // if false -> allow plain old HTML ;)
      smartLists: true,
      smartypants: false,
      highlight: function(code, lang) {
        // in case, there is code without language specified
        if (lang) {
          return hljs.highlight(lang, code).value;
        } else {
          return hljs.highlightAuto(code).value;
        }
      }
    });

    profileVm.editorSetting = {
      lineWrapping: true
    };

    profileVm.initCodeMirror = initCodeMirror;

    function initCodeMirror() {
      $scope.cmOption = {}
      $scope.cmOption2 = {}
      setTimeout(function() {
        $scope.cmOption = profileVm.editorSetting;
        $scope.cmOption2 = profileVm.editorSetting;
        // To refresh the reject mail editor which is hidden initially.
        $(".CodeMirror").length > 0 && $(".CodeMirror")[1].CodeMirror.refresh()
      }, 500);
    }
  }

  function editProfileController($scope, $rootScope, $state, profileService) {
    editProfileVm = this;
    editProfileVm.update = updateProfile;
    editProfileVm.info = {};

    profileService.getProfile()
      .then(function(data) {
        editProfileVm.info = {}
        editProfileVm.info.name = data['info'][0]["company.name"]
        editProfileVm.info.email = data['info'][0]["company.email"]
        editProfileVm.info.invite_email = decodeURI(data['info'][0]["company.invite_email"])
        editProfileVm.info.invite_email = editProfileVm.info.invite_email === "undefined" ? "" : editProfileVm.info.invite_email
        editProfileVm.info.reject_email = decodeURI(data['info'][0]["company.reject_email"])
        editProfileVm.info.reject_email = editProfileVm.info.reject_email === "undefined" ? "" : editProfileVm.info.reject_email
        editProfileVm.info.reject = data['info'][0]["company.reject"] === "true"
        editProfileVm.info.backup = parseInt(data['info'][0]["backup"])
        editProfileVm.info.backup_days = parseInt(data['info'][0]["backup_days"])
        editProfileVm.info.backup = editProfileVm.info.backup === undefined ? 60 : editProfileVm.info.backup
        editProfileVm.info.backup_days = editProfileVm.info.backup_days === undefined ? 5 : editProfileVm.info.backup_days
      }, function(err) {
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

      editProfileVm.info.invite_email = encodeURI(editProfileVm.info.invite_email)
      editProfileVm.info.reject_email = encodeURI(editProfileVm.info.reject_email)
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

    $rootScope.$on('$viewContentLoaded', function() {
      profileVm.initCodeMirror();
    });

    profileVm.initCodeMirror();
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
