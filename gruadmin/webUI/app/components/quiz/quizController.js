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

			if(data.Success) {
				SNACKBAR({
					message: data.message,
					messageType: "error",
				})
			}

			var quizArr = [];
			var ifObject = mainVm.isObject(data.debug.quiz)
			if(ifObject) {
				tagsArr.push(data.debug.quiz)
				quizVm.allQuizes = tagsArr;
			} else {
				tagsArr = data.debug.quiz;
				quizVm.allQuizes = data.debug.quiz;
			}
		}, function(err){
			console.log(err);
		})

		questionService.getAllQuestions().then(function(data){
			var data = JSON.parse(data);

			var questionArr = [];
			var ifObject = mainVm.isObject(data.debug.question)
			if(ifObject) {
				questionArr.push(data.debug.question)
				mainVm.allQuestions = questionArr;
			} else {
				mainVm.allQuestions = data.debug.question;
			}
		}, function(err){
			console.log(err)
		})

		function removeSelecteQuestion(key) {
			delete quizVm.newQuiz.selectedQuestion[key];
			if(!Object.keys(quizVm.newQuiz.selectedQuestion).length) {
				delete quizVm.newQuiz.selectedQuestion;
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