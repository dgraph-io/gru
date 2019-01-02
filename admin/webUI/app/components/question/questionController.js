(function() {
  function questionController(
    $scope,
    $rootScope,
    $http,
    $q,
    $state,
    $stateParams,
    questionService
  ) {
    // VARIABLE DECLARATION
    mainVm.pageName = "question";
    questionVm = this;
    questionVm.optionsCount = 4;

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

    questionVm.editorSetting = {
      lineNumbers: true,
      lineWrapping: true,
      indentWithTabs: true
    };

    // FUNCTION DECLARATION
    questionVm.addNewTag = addNewTag;
    questionVm.validateInput = validateInput;
    questionVm.initCodeMirror = initCodeMirror;
    questionVm.initOptionEditor = initOptionEditor;
    questionVm.isCorrect = isCorrect;
    questionVm.getAllTags = getAllTags;
    questionVm.onTagSelect = onTagSelect;
    questionVm.markDownFormat = markDownFormat;
    questionVm.transitionToQuestion = transitionToQuestion;
    questionVm.getAllTags();

    // FUNCTION DEFINITION

    function initCodeMirror() {
      $scope.cmOption = {};
      setTimeout(function() {
        $scope.cmOption = questionVm.editorSetting;
      }, 500);
    }

    function initOptionEditor() {
      var setting = {};
      for (var i = 0; i < questionVm.optionsCount; i++) {
        setting["option" + i] = questionVm.editorSetting;
      }
      return setting;
    }

    $rootScope.$on("$viewContentLoaded", function() {
      // questionVm.initCodeMirror();
    });

    function markDownFormat(content) {
      return marked(content);
    }

    // Get all Tags
    function getAllTags() {
      mainVm.getAllTags().then(
        function(data) {
          if (!data || !data.data || !data.data.tags) {
            questionVm.allTags = [];
            return;
          }
          questionVm.allTags = data.data.tags;
        },
        function(err) {
          console.log(err);
        }
      );
    }

    function addNewTag(new_tag) {
      return {
        id: "",
        name: new_tag
      };
    }

    function validateInput(inputs) {
      if (!inputs.name) {
        return "Please enter valid question name";
      }
      if (!inputs.text) {
        return "Please enter valid question text";
      }
      if (inputs.positive == null || isNaN(inputs.positive)) {
        return "Please enter valid positve marks";
      }
      if (inputs.negative == null || isNaN(inputs.negative)) {
        return "Please enter valid negative marks";
      }
      if (Object.keys(inputs.options).length != questionVm.optionsCount) {
        return "Please enter all the options";
      }

      hasCorrectAnswer = false;
      hasEmptyName = false;
      correct = 0;
      angular.forEach(inputs.options, function(value, key) {
        if (value.is_correct) {
          hasCorrectAnswer = true;
          correct++;
        }
        if (!value.name) {
          hasEmptyName = true;
        }
      });
      if (hasEmptyName) {
        return "Please enter option name correctly";
      }
      if (!hasCorrectAnswer) {
        return "Please mark at least one correct answer";
      }

      if (!inputs.tags.length) {
        return "Minimum one tag is required";
      }
      if (correct > 1 && inputs.negative < inputs.positive) {
        return "For questions with multiple correct answers, negative score should be more than positive.";
      }

      return false;
    }

    function isCorrect(option, correct_options) {
      var uid = option.uid;
      if (!correct_options) {
        return false;
      }
      var optLength = correct_options.length;

      for (var i = 0; i < optLength; i++) {
        if (correct_options[i].uid == uid) {
          option.is_correct = true;
          return true;
        }
      }
      return false;
    }

    function onTagSelect(item, model, isEdit) {
      for (var i = 0; i < questionVm.allTags.length; i++) {
        if (item.name == questionVm.allTags[i].name && !item.uid) {
          delete model.id;
          delete model.isTag;
          model.uid = questionVm.allTags[i].uid;
        }
      }

      if (isEdit) {
        $rootScope.$broadcast("onSelect", {
          item: item,
          model: model
        });
      }
    }

    function transitionToQuestion(questionId) {
      $state.transitionTo("question.all", {
        quesID: questionId
      });
    }
  }

  function addQuestionController(
    $scope,
    $rootScope,
    $http,
    $q,
    $state,
    $stateParams,
    questionService
  ) {
    // See if update
    addQueVm = this;
    addQueVm.newQuestion = {};
    addQueVm.newQuestion.tags = [];
    addQueVm.cmModel = "";

    setTimeout(function() {
      addQueVm.editor = questionVm.initOptionEditor();
    }, 500);

    $scope.$watch("addQueVm.cmModel", function(current, original) {
      addQueVm.outputMarked = marked(current);
    });

    //FUnction Declaration
    addQueVm.addQuestionForm = addQuestionForm;
    addQueVm.resetForm = resetForm;

    // Check if user is authorized
    function addQuestionForm() {
      var options = [];
      var newOptions = angular.copy(addQueVm.newQuestion.optionsBak);
      angular.forEach(newOptions, function(value, key) {
        if (!value.is_correct) {
          value.is_correct = false;
        }
        value.name = escape(value.name);
        options.push(value);
      });
      addQueVm.newQuestion.options = options;
      addQueVm.newQuestion.text = escape(addQueVm.cmModel);
      var validataionError = questionVm.validateInput(addQueVm.newQuestion);
      if (validataionError) {
        SNACKBAR({
          message: validataionError,
          messageType: "error"
        });
        return;
      }

      var requestData = angular.copy(addQueVm.newQuestion);
      requestData.notes = requestData.notes;

      // Hit the API
      questionService.saveQuestion(JSON.stringify(requestData)).then(
        function(data) {
          addQueVm.newQuestion = {};
          addQueVm.cmModel = "";

          if (data.code == "Error") {
            SNACKBAR({
              message: data.message,
              messageType: "error"
            });
          }
          if (data.Success) {
            SNACKBAR({
              message: data.Message
            });
          }
          $state.transitionTo("question.all");
        },
        function(err) {
          console.log(err);
        }
      );
    }

    $rootScope.$on("$viewContentLoaded", function() {
      questionVm.initCodeMirror();
    });

    questionVm.initCodeMirror();

    function resetForm() {
      addQueVm.cmModel = "";
      addQueVm.newQuestion = {};
      addQueVm.newQuestion.tags = [];
      $(".mdl-textfield.is-dirty").removeClass("is-dirty");
    }
  }

  function allQuestionController(
    $scope,
    $rootScope,
    $http,
    $q,
    $state,
    $stateParams,
    questionService
  ) {
    // VARIABLE DECLARATION
    allQVm = this;
    allQVm.showLazyLoader = false;
    allQVm.noItemFound = false;
    mainVm.allQuestions = [];

    // FUNCTION DECLARATION
    allQVm.getAllQuestions = getAllQuestions;
    allQVm.getQuestion = getQuestion;
    allQVm.setQuestion = setQuestion;
    allQVm.toggleFilter = toggleFilter;
    allQVm.filterBy = filterBy;
    allQVm.removeAllFilter = removeAllFilter;
    allQVm.setFirstQuestion = setFirstQuestion;
    allQVm.searchText = "";

    // INITITIALIZERS
    allQVm.getAllQuestions();
    questionVm.getAllTags();

    // FUNCTION DEFINITIONS
    function getAllQuestions() {
      allQVm.showLazyLoader = true;
      // TODO - Fetch only meta for all questions, and details for just the
      // first question because we anyway refetch the questions from the server
      // on click. In the meta call, fetch tags and multiple status too so that
      // we can do filtering based on that.
      questionService.getAllQuestions(false).then(
        function(questions) {
          allQVm.showLazyLoader = false;

          if (!questions) {
            allQVm.noItemFound = true;
            return;
          }

          if (questions) {
            if (mainVm.allQuestions && mainVm.allQuestions.length) {
              for (var i = 0; i < questions.length; i++) {
                mainVm.allQuestions.push(questions[i]);
              }
            } else {
              mainVm.allQuestions = questions;
            }
            var questionIndex = -1;
            if ($stateParams.quesID) {
              questionIndex = mainVm.allQuestions.findIndex(function(q) {
                return q.uid == $stateParams.quesID;
              });
            }
            questionIndex = Math.max(questionIndex, 0)
            allQVm.setQuestion(
              mainVm.allQuestions[questionIndex],
              questionIndex,
            );
          }

          allQVm.lastQuestion =
            mainVm.allQuestions[mainVm.allQuestions.length - 1].uid;
        },
        function(err) {
          allQVm.showLazyLoader = false;
          console.log(err);
        }
      );
    }

    function getQuestion(questionId) {
      // When question is clicked on the side nav bar, we fetch its
      // information from backend and refresh it.
      questionService.getQuestion(questionId).then(function(question) {
        allQVm.question = question;
      });
    }

    function setQuestion(question, index) {
      allQVm.question = question;
      allQVm.questionIndex = index;
    }

    function toggleFilter(filter_value, key) {
      allQVm.filter = allQVm.filter || {};
      if (key == "tag") {
        allQVm.filter.tag || (allQVm.filter.tag = []);
        var tagIndex = mainVm.indexOfObject(allQVm.filter.tag, filter_value);
        // If tag is already there in our array, then we remove it.
        if (tagIndex > -1) {
          allQVm.filter.tag.splice(tagIndex, 1);
        } else {
          allQVm.filter.tag.push(filter_value);
        }
      }

      if (!key) {
        allQVm.filter[filter_value] = allQVm.filter[filter_value]
          ? false
          : true;
        if (filter_value == "multiple") {
          allQVm.filter.single = false;
        } else if (filter_value == "single") {
          allQVm.filter.multiple = false;
        }
      }

      allQVm.setFirstQuestion();
    }

    // TODO : Write modular code Filtering
    function filterBy(question) {
      textFilterMatch =
        question.name.toUpperCase().indexOf(allQVm.searchText.toUpperCase()) !=
        -1;

      if (allQVm.filter && allQVm.filter.tag && allQVm.filter.tag.length) {
        var found = false;
        var tagFound = true;
        var tagsLen = allQVm.filter.tag.length;
        for (var i = 0; i < tagsLen; i++) {
          var tagIndex = mainVm.indexOfObject(
            question["question.tag"],
            allQVm.filter.tag[i]
          );
          if (tagIndex == -1) {
            tagFound = false;
            break;
          }
          if (
            tagIndex > -1 &&
            (allQVm.filter.multiple && question["question.correct"].length == 1)
          ) {
            tagFound = false;
          }
          if (
            tagIndex > -1 &&
            (allQVm.filter.single && question["question.correct"].length > 1)
          ) {
            tagFound = false;
          }
          if (!tagFound) break;
        }
        return textFilterMatch && tagFound;
      } else if (allQVm.filter && allQVm.filter.multiple) {
        if (question["question.correct"].length > 1) {
          return textFilterMatch && true;
        } else {
          return textFilterMatch && false;
        }
      } else if (allQVm.filter && allQVm.filter.single) {
        if (question["question.correct"].length == 1) {
          return textFilterMatch && true;
        } else {
          return textFilterMatch && false;
        }
      } else {
        return textFilterMatch && true;
      }
    }

    function removeAllFilter() {
      delete allQVm.filter;
      allQVm.setFirstQuestion();
    }

    function setFirstQuestion() {
      setTimeout(function() {
        var question = $(".side-tabs");
        if (question.length) {
          question[0].click();
        }
      }, 300);
    }
  } // AllQuestionController

  function editQuestionController(
    $scope,
    $rootScope,
    $state,
    $stateParams,
    questionService
  ) {
    editQuesVm = this;
    editQuesVm.newQuestion = {};

    // Functin Declaratin
    editQuesVm.updateQuestionForm = updateQuestionForm;
    editQuesVm.onRemoveTag = onRemoveTag;
    editQuesVm.initMarkeDownPreview = initMarkeDownPreview;

    // INITIALIZERS
    questionVm.initCodeMirror();
    setTimeout(function() {
      editQuesVm.editor = questionVm.initOptionEditor();
    }, 500);

    questionVm.getAllTags();

    questionService.getQuestion($stateParams.quesID).then(
      function(question) {
        editQuesVm.newQuestion = question;
        editQuesVm.cmModel = unescape(editQuesVm.newQuestion.text);
        question.optionsBak = angular.copy(
          question["question.option"]
        );
        for (var i = 0; i < question.optionsBak.length; i++) {
          question.optionsBak[i].name = unescape(question.optionsBak[i].name);
        }
        editQuesVm.newQuestion.positive = parseFloat(question.positive);
        editQuesVm.newQuestion.negative = parseFloat(question.negative);

        editQuesVm.originalQuestion = angular.copy(editQuesVm.newQuestion);

        editQuesVm.initMarkeDownPreview();
      },
      function(err) {
        console.log(err);
      }
    );

    $rootScope.$on("onSelect", function(e, data) {
      if (data.model.is_delete) {
        var idx = mainVm.indexOfObject(
          editQuesVm.newQuestion.deletedTag,
          data.model
        );
        if (idx >= 0) {
          editQuesVm.newQuestion.deletedTag.splice(idx, 1);
        }
        data.model.is_delete = false;
      }
    });

    function updateQuestionForm() {
      if (editQuesVm.newQuestion.deletedTag) {
        editQuesVm.newQuestion.tags = editQuesVm.newQuestion[
          "question.tag"
        ].concat(editQuesVm.newQuestion.deletedTag);
      } else {
        editQuesVm.newQuestion.tags = editQuesVm.newQuestion["question.tag"];
      }

      editQuesVm.newQuestion.options = angular.copy(
        editQuesVm.newQuestion.optionsBak
      );
      editQuesVm.newQuestion.text = escape(editQuesVm.cmModel);

      angular.forEach(editQuesVm.newQuestion.options, function(value, key) {
        editQuesVm.newQuestion.options[key].name = escape(value.name);
      });
      var validataionError = questionVm.validateInput(editQuesVm.newQuestion);
      if (validataionError) {
        SNACKBAR({
          message: validataionError,
          messageType: "error"
        });
        return;
      }

      var deletedAllTags = true;
      var allTags = editQuesVm.newQuestion.tags;
      var tagsLength = allTags.length;
      for (var i = 0; i < tagsLength; i++) {
        if (!allTags[i].is_delete) {
          deletedAllTags = false;
        }
      }

      if (deletedAllTags) {
        SNACKBAR({
          message: "Please enter at least one tag",
          messageType: "error"
        });
        return;
      }

      var requestData = angular.copy(editQuesVm.newQuestion);
      requestData.notes = requestData.notes;
      questionService.editQuestion(requestData).then(
        function(data) {
          SNACKBAR({
            message: data.Message
          });
          $state.transitionTo("question.all", {
            quesID: $stateParams.quesID
          });
        },
        function(err) {
          console.log(err);
        }
      );
    }

    function resetQuestion(index) {
      editQuesVm.newQuestion = angular.copy(editQuesVm.originalQuestion);

      $rootScope.upgradeMDL();
    }

    function onRemoveTag(tag, model, question) {
      if (tag.uid) {
        tag.is_delete = true;
        question.deletedTag = question.deletedTag || [];
        question.deletedTag.push(tag);
      }
    }

    function initMarkeDownPreview() {
      $scope.$watch("editQuesVm.cmModel", function(current, original) {
        editQuesVm.outputMarked = marked(current, {
          gfm: true
        });
      });
    }
  }

  var editQuestionDependency = [
    "$scope",
    "$rootScope",
    "$state",
    "$stateParams",
    "questionService",
    editQuestionController
  ];
  angular
    .module("GruiApp")
    .controller("editQuestionController", editQuestionDependency);

  var allQuestionDependency = [
    "$scope",
    "$rootScope",
    "$http",
    "$q",
    "$state",
    "$stateParams",
    "questionService",
    allQuestionController
  ];
  angular
    .module("GruiApp")
    .controller("allQuestionController", allQuestionDependency);

  var addQuestionDependency = [
    "$scope",
    "$rootScope",
    "$http",
    "$q",
    "$state",
    "$stateParams",
    "questionService",
    addQuestionController
  ];
  angular
    .module("GruiApp")
    .controller("addQuestionController", addQuestionDependency);

  var questionDependency = [
    "$scope",
    "$rootScope",
    "$http",
    "$q",
    "$state",
    "$stateParams",
    "questionService",
    questionController
  ];
  angular
    .module("GruiApp")
    .controller("questionController", questionDependency);
})();
