(function(){

	function quizController($scope, $rootScope, $stateParams, $http, $state, quizService, questionService) {

	// VARIABLE DECLARATION
		mainVm.pageName = "quiz";
		quizVm = this;
		quizVm.newQuiz = {};

	// FUNCTION DECLARATION
		quizVm.removeSelecteQuestion = removeSelecteQuestion;
		quizVm.addQuizForm = addQuizForm;

	// FUNCTION DEFINITION
		
		// Function for fetching next question

		questionService.getAllQuestions().then(function(data){
			var data = JSON.parse(data);
			mainVm.allQuestions = data.debug[0].question;			
		}, function(err){
			console.log(err)
		})

		function removeSelecteQuestion(key) {
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

			areInValidateInput = validateInput(requestData);
			if(areInValidateInput) {
				SNACKBAR({
					message: areInValidateInput,
					messageType: "error",
				})
				return
			}
			var qustionsClone = angular.copy(quizVm.newQuiz.questions)
			angular.forEach(qustionsClone, function(value, key) {
			  questions.push(value._uid_);
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

	function allQuizController($scope, $rootScope, $stateParams, $http, $state, quizService, questionService) {

		quizService.getAllQuizes().then(function(data){
			var data = JSON.parse(data);
			quizVm.allQuizes = data.debug[0].quiz;
		}, function(err){
			console.log(err);
		})

	}

	var allQuizDependency = [
	    "$scope",
	    "$rootScope",
	    "$stateParams",
	    "$http",
	    "$state",
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