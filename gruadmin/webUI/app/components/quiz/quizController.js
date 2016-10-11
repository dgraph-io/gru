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
		quizVm.getAllQuestions = getAllQuestions;

	// FUNCTION DEFINITION
		quizVm.getAllQuestions();
		
		// Function for fetching next question

		function getAllQuestions() {
			quesRequest = {id: ""};
			questionService.getAllQuestions(quesRequest).then(function(data){
				var data = JSON.parse(data);
				mainVm.allQuestions = data.debug[0].question;			
			}, function(err){
				console.log(err)
			});
		}

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

			requestData.duration = (requestData.hours || 0) + "h" + (requestData.minutes || 0) + "m" + (requestData.seconds || 0) + "s";
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
			if(!inputs.minutes && !inputs.hours) {
				return "Please enter valid time"
			}
			// if(!inputs.duration) {
			// 	return "Please enter valid Duration"
			// }
			// if(!inputs.start_date) {
			// 	return "Please enter valid Start date"
			// }
			// if(!inputs.end_date) {
			// 	return "Please enter valid End date"
			// }	
			if(!inputs.questions) {
				return "Please add question to the quiz before submitting"
			}

			return false
		}

	}

	function allQuizController(quizService, questionService) {
		quizVm.newQuiz = {};
		
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
		editQuizVm.selectedQuestion;

		// Function Declaration
		editQuizVm.editQuiz = editQuiz;
		editQuizVm.onQuestionRemove = onQuestionRemove;
		editQuizVm.onNewQuestionRemove = onNewQuestionRemove;
		editQuizVm.addNewQuestion = addNewQuestion;
		editQuizVm.getTimeObj = getTimeObj;

		// INITITALIZER
		quizVm.getAllQuestions();

		quizService.getQuiz($stateParams.quizID)
		.then(function(data){
			quizVm.newQuiz = data.root[0];

			editQuizVm.selectedQuestion = data.root[0]['quiz.question'];
			quizVm.newQuiz.newQuestions = [];

			var timeObj = editQuizVm.getTimeObj(quizVm.newQuiz.duration);

			var duration =  Duration.parse(quizVm.newQuiz.duration)
			var seconds = duration.seconds();
			quizVm.newQuiz.hours = parseInt( seconds / 3600 ) % 24;
			quizVm.newQuiz.minutes = parseInt( seconds / 60 ) % 60;
			quizVm.newQuiz.seconds = parseInt(seconds) % 60;
		}, function(err){
			console.log(err);
		});

		function editQuiz() {
			quizVm.newQuiz.questions = angular.copy(quizVm.newQuiz['quiz.question']);
			areInValidateInput = quizVm.validateInput(quizVm.newQuiz);
			if(areInValidateInput) {
				SNACKBAR({
					message: areInValidateInput,
					messageType: "error",
				})
				return
			}

			var newQues = quizVm.newQuiz.newQuestions;
			if(newQues) {
				for(var i =0; i < newQues.length; i++) {
					quizVm.newQuiz.questions.push({
						_uid_: newQues[i]._uid_,
						text: newQues[i].text,
					});
				}
			}

			quizVm.newQuiz.duration = (quizVm.newQuiz.hours || 0) + "h" + (quizVm.newQuiz.minutes || 0) + "m" + (quizVm.newQuiz.seconds || 0) + "s";

			// API CALL
			quizService.editQuiz(quizVm.newQuiz)
			.then(function(data){
				SNACKBAR({
					message: data.Message,
					messageType: "error",
				});
				quizVm.newQuiz = {};
				$state.transitionTo("quiz.all");
			}, function(err){
				console.log(err);
			})

		}

		function addNewQuestion(question, index) {
			var questionLength = editQuizVm.selectedQuestion.length;

			if(question.is_checked) {
				for(var i = 0; i< questionLength; i++) {
					var currentQues = editQuizVm.selectedQuestion[i];
					if(currentQues._uid_ == question._uid_) {
						if(currentQues.is_delete === true) {
							currentQues.is_delete = false;
							return;
						} else {
							SNACKBAR({
								message: "Already selected, uncheck to remove it",
								messageType: "error",
							});
							return;
						}
					}
				}

				question.index = index;
				quizVm.newQuiz.newQuestions.push(question);
			} else {
				for(var i = 0; i< questionLength; i++) {
					var currentQues = editQuizVm.selectedQuestion[i];
					if(currentQues._uid_ == question._uid_) {
						if(!currentQues.is_delete) {
							currentQues.is_delete = true;
							return;
						}
					}
				}
				var idx = mainVm.indexOfObject(quizVm.newQuiz.newQuestions, question);
				if (idx >= 0) {
					quizVm.newQuiz.newQuestions.splice(idx, 1)
				}
			}
		}

		function onQuestionRemove(question) {
			if(question._uid_) {
				question.is_delete = true;
			}
			console.log(question);
		}

		function onNewQuestionRemove(question) {
			mainVm.allQuestions[question.index].is_checked = false;

			var idx = mainVm.indexOfObject(quizVm.newQuiz.newQuestions, question);

			if (idx >= 0) {
				quizVm.newQuiz.newQuestions.splice(idx, 1)
			}
		}

		editQuizVm.isSelected = function(question_id) {
			for(var i = 0; i<editQuizVm.selectedQuestion.length; i++) {
				if(editQuizVm.selectedQuestion[i]._uid_ == question_id) {
					return true;
				}
			}
			return false;
		}

		function getTimeObj(duration){
			console.log(duration);
			var timeArr = duration.split("-") 
			return {
				hours: parseInt(timeArr[0]),
				minutes: parseInt(timeArr[1]),
				seconds: parseInt(timeArr[2]),
			}
		}
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