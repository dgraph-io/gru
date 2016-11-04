(function(){

	function questionController($scope, $rootScope, $http, $q, $state, $stateParams, questionService) {

	// VARIABLE DECLARATION
		mainVm.pageName = "question"
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
	    highlight: function (code, lang) {
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
		  indentWithTabs: true,
		};
		
	// FUNCTION DECLARATION
		questionVm.addNewTag = addNewTag;
		questionVm.validateInput = validateInput;
		questionVm.initCodeMirror = initCodeMirror;
		questionVm.initOptionEditor = initOptionEditor;
		questionVm.isCorrect = isCorrect;
		questionVm.getAllTags = getAllTags;
		questionVm.getUniqueTags = getUniqueTags;
		questionVm.onTagSelect = onTagSelect;
		questionVm.markDownFormat = markDownFormat;
		questionVm.transitionToQuestion = transitionToQuestion;
		questionVm.getAllTags();

	// FUNCTION DEFINITION

		function initCodeMirror(){
			$scope.cmOption = {}
			setTimeout(function() {
				$scope.cmOption = questionVm.editorSetting;
			}, 500);
		}

		function initOptionEditor() {
			var setting = {};
			for(var i = 0; i < questionVm.optionsCount; i++) {
				setting["option"+i] = questionVm.editorSetting;
			}
			return setting;
		}

		$rootScope.$on('$viewContentLoaded', function() {
      // questionVm.initCodeMirror();
    });

		function markDownFormat(content) {
			return marked(content);
		}

		// Get all Tags
		function getAllTags() {
			mainVm.getAllTags().then(function(data){
				questionVm.allTags = getUniqueTags(data.debug[0].question);
			}, function(err){
				console.log(err);
			})	
		}

		function addNewTag(new_tag) {
			return {
				"id": "",
				name: new_tag
			}
		}

		function getUniqueTags(allTags) {
			if(!allTags) {
				return []
			}
			var allUniqueTags = [];
			for(var i=0; i< allTags.length; i++) {
				var tagsArr = allTags[i]["question.tag"];
				if(tagsArr) {
					for (var j =0; j< tagsArr.length; j++) {
						isUnique = mainVm.indexOfObject(allUniqueTags, tagsArr[j]) == -1;
						if(isUnique) {
							allUniqueTags.push(tagsArr[j]);
						}
					}
				}
			}
			return allUniqueTags;
		}

		function validateInput(inputs) {
			if(!inputs.name) {
				return "Please enter valid question name"
			}
			if(!inputs.text) {
				return "Please enter valid question text"
			}
			if(inputs.positive == null || isNaN(inputs.positive)) {
				return "Please enter valid positve marks"
			}
			if(inputs.negative == null || isNaN(inputs.negative)) {
				return "Please enter valid negative marks"
			}
			if(Object.keys(inputs.options).length != questionVm.optionsCount) {
				return "Please enter all the options"
			}

			hasCorrectAnswer = false,
			hasEmptyName = false;
			correct = 0
			angular.forEach(inputs.options, function(value, key) {
				if(value.is_correct) {
					hasCorrectAnswer = true;
					correct++
				}
				if(!value.name) {
					hasEmptyName = true;
				}
			});
			if(hasEmptyName) {
				return "Please enter option name correctly"
			}
			if(!hasCorrectAnswer) {
				return "Please mark atleast one correct answer"
			}

			if(!inputs.tags.length) {
				return "Minimum one tag is required"
			}
			if(correct > 1 && inputs.negative < inputs.positive){
				return "For questions with multiple correct answers, negative score should be more than positive."
			}

			return false;
		}

		function isCorrect(option, correct_options) {
			var uid = option._uid_;
			if(!correct_options) {
				return false
			}
			var optLength = correct_options.length;

			for(var i = 0; i < optLength; i++) {
				if(correct_options[i]._uid_ == uid) {
					option.is_correct = true
					return true
				}
			}
			return false
		}

		function onTagSelect(item, model, isEdit) {
			for(var i = 0; i < questionVm.allTags.length; i++) { 
				if(item.name == questionVm.allTags[i].name && !item._uid_){ 
					delete model.id;
					delete model.isTag;
					model._uid_ = questionVm.allTags[i]._uid_;
				} 
			}

			if(isEdit) {
				$rootScope.$broadcast("onSelect", { 
					item: item,
					model: model
				});
			}
		}


		function transitionToQuestion(questionId) {
			$state.transitionTo("question.all", {
				quesID: questionId,
			})
		}
	}

	function addQuestionController($scope, $rootScope, $http, $q, $state, $stateParams, questionService) {
		// See if update
		addQueVm = this;
		addQueVm.newQuestion = {};
		addQueVm.newQuestion.tags = [];
		addQueVm.cmModel = "";

		setTimeout(function() {
			addQueVm.editor = questionVm.initOptionEditor();
		}, 500);

		$scope.$watch('addQueVm.cmModel', function(current, original) {
      addQueVm.outputMarked = marked(current);
    });

		//FUnction Declaration
		addQueVm.addQuestionForm = addQuestionForm;
		addQueVm.resetForm = resetForm;

		// Check if user is authorized
		function addQuestionForm() {
			var options = []
			var newOptions = angular.copy(addQueVm.newQuestion.optionsBak)
			angular.forEach(newOptions, function(value, key) {
				if(!value.is_correct) {
					value.is_correct = false;
				}
				value.name = escape(value.name);
			  options.push(value);
			});
			addQueVm.newQuestion.options = options;
			addQueVm.newQuestion.text = escape(addQueVm.cmModel);
			var areInValidateInput = questionVm.validateInput(addQueVm.newQuestion);
			if(areInValidateInput) {
				SNACKBAR({
					message: areInValidateInput,
					messageType: "error",
				})
				return
			}

			var requestData = angular.copy(addQueVm.newQuestion);
			requestData.notes = requestData.notes || "none";

			// Hit the API
			questionService.saveQuestion(JSON.stringify(requestData))
			.then(function(data){
				addQueVm.newQuestion = {};
				addQueVm.cmModel = "";

				if(data.code == "Error") {
					SNACKBAR({
						message: data.message,
						messageType: "error",
					})
				}
				if(data.Success) {
					SNACKBAR({
						message: data.Message,
					})
				}
				$state.transitionTo("question.all")
			}, function(err){
				console.log(err);
			})
		}

		$rootScope.$on('$viewContentLoaded', function() {
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

	function allQuestionController($scope, $rootScope, $http, $q, $state, $stateParams, questionService) {
		// VARIABLE DECLARATION
		allQVm = this;
		allQVm.showLazyLoader = false;
		allQVm.lazyStatus = true;
		allQVm.noItemFound = false;
		mainVm.allQuestions = [];

		// FUNCTION DECLARATION
		allQVm.getAllQuestions = getAllQuestions;
		allQVm.getQuestion = getQuestion;
		allQVm.setQuestion = setQuestion;

		// INITITIALIZERS
		console.log($stateParams);
		allQVm.getAllQuestions();
		questionVm.getAllTags();

		// FUNCTION DEFINITIONS
		function getAllQuestions(questionID) {

			if(questionID && !allQVm.lazyStatus && allQVm.noItemFound) {
				return 
			}
			// API HIT
			var data = {
				"id": questionID || "",
			};
			var hideLoader = questionID ? true : false;

			allQVm.showLazyLoader = true;
			// TODO - Fetch only meta for all questions, and details for just the
			// first question because we anyway refetch the questions from the server
			// on click. In the meta call, fetch tags and multiple status too so that
			// we can do filtering based on that.
			questionService.getAllQuestions(data, hideLoader).then(function(data){
				if(data.code == "ErrorInvalidRequest" || !data.debug[0].question) {
					allQVm.noItemFound = true;
					allQVm.showLazyLoader = false;
					return
				}

				var dataArray = data.debug[0].question;
				if(data.debug && dataArray) {
					if(mainVm.allQuestions && mainVm.allQuestions.length) {
						dataArrayLength = dataArray.length;
						for(var i = 0; i < dataArrayLength; i++) {
							mainVm.allQuestions.push(dataArray[i]);
						}
					} else {
						mainVm.allQuestions = data.debug[0].question;
						
						if($stateParams.quesID) {
							var length = mainVm.allQuestions.length;
							var gotQuestion = false;
							for(var i = 0; i < length; i++) {
								if(mainVm.allQuestions[i]._uid_ == $stateParams.quesID){
									var gotQuestion = true;
									allQVm.question = mainVm.allQuestions[i];
									break;
								}
							}
							if(!gotQuestion) {
								allQVm.setQuestion(mainVm.allQuestions[0], 0);
							}
						} else {
							allQVm.setQuestion(mainVm.allQuestions[0], 0);
						}
					}

					mainVm.allQuestionsBak = angular.copy(mainVm.allQuestions);
					allQVm.lastQuestion = mainVm.allQuestions[mainVm.allQuestions.length - 1]._uid_;
				}

				allQVm.lazyStatus = true;
				allQVm.showLazyLoader = false;
			}, function(err){
				allQVm.showLazyLoader = false;
				console.log(err)
			});
		}

		function getQuestion(questionId) {
			// When questionis clicked on the side nav bar, we fetch its 
			// information from backend and refresh it.
			questionService.getQuestion(questionId).then(function(data){
				allQVm.question = data.root[0];
			})

		}

		$(document).ready(function(){

			$(window).unbind('scroll');
			setTimeout(function() {
				window.addEventListener('scroll', function(){
				  if($("#question-listing").length){
				    var contentLen = $('.mdl-layout__content').scrollTop() + $('.mdl-layout__content').height();
				    var docHeight = getDocHeight("question-listing");
						if(contentLen >= docHeight && allQVm.lazyStatus && mainVm.allQuestions && mainVm.allQuestions.length) {
							allQVm.getAllQuestions(allQVm.lastQuestion);
							allQVm.lazyStatus = false;
						}
					}
				}, true);
			}, 100);

		});

		function setQuestion(question, index) {
			allQVm.question = question;
			allQVm.questionIndex = index;
		}

	} // AllQuestionController

	function editQuestionController($scope, $rootScope, $state, $stateParams, questionService) {
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

		questionService.getQuestion($stateParams.quesID)
		.then(function(data){
			editQuesVm.newQuestion = data.root[0];
			editQuesVm.cmModel = unescape(editQuesVm.newQuestion.text);
			editQuesVm.newQuestion.optionsBak = angular.copy(data.root[0]['question.option']);
			for(var i = 0; i < editQuesVm.newQuestion.optionsBak.length; i++) {
				editQuesVm.newQuestion.optionsBak[i].name = unescape(editQuesVm.newQuestion.optionsBak[i].name)
			}
			editQuesVm.newQuestion.positive = parseFloat(data.root[0].positive);
			editQuesVm.newQuestion.negative = parseFloat(data.root[0].negative);

			editQuesVm.originalQuestion = angular.copy(editQuesVm.newQuestion);

			editQuesVm.initMarkeDownPreview();
		}, function(err){
			console.log(err);
		})

		$rootScope.$on("onSelect", function(e, data) {
			if(data.model.is_delete) {
				var idx = mainVm.indexOfObject(editQuesVm.newQuestion.deletedTag, data.model);
				if (idx >= 0) {
					editQuesVm.newQuestion.deletedTag.splice(idx, 1)
				}
				data.model.is_delete =  false
			}
		})

		function updateQuestionForm() {

			if(editQuesVm.newQuestion.deletedTag) {
				editQuesVm.newQuestion.tags = editQuesVm.newQuestion["question.tag"].concat(editQuesVm.newQuestion.deletedTag);
			} else {
				editQuesVm.newQuestion.tags = editQuesVm.newQuestion["question.tag"];
			}

			editQuesVm.newQuestion.options = angular.copy(editQuesVm.newQuestion.optionsBak);
			editQuesVm.newQuestion.text = escape(editQuesVm.cmModel);

			angular.forEach(editQuesVm.newQuestion.options, function(value, key) {
				editQuesVm.newQuestion.options[key].name = escape(value.name);
			});
			var areInValidateInput = questionVm.validateInput(editQuesVm.newQuestion);
			if(areInValidateInput) {
				SNACKBAR({
					message: areInValidateInput,
					messageType: "error",
				})
				return
			}

			var deletedAllTags = true;
			var allTags = editQuesVm.newQuestion.tags;
			var tagsLength = allTags.length;
			for(var i = 0; i < tagsLength; i++) {
				if(!allTags[i].is_delete) {
					deletedAllTags =  false;
				}
			}

			if(deletedAllTags) {
				SNACKBAR({
					message: "Please enter at least one tag",
					messageType: "error",
				})
				return
			}

			var requestData = angular.copy(editQuesVm.newQuestion);
			requestData.notes = requestData.notes || "none";
			questionService.editQuestion(requestData).then(function(data){
				SNACKBAR({
					message: data.Message,
				});
				$state.transitionTo("question.all", {
					quesID: $stateParams.quesID,
				})
			}, function(err){
				console.log(err);
			})
		}

		function resetQuestion(index) {
			editQuesVm.newQuestion = angular.copy(editQuesVm.originalQuestion);

			$rootScope.updgradeMDL();
		}

		function onRemoveTag(tag, model, question) {
			if(tag._uid_) {
				tag.is_delete = true;
				question.deletedTag = question.deletedTag || [];
				question.deletedTag.push(tag);
			}
		}

		function initMarkeDownPreview() {
			$scope.$watch('editQuesVm.cmModel', function(current, original) {
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
	angular.module('GruiApp').controller('editQuestionController', editQuestionDependency);

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
	angular.module('GruiApp').controller('allQuestionController', allQuestionDependency);

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
	angular.module('GruiApp').controller('addQuestionController', addQuestionDependency);

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
	angular.module('GruiApp').controller('questionController', questionDependency);

})();
