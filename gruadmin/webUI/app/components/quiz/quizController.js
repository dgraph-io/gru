(function(){

	function quizController($scope, $rootScope, $stateParams, $http, $state, quizService, questionService) {

	// VARIABLE DECLARATION
		mainVm.pageName = "quiz";
		quizVm = this;
		quizVm.newQuiz = {};

	// FUNCTION DECLARATION
		quizVm.removeSelectedQuestion = removeSelectedQuestion;
		quizVm.addQuizForm = addQuizForm;
		quizVm.validateInput = validateInput;

	// FUNCTION DEFINITION
		
		// Function for fetching next question

		questionService.getAllQuestions().then(function(data){
			var data = JSON.parse(data);
			mainVm.allQuestions = data.debug[0].question;			
		}, function(err){
			console.log(err)
		})

		function removeSelectedQuestion(key) {
			delete quizVm.newQuiz.questions[key];
			if(!Object.keys(quizVm.newQuiz.questions).length) {
				delete quizVm.newQuiz.questions;
			}
		}

		function addQuizForm() {
			var questions = []
			var requestData = {};
			requestData = angular.copy(quizVm.newQuiz);
			console.log(requestData);

			areInValidateInput = quizVm.validateInput(requestData);
			if(areInValidateInput) {
				SNACKBAR({
					message: areInValidateInput,
					messageType: "error",
				})
				return
			}
			var qustionsClone = angular.copy(quizVm.newQuiz.questions)
			angular.forEach(qustionsClone, function(value, key) {
			  questions.push({_uid_: value._uid_});
			});
			requestData.questions = questions;

			quizService.saveQuiz(requestData)
			.then(function(data){
				quizVm.newQuiz = {}
				SNACKBAR({
					message: data.Message,
					messageType: "error",
				})
				$state.transitionTo("quiz.all");
			}, function(err){
				console.log(err);
			})
		}

		function validateInput(inputs) {
			if(!inputs.name) {
				return "Please enter valid Quiz name"
			}
			if(!inputs.duration) {
				return "Please enter valid Duration"
			}
			if(!inputs.duration) {
				return "Please enter valid Duration"
			}
			if(!inputs.start_date) {
				return "Please enter valid Start date"
			}
			if(!inputs.end_date) {
				return "Please enter valid End date"
			}	
			if(!inputs.questions) {
				return "Please add question to the quiz before submitting"
			}

			return false
		}

	}

	function allQuizController(quizService, questionService) {
		quizService.getAllQuizes().then(function(data){
			var data = JSON.parse(data);
			quizVm.allQuizes = data.debug[0].quiz;
		}, function(err){
			console.log(err);
		})
	}

	function editQuizController($rootScope, $stateParams, $state, quizService) {
		editQuizVm = this;
		quizVm.newQuiz = {};

		// Function Declaration
		editQuizVm.editQuiz = editQuiz;
		editQuizVm.onQuestionRemove = onQuestionRemove;


		quizService.getQuiz($stateParams.quizID)
		.then(function(data){
			quizVm.newQuiz = data.root[0];

			selectedQuestion = data.root[0]['quiz.question'];
		}, function(err){
			console.log(err);
		});

		function editQuiz() {
			quizVm.newQuiz.questions = quizVm.newQuiz['quiz.question'];
			areInValidateInput = quizVm.validateInput(quizVm.newQuiz);
			if(areInValidateInput) {
				SNACKBAR({
					message: areInValidateInput,
					messageType: "error",
				})
				return
			}
			console.log(quizVm.newQuiz);

			quizService.editQuiz(quizVm.newQuiz)
			.then(function(data){
				SNACKBAR({
					message: data.Message,
					messageType: "error",
				});

				$state.transitionTo("quiz.all");
			}, function(err){
				console.log(err);
			})

		}

		function onQuestionRemove(question) {
			if(question._uid_) {
				question.is_delete = true;
			}
			console.log(question);
		}

		// editQuizVm.isSelected = function(question_id) {
		// 	for(var i=0; i<selectedQuestion.length; i++) {
		// 		if(selectedQuestion[i]._uid_ == question_id) {
		// 			return true;
		// 		}
		// 	}
		// 	return false;
		// }
	}

	var editQuizDependency = [
	    "$rootScope",
	    "$stateParams",
	    "$state",
	    "quizService",
	    editQuizController
	];
	angular.module('GruiApp').controller('editQuizController', editQuizDependency);

	var allQuizDependency = [
	    "quizService",
	    "questionService",
	    allQuizController
	];
	angular.module('GruiApp').controller('allQuizController', allQuizDependency);

	var quizDependency = [
	    "$scope",
	    "$rootScope",
	    "$stateParams",
	    "$http",
	    "$state",
	    "quizService",
	    "questionService",
	    quizController
	];
	angular.module('GruiApp').controller('quizController', quizDependency);

})();