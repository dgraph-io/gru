angular.module("GruiApp").controller("questionController", [
  "$scope",
  "$rootScope",
  "$http",
  "$state",
  "$stateParams",
  "questionService",
  "MainService",
  function questionController(
    $scope,
    $rootScope,
    $http,
    $state,
    $stateParams,
    questionService,
    MainService
  ) {
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

    questionVm.isCorrect = isCorrect;
    questionVm.onTagSelect = onTagSelect;

    questionVm.markDownFormat = function(content) {
      return marked(content);
    }

    questionVm.initCodeMirror = function() {
      $scope.cmOption = {};
      setTimeout(function() {
        $scope.cmOption = questionVm.editorSetting;
      }, 500);
    }

    questionVm.initOptionEditor = function() {
      var setting = {};
      for (var i = 0; i < questionVm.optionsCount; i++) {
        setting["option" + i] = questionVm.editorSetting;
      }
      return setting;
    }

    questionVm.getAllTags = function() {
      MainService.get("/get-all-tags").then(
        function(data) {
          if (!data || !data.data || !data.data.tags) {
            questionVm.allTags = [];
            return;
          }
          questionVm.allTags = data.data.tags;
        });
    }

    questionVm.getAllTags();

    questionVm.addNewTag = function(new_tag) {
      return {
        id: "",
        name: new_tag
      };
    }

    questionVm.validateInput = function(question) {
      if (!question.name) {
        return "Please enter valid question name";
      }
      if (!question.text) {
        return "Please enter valid question text";
      }
      if (question.positive == null || isNaN(question.positive)) {
        return "Please enter valid positve marks";
      }
      if (question.negative == null || isNaN(question.negative)) {
        return "Please enter valid negative marks";
      }
      if (Object.keys(question.options).length != questionVm.optionsCount) {
        return "Please enter all the options";
      }

      hasCorrectAnswer = false;
      correct = 0;
      angular.forEach(question.options, function(value) {
        if (value.is_correct) {
          hasCorrectAnswer = true;
          correct++;
        }
        if (!value.name) {
          return "Please enter option name correctly";
        }
      });
      if (!hasCorrectAnswer) {
        return "Please mark at least one correct answer";
      }

      if (!question.tags || !question.tags.length) {
        return "Minimum one tag is required";
      }
      if (correct > 1 && question.negative < question.positive) {
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

    function onTagSelect(item, model) {
      for (var i = 0; i < questionVm.allTags.length; i++) {
        if (item.name == questionVm.allTags[i].name && !item.uid) {
          delete model.id;
          delete model.isTag;
          model.uid = questionVm.allTags[i].uid;
        }
      }
    }
  }
]);

angular.module("GruiApp").controller("allQuestionController", [
  "$scope",
  "$rootScope",
  "$http",
  "$state",
  "$stateParams",
  "questionService",
  function allQuestionController(
    $scope,
    $rootScope,
    $http,
    $state,
    $stateParams,
    questionService
  ) {
    allQVm = this;
    allQVm.showLazyLoader = false;
    mainVm.allQuestions = [];

    allQVm.getAllQuestions = getAllQuestions;
    allQVm.getQuestion = getQuestion;
    allQVm.toggleFilter = toggleFilter;
    allQVm.filterBy = filterBy;
    allQVm.removeAllFilter = removeAllFilter;
    allQVm.setFirstQuestion = setFirstQuestion;
    allQVm.searchText = "";

    allQVm.getAllQuestions();
    questionVm.getAllTags();

    function getAllQuestions() {
      allQVm.showLazyLoader = true;

      questionService.getAllQuestions(false).then(
        function(questions) {
          allQVm.showLazyLoader = false;

          if (!questions) {
            mainVm.allQuestions = [];
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
            allQVm.question = mainVm.allQuestions[questionIndex]
            allQVm.questionIndex = questionIndex;
          }
        },
        function(err) {
          allQVm.showLazyLoader = false;
          console.error(err);
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
        for (var i = 0; i < allQVm.filter.tag.length; i++) {
          var tagIndex = mainVm.indexOfObject(
            question.tags,
            allQVm.filter.tag[i]
          );
          if (tagIndex == -1) {
            tagFound = false;
            break;
          }
          if (
            tagIndex > -1 &&
            (allQVm.filter.multiple && question.correct.length == 1)
          ) {
            tagFound = false;
          }
          if (
            tagIndex > -1 &&
            (allQVm.filter.single && question.correct.length > 1)
          ) {
            tagFound = false;
          }
          if (!tagFound) break;
        }
        return textFilterMatch && tagFound;
      } else if (allQVm.filter && allQVm.filter.multiple) {
        if (question.correct.length > 1) {
          return textFilterMatch && true;
        } else {
          return textFilterMatch && false;
        }
      } else if (allQVm.filter && allQVm.filter.single) {
        return (question.correct.length == 1) && !!textFilterMatch;
      } else {
        return !!textFilterMatch;
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
  }
]);
