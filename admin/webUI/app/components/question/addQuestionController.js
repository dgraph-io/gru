angular.module("GruiApp").controller("addQuestionController", [
  "$scope",
  "$rootScope",
  "$http",
  "$state",
  "$stateParams",
  "questionService",
  function addQuestionController(
    $scope,
    $rootScope,
    $http,
    $state,
    $stateParams,
    questionService
  ) {
    addQueVm = this;

    addQueVm.loadEmptyQuestion = function() {
      addQueVm.newQuestion = {
        options: [],
        tags: [],
      };
      for (var i = 0; i < questionVm.optionsCount; i++) {
        addQueVm.newQuestion.options.push({
          is_correct: false,
          name: '',
        });
      }
    }

    addQueVm.loadEmptyQuestion();

    setTimeout(function() {
      addQueVm.editor = questionVm.initOptionEditor();
    }, 500);

    addQueVm.markdownPreview = function() {
      return marked(addQueVm.newQuestion.text || "");
    }

    addQueVm.submitForm = function() {
      var validataionError = questionVm.validateInput(addQueVm.newQuestion);
      if (validataionError) {
        SNACKBAR({
          message: validataionError,
          messageType: "error"
        });
        return;
      }

      questionService.saveQuestion(addQueVm.newQuestion).then(
        function(data) {
          if (data.code != "Error") {
            addQueVm.loadEmptyQuestion();
          }

          SNACKBAR({
            message: data.message || data.Message,
            messageType: data.code == "Error" ? "error" : "success",
          });

          $state.transitionTo("question.all");
        });
    }

    $rootScope.$on("$viewContentLoaded", function() {
      questionVm.initCodeMirror();
    });
    questionVm.initCodeMirror();

    addQueVm.resetForm = function() {
      addQueVm.loadEmptyQuestion();
    }
  }
]);
