angular.module('GruiApp').controller('profileController', [
  "$scope",
  "$rootScope",
  "profileService",
  function profileController($scope, $rootScope, profileService) {
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
        // in case there is code without language specified
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

    profileVm.initCodeMirror = function initCodeMirror() {
      $scope.cmOption = {}
      $scope.cmOption2 = {}
      setTimeout(function() {
        $scope.cmOption = profileVm.editorSetting;
        $scope.cmOption2 = profileVm.editorSetting;
        // To refresh the reject mail editor which is hidden initially.
        $(".CodeMirror").length > 0 && $(".CodeMirror")[1].CodeMirror.refresh()
      }, 500);
    };
  }
]);


angular.module('GruiApp').controller('editProfileController', [
  "$scope",
  "$rootScope",
  "$state",
  "profileService",
  function editProfileController($scope, $rootScope, $state, profileService) {
    editProfileVm = this;
    editProfileVm.update = updateProfile;
    editProfileVm.info = {};

    profileService.getProfile()
      .then(function(data) {
        editProfileVm.info = {}
        if (!data.info || !data.info[0]) {
          return;
        }
        var info = data.info[0];

        editProfileVm.info.uid = info.uid;
        editProfileVm.info.name = info["company.name"]
        editProfileVm.info.email = info["company.email"]
        editProfileVm.info.invite_email = decodeURI(info["company.invite_email"])
        editProfileVm.info.invite_email = editProfileVm.info.invite_email === "undefined" ? "" : editProfileVm.info.invite_email
        editProfileVm.info.reject_email = decodeURI(info["company.reject_email"])
        editProfileVm.info.reject_email = editProfileVm.info.reject_email === "undefined" ? "" : editProfileVm.info.reject_email
        editProfileVm.info.reject = info["company.reject"] === "true"
        editProfileVm.info.backup = parseInt(info.backup)
        editProfileVm.info.backup_days = parseInt(info.backup_days)
        editProfileVm.info.backup = isNaN(editProfileVm.info.backup) ? 1 : editProfileVm.info.backup
        editProfileVm.info.backup_days = isNaN(editProfileVm.info.backup_days) ? 5 : editProfileVm.info.backup_days
      }, function(err) {
        console.error(err)
      });

    function valid(input) {
      if (!input.name) {
        return "Name shouldn't be empty.";
      }
      if (!isValidEmail(input.email)) {
        return "Please enter a valid email."
      }
      if (input.reject && !input.reject_email) {
        return "Rejection email can't be empty.";
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
          SNACKBAR({
            message: "Profile updated successfully.",
            messageType: "success",
          })
          $state.transitionTo("root")
        }, function(err) {
          console.error(err)
          SNACKBAR({
            message: "Something went wrong: " + JSON.stringify(err),
            messageType: "error",
          })
        })
    }

    $rootScope.$on('$viewContentLoaded', function() {
      profileVm.initCodeMirror();
    });

    profileVm.initCodeMirror();
  }
]);
