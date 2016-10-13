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
		cqVm.timerObj = {};

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
				if(!cqVm.question) {
					// Initialize timer if first time api call.
					cqVm.getTime();
				}
				startTimer(0, document.querySelector('#time-taken'), true);

				cqVm.question = data;
				if(data._uid_ == "END") {
					cqVm.quizEnded = true;
					cqVm.total_score = data.score;

					clearAllTimers();
				}

				if(cqVm.question.multiple == "true") {
					cqVm.answer = {};
				}

				$rootScope.updgradeMDL();
			}, function(err){	
				console.log(err);
				clearAllTimers();
				cqVm.apiError = true;
				SNACKBAR({
					message: "Something Went Wrong",
					messageType: "error",
				})
			})
		}

		function clearAllTimers() {
			clearInterval(cqVm.timerObj.time_taken);
			clearInterval(cqVm.timerObj.time_left);
			clearInterval(cqVm.timerObj.getTime);
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

		function manipulateTime (timer, display) {
      hours = Math.floor(timer / 3600);
      minutes = parseInt((timer / 60) % 60, 10);
      seconds = parseInt(timer % 60, 10);

      console.log(hours, minutes, seconds);
      hours = hours < 10 ? "0" + hours : hours;
      minutes = minutes < 10 ? "0" + minutes : minutes;
      seconds = seconds < 10 ? "0" + seconds : seconds;

      display.textContent = hours + ":"+ minutes + ":" + seconds;
		}

	  function startTimer(duration, display, isReverse) {
      var timer = duration, hours, minutes, seconds;

      if(isReverse) {
      	clearInterval(cqVm.timerObj.time_taken);
      	cqVm.timerObj.time_taken = setInterval(function () {
      		manipulateTime(timer, display);
          timer++;
      	}, 1000);
      } else {
				cqVm.timerObj.time_left = setInterval(function () {
      		manipulateTime(timer, display);
        	if (--timer < 0) {
            stopTimer();
        	}
      	}, 1000);
      }

      if(cqVm.quizEnded) {
				clearAllTimers();
			}
    }

    function stopTimer() {
      clearInterval(cqVm.timerObj.time_left);
    }

    function initTimer(totalTime) {
      var duration = Duration.parse(totalTime)
      console.log(duration);
      display = document.querySelector('#time');

      stopTimer(); //Reset Time left interval
      startTimer(duration.seconds(), display);
    };


		function getTime() {
			// Hit the PING api
			candidateService.getTime()
			.then(function(data){
				cqVm.initTimer(data.time_left)
			}, function(err){
				console.log(err);
			})
		}

		cqVm.timerObj.getTime = setInterval(function(){
			cqVm.getTime();
		}, 5000);

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