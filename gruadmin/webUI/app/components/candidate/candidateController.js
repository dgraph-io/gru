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
		cqVm.calcTimeTaken = calcTimeTaken;

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

					cqVm.calcTimeTaken();
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

		function manipulateTime (timer, display, isTimeLeft) {
      hours = Math.floor(timer / 3600);
      minutes = parseInt((timer / 60) % 60, 10);
      seconds = parseInt(timer % 60, 10);

      if(isTimeLeft) {
      	cqVm.finalTimeLeft = {
	      	hours: hours,
	      	minutes: minutes,
	      	seconds: seconds,
	      }
      }

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
      		manipulateTime(timer, display, true);
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

		function calcTimeTaken(){
			var quizTime = JSON.parse(localStorage.getItem("quiz_time"));

			var hours = quizTime.hours - cqVm.finalTimeLeft.hours;
			var minutes = quizTime.minutes - cqVm.finalTimeLeft.minutes;
			var seconds = quizTime.seconds - cqVm.finalTimeLeft.seconds;

			timeTakenSec = hours * 3600 + minutes * 60 + seconds;

			var timeTaken = {
        hours:  Math.floor(timeTakenSec / 3600),
        minutes: parseInt((timeTakenSec / 60) % 60, 10),
        seconds: parseInt(timeTakenSec % 60, 10),
      }

      // Adding prefix
      hours = timeTaken.hours < 10 ? "0" + timeTaken.hours : timeTaken.hours;
      minutes = timeTaken.minutes < 10 ? "0" + timeTaken.minutes : timeTaken.minutes;
      seconds = timeTaken.seconds < 10 ? "0" + timeTaken.seconds : timeTaken.seconds;

      var text = hours + ":"+ minutes + ":" + seconds; 
      $("#time-taken").text(text);
		}

	}

	function quizLandingController($state, $stateParams, $q, $http) {

	// VARIABLE DECLARATION
		qlVm = this;
		qlVm.invalidUser = false;
		mainVm.pageName = "quiz-landing";

		if(!$stateParams.quiz_token) {
			console.log("Not a valid CANDIDATE");
		}

	// FUNCTION DECLARATION
		qlVm.validateQuiz = validateQuiz;

	// FUNCTION DEFINITION
		qlVm.validateQuiz();

		// Check if user is authorized
		function validateQuiz() {
			var req = {
				method: 'POST',
        url: mainVm.candidate_url + "/validate/" + $stateParams.quiz_token,
			}

			$http(req)
      .then(function(data) {
    
      		var token = data.data.token;

      		if(token) {
      			localStorage.setItem('candidate_token', token);
      			$state.transitionTo("candidate.landing");

      			qlVm.time = mainVm.parseGoTime(data.data.duration);
      		} else {
      			qlVm.invalidUser = true;
      		}
        },
        function(response, code) {
      		qlVm.invalidUser = true;
        }
      );
		}
	}

	var quizLandingDependency = [
	    "$state",
	    "$stateParams",
	    "$q",
	    "$http",
	    quizLandingController
	];
	angular.module('GruiApp').controller('quizLandingController', quizLandingDependency);


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