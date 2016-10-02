(function(){

	function questionController($scope, $rootScope, $http, $q, $state, $stateParams, questionService) {

	// VARIABLE DECLARATION
		mainVm.pageName = "question"
		questionVm = this;
		questionVm.newQuestion = {};
		questionVm.optionsCount = 4;
		questionVm.newQuestion.tags = [];
		
		// See if update
		var qid = $stateParams.qid;

	// FUNCTION DECLARATION
		questionVm.addQuestionForm = addQuestionForm;
		questionVm.addNewTag = addNewTag;
		questionVm.validateInput = validateInput;

	// FUNCTION DEFINITION

		// Get all Tags
		mainVm.getAllTags().then(function(data){
			var data = JSON.parse(data);
			questionVm.allTags = getUniqueTags(data.debug[0].question);
		}, function(err){
			console.log(err);
		})	

		// Check if user is authorized
		function addQuestionForm() {
			var options = []
			var newOptions = angular.copy(questionVm.newQuestion.optionsBak)
			angular.forEach(newOptions, function(value, key) {
				if(!value.is_correct) {
					value.is_correct = false;
				}
			  options.push(value);
			});
			questionVm.newQuestion.options = options;
			areInValidateInput = validateInput(questionVm.newQuestion);
			if(areInValidateInput) {
				SNACKBAR({
					message: areInValidateInput,
					messageType: "error",
				})
				return
			}

			// Hit the API
			questionService.saveQuestion(questionVm.newQuestion)
			.then(function(data){
				questionVm.newQuestion = {};

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
						allUniqueTags.push(tagsArr[j]);
					} 
				} 
			}
			return allUniqueTags;
		}

		function validateInput(inputs) {
			console.log(inputs)
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
	}

	function addQuestionController($scope, $rootScope, $http, $q, $state, $stateParams, questionService) {
		// See if update
		var qid = $stateParams.qid;
		var idx = $stateParams.index;
		questionVm.is_update = false;
		questionVm.update_label = "Add";
		if(mainVm.allQuestions && mainVm.allQuestions[idx]) {
			questionVm.is_update = true;
			questionVm.update_label = "Update";

			var current_question = mainVm.allQuestions[idx];
			questionVm.newQuestion = current_question;
			questionVm.newQuestion.positive = parseFloat(current_question.positive);
			questionVm.newQuestion.negative = parseFloat(current_question.negative);

			var qOptions = current_question["question.option"]
			questionVm.newQuestion.options = {};
			for(var i=0; i<qOptions.length; i++) {
				questionVm.newQuestion.options['option' + i] = {};
				questionVm.newQuestion.options['option' + i].name = qOptions[i].name;

				var correct_options = current_question["question.correct"];
				if(mainVm.isObject(correct_options)) {
					if(current_question["question.correct"].name == qOptions[i].name) {
						questionVm.newQuestion.options['option' + i].is_correct = true;
					}
				} else {
					for(var j = 0; j < correct_options.length; j++) {
						if(correct_options[j].name == qOptions[i].name) {
							questionVm.newQuestion.options['option' + i].is_correct = true;
						}
					}
				}
			}

			if(mainVm.isObject(current_question["question.tag"])) {
				var tagArr = []
				tagArr.push(current_question["question.tag"])
				questionVm.newQuestion.tags = tagArr;
			} else {
				questionVm.newQuestion.tags = current_question["question.tag"];
			}
		}
	}

	function allQuestionController($scope, $rootScope, $http, $q, $state, $stateParams, questionService) {
		allQVm = this;
		allQVm.updatedQuestion = {}; //Updated question modal
		allQVm.updatedQuestion.options = {};

		questionService.getAllQuestions().then(function(data){
			var data = JSON.parse(data);

			mainVm.allQuestions = data.debug[0].question;
			mainVm.allQuestionsBak = angular.copy(mainVm.allQuestions);
		}, function(err){
			console.log(err)
		})

		allQVm.editQuestion = editQuestion;
		allQVm.isCorrect = isCorrect;
		allQVm.resetQuestion = resetQuestion;
		allQVm.onRemoveTag = onRemoveTag;

		function editQuestion(index, question) {
			if(isNaN(question.positive) || isNaN(question.negative)) {
				SNACKBAR({
					message: "Question marks must be a number!",
					messageType: "error",
				})
				return
			}


			if(question.deletedTag) {
				question.tags = question["question.tag"].concat(question.deletedTag);
			} else {
				question.tags = question["question.tag"];
			}
			question.options = question["question.option"]
			question.positive = parseFloat(question.positive)
			question.negative = parseFloat(question.negative)
			
			questionService.editQuestion(question).then(function(data){
				SNACKBAR({
					message: data.Message,
				})
			}, function(err){
				console.log(err);
			})
		}

		function onRemoveTag(tag, model, question) {
			if(tag._uid_) {
				tag.is_delete = true;
				question.deletedTag = question.deletedTag || [];
				question.deletedTag.push(tag);
			}
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

		function resetQuestion(index) {
			$scope.editQuestion = false;
			mainVm.allQuestions[index] = angular.copy(mainVm.allQuestionsBak[index]);

			$rootScope.updgradeMDL();
		}

	}

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