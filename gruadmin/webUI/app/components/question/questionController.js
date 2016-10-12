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
		
		// See if update
		var qid = $stateParams.qid;

	// FUNCTION DECLARATION
		questionVm.addNewTag = addNewTag;
		questionVm.validateInput = validateInput;
		questionVm.initCodeMirror = initCodeMirror;
		questionVm.isCorrect = isCorrect;
		questionVm.getAllTags = getAllTags;
		questionVm.getUniqueTags = getUniqueTags;
		questionVm.onTagSelect = onTagSelect;
		questionVm.markDownFormat = markDownFormat;
		questionVm.getAllTags();

	// FUNCTION DEFINITION

		function initCodeMirror(){
			$scope.cmOption = {}
			setTimeout(function() {
				$scope.cmOption = {
			    lineNumbers: true,
			    indentWithTabs: true,
			    // mode: 'javascript',
			  }
			}, 200);
		}

		function markDownFormat(content) {
			return marked(content);
		}

		// Get all Tags
		function getAllTags() {
			mainVm.getAllTags().then(function(data){
				var data = JSON.parse(data);
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
			// var allUniqueTags = {};
			// var tagsLength = allTags.length;
			// for(var i = 0; i < tagsLength; i ++ ) { 
			// 	allUniqueTags[allTags[i]._uid_] = allTags[i];
			// }
			return allUniqueTags;
		}

		function validateInput(inputs) {
			console.log(inputs)
			if(!inputs.name) {
				return "Please enter valid question name"
			}
			if(!inputs.text) {
				return "Please enter valid question text"
			}
			if(!inputs.positive) {
				return "Please enter valid positve marks"
			}
			if(!inputs.negative) {
				return "Please enter valid negative marks"
			}
			if(Object.keys(inputs.options).length != questionVm.optionsCount) {
				return "Please enter all the options"
			}

			hasCorrectAnswer = false
			angular.forEach(inputs.options, function(value, key) {
				if(value.is_correct) {
					hasCorrectAnswer = true;
				}
			});
			if(!hasCorrectAnswer) {
				return "Please mark atleast one correct answer"
			}

			if(!inputs.tags.length) {
				return "Please minimum one tag is required"
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

		function onTagSelect(item, model) {
			for(var i = 0; i < questionVm.allTags.length; i++) { 
				if(item.name == questionVm.allTags[i].name && !item._uid_){ 
					delete model.id;
					delete model.isTag;
					model._uid_ = questionVm.allTags[i]._uid_;
				} 
			}
		}
	}

	function addQuestionController($scope, $rootScope, $http, $q, $state, $stateParams, questionService) {
		// See if update
		addQueVm = this;
		addQueVm.newQuestion = {};
		addQueVm.newQuestion.tags = [];
		addQueVm.cmModel = "";

		$scope.$watch('addQueVm.cmModel', function(current, original) {
        addQueVm.outputMarked = marked(current);
    });

		//FUnction Declaration
		addQueVm.addQuestionForm = addQuestionForm;

		// Check if user is authorized
		function addQuestionForm() {
			var options = []
			var newOptions = angular.copy(addQueVm.newQuestion.optionsBak)
			angular.forEach(newOptions, function(value, key) {
				if(!value.is_correct) {
					value.is_correct = false;
				}
			  options.push(value);
			});
			addQueVm.newQuestion.options = options;
			addQueVm.newQuestion.text = escape(addQueVm.cmModel);
			areInValidateInput = questionVm.validateInput(addQueVm.newQuestion);
			if(areInValidateInput) {
				SNACKBAR({
					message: areInValidateInput,
					messageType: "error",
				})
				return
			}

			// Hit the API
			questionService.saveQuestion(JSON.stringify(addQueVm.newQuestion))
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
						// messageType: "error",
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

		// INITITIALIZERS
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
			questionService.getAllQuestions(data, hideLoader).then(function(data){
				var data = JSON.parse(data);
				if(data.code == "ErrorInvalidRequest" || !data.debug[0].question) {
					allQVm.noItemFound = true;
					allQVm.showLazyLoader = false;
					return
				}

				var dataArray = data.debug[0].question;
				if(data.debug && dataArray) {
					if(mainVm.allQuestions) {
						dataArrayLength = dataArray.length;
						for(var i = 0; i < dataArrayLength; i++) {
               mainVm.allQuestions.push(dataArray[i]);
            }
					} else {
						mainVm.allQuestions = data.debug[0].question;
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

	} // AllQuestionController

	function editQuestionController($scope, $state, $stateParams, questionService) {
		editQuesVm = this;
		editQuesVm.newQuestion = {};

		// Functin Declaratin
		editQuesVm.updateQuestionForm = updateQuestionForm;
		editQuesVm.onRemoveTag = onRemoveTag;
		editQuesVm.initMarkeDownPreview = initMarkeDownPreview;
		
		// INITIALIZERS
		questionVm.initCodeMirror();
		questionVm.getAllTags();

		questionService.getQuestion($stateParams.quesID)
		.then(function(data){
			editQuesVm.newQuestion = data.root[0];
			editQuesVm.cmModel = unescape(editQuesVm.newQuestion.text);
			editQuesVm.newQuestion.optionsBak = angular.copy(data.root[0]['question.option']);
			editQuesVm.newQuestion.positive = parseFloat(data.root[0].positive);
			editQuesVm.newQuestion.negative = parseFloat(data.root[0].negative);

			editQuesVm.originalQuestion = angular.copy(editQuesVm.newQuestion);

			editQuesVm.initMarkeDownPreview();
		}, function(err){
			console.log(err);
		})

		function updateQuestionForm() {

			if(editQuesVm.newQuestion.deletedTag) {
				editQuesVm.newQuestion.tags = editQuesVm.newQuestion["question.tag"].concat(editQuesVm.newQuestion.deletedTag);
			} else {
				editQuesVm.newQuestion.tags = editQuesVm.newQuestion["question.tag"];
			}

			editQuesVm.newQuestion.options = editQuesVm.newQuestion.optionsBak;
			editQuesVm.newQuestion.text = escape(editQuesVm.cmModel);
			console.log(editQuesVm.newQuestion);
			questionService.editQuestion(editQuesVm.newQuestion).then(function(data){
				SNACKBAR({
					message: data.Message,
				});
				$state.transitionTo("question.all")
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
	        editQuesVm.outputMarked = marked(current);
	    });
		}
	}

	var editQuestionDependency = [
			"$scope",
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