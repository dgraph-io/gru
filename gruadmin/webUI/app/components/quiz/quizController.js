(function(){

	function quizController($scope, $rootScope, $stateParams, $http, $state, quizService, questionService) {

	// VARIABLE DECLARATION
		mainVm.pageName = "quiz";
		quizVm = this;
		quizVm.newQuiz = {};

		var quiz = {
			debug: {
				quiz: [
					{
						"name": "Quiz for Backend Developer",
						"_uid_": "0x1232323434"
					},{
						"name": "Quiz for Frontend Developer",
						"_uid_": "0x145355y356"
					},{
						"name": "Quiz for Devops",
						"_uid_": "0x1465868434"
					},
				],
			}
		}

	// FUNCTION DECLARATION
		quizVm.removeSelecteQuestion = removeSelecteQuestion;
		quizVm.addQuizForm = addQuizForm;

	// FUNCTION DEFINITION
		
		// Function for fetching next question

		quizService.getAllQuizes().then(function(data){
			var data = JSON.parse(data);
			quizVm.allQuizes = data.debug[0].quiz;
		}, function(err){
			console.log(err);
		})

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

			if(!quizVm.newQuiz.questions) {
				alert("Please add questin to the quiz!");
				return;
			}
			var qustionsClone = angular.copy(quizVm.newQuiz.questions)
			angular.forEach(qustionsClone, function(value, key) {
			  questions.push(value._uid_);
			});
			requestData.questions = questions;

			quizService.saveQuiz(requestData)
			.then(function(data){
				quizVm.newQuiz = {}
				console.log(data);
				alert(data.Message);
			}, function(err){
				console.log(err);
			})
		}

	}

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