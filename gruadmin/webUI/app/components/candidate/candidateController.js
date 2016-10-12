(function(){

	function candidateController($state) {

	// VARIABLE DECLARATION
		candidateVm = this;
		mainVm.pageName = "candidate-page";

	// FUNCTION DECLARATION
		candidateVm.checkValidity = checkValidity;

	// INITIALIZER
		candidateVm.isValidUser = candidateVm.checkValidity();

	// FUNCTION DEFINITION

		// Check if user is authorized
		function checkValidity() {
			var ctoken = localStorage.getItem("candidate_token");

			if(ctoken) {
				return true
			} else {
				return false
			}
		}
	}

	function candidateQuizController($rootScope, $state, candidateService) {
		// VARIABLE DECLARATION
		cqVm = this;
		cqVm.total_score = 0
		candidateVm.isValidUser = candidateVm.checkValidity();
		mainVm.pageName = "candidate-quiz-page";

	// FUNCTION DECLARATION
		cqVm.getQuestion = getQuestion;
		cqVm.submitAnswer = submitAnswer;
		cqVm.getTime = getTime;
		cqVm.initTimer = initTimer;


	// INITIALIZERS
		cqVm.getQuestion();

	// FUNCTION DEFINITION

		// Get Question
		function getQuestion() {

			candidateService.getQuestion()
			.then(function(data) {
				cqVm.question = data;
				if(data._uid_ == "END") {
					cqVm.quizEnded = true;
					cqVm.total_score = data.score;
				}

				if(cqVm.question.multiple == "true") {
					cqVm.answer = {};
				}
				console.log(data);
				$rootScope.updgradeMDL();
			}, function(err){	
				console.log(err);
				SNACKBAR({
					message: "Something Went Wrong",
					messageType: "error",
				})
			})
		}

		function submitAnswer(skip){
			if(!skip && (!cqVm.answer || angular.equals({}, cqVm.answer) || cqVm.answer == "") ) {
				SNACKBAR({
					message: "Please select answer or Skip the question",
					messageType: "error",
				})
				return
			}
			var requestData = {
				qid: cqVm.question._uid_,
				cuid: cqVm.question.cuid,
			}	

			if(!skip) {
				// If multiple Answer
				if(mainVm.isObject(cqVm.answer)) {
					requestData.aid = ""
					for (var key in cqVm.answer) {
					  if (cqVm.answer.hasOwnProperty(key)) {
					    requestData.aid += key + ",";
					  }
					}
					requestData.aid = requestData.aid.slice(0, -1);
				} else {
					requestData.aid = cqVm.answer;
				}
			} else {
				requestData.aid = "skip";
			}
			candidateService.submitAnswer(requestData)
			.then(function(data){
				cqVm.answer = "";
				if(data.status == 200) {
					cqVm.getQuestion();
				}
			}, function(err){
				console.log(err);
			})
		}

		function getTime() {
			// Hit the PING api
			candidateService.getTime()
			.then(function(data){
				var time_left = data.time_left.split(".")[0]
				cqVm.initTimer(time_left)
			}, function(err){
				console.log(err);
			})
		}

		function initTimer(time_left) {
			console.log(time_left)
		}
	}


	// CANDIDATE QUIZ
	var candidateQuizDependency = [
			"$rootScope",
	    "$state",
	    "candidateService",
	    candidateQuizController
	];
	angular.module('GruiApp').controller('candidateQuizController', candidateQuizDependency);

	// MAIN CANDIDATE CONTROLLER
	var candidateDependency = [
	    "$state",
	    candidateController
	];
	angular.module('GruiApp').controller('candidateController', candidateDependency);

})();