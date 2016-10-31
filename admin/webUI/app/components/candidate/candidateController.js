(function(){

	function candidateController($state) {

	// VARIABLE DECLARATION
		candidateVm = this;
		mainVm.pageName = "candidate-page";

	// FUNCTION DECLARATION
		candidateVm.checkValidity = checkValidity;

	// INITIALIZER
		candidateVm.isValidUser = candidateVm.checkValidity();

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

	// FUNCTION DEFINITION

		// Check if user is authorized
		function checkValidity() {
			var ctoken = JSON.parse(localStorage.getItem("candidate_info"));

			if(ctoken.token) {
				candidateVm.info = ctoken;
				return true
			} else {
				return false
			}
		}
	}

	function candidateQuizController($scope, $rootScope, $state, candidateService) {
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
		cqVm.stopQuiz = stopQuiz;
		cqVm.submitFeedback = submitFeedback;

	// INITIALIZERS
		if(candidateVm.isValidUser) {
			cqVm.getQuestion();
		}

		$rootScope.$on("endQuiz", function(e, data){
			cqVm.stopQuiz();
			cqVm.apiError = data.message;
			mainVm.hideNotification(true);
		})

	// FUNCTION DEFINITION

		// Get Question
		function getQuestion() {

			candidateService.getQuestion()
			.then(function(data) {
				$timeTakenElem = document.querySelector('#time-taken');
				if(!cqVm.question) {
					// Initialize timer if first time api call.
					cqVm.getTime();
					startTimer(0, $timeTakenElem, true);
				} else {
					$timeTakenElem.textContent = "00:00:00";
					startTimer(1, $timeTakenElem, true);
				}

				cqVm.question = data;
				if(data._uid_ == "END") {
					cqVm.stopQuiz();
					cqVm.total_score = data.score;
				}

				if(cqVm.question.multiple == "true") {
					cqVm.answer = {};
				}

				setTimeout(function() {
					componentHandler.upgradeAllRegistered();
				}, 10);
			}, function(err){	
				if(err.status != 0) {
					if(err.data.Message) {
						cqVm.apiError = err.data.Message;
						SNACKBAR({
							message: err.data.Message,
							messageType: "error",
						})
					}
				}
			})
		}

		function stopQuiz() {
			cqVm.quizEnded = true;
			clearAllTimers();
			cqVm.calcTimeTaken();
		}

		function clearAllTimers() {
			clearInterval(cqVm.timerObj.time_taken);
			clearInterval(cqVm.timerObj.time_left);
			clearInterval(cqVm.timerObj.getTime);
		}

		function submitAnswer(skip){
			cqVm.retry = true;
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
				cqVm.retry = false;
				cqVm.answer = "";
				if(data.status == 200) {
					cqVm.getQuestion();
				} else {
					clearAllTimers();
				}
			}, function(err){
				mainVm.showAjaxLoader = true;
				cqVm.clearUnwatch = $scope.$watch(angular.bind(mainVm, function () {
			    return mainVm.showNotification;
			  }), function (newValue, oldValue) {
			  	//If showNotification is false, it means server connecte, and RETRY	is true
			  	if(newValue == false && cqVm.retry) { 
			  		cqVm.submitAnswer(skip);
			  	}
				});
				if(err.status == 400) {

				}
			})
		}

		function manipulateTime(timer, display, isTimeLeft) {
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
      display = document.querySelector('#time');

      stopTimer(); //Reset Time left interval
      if(display) {
      	startTimer(duration.seconds(), display);
      }
    };


		function getTime() {
			// Hit the PING api
			candidateService.getTime()
			.then(function(data){
				mainVm.consecutiveError = 0
				if(data.time_left != "-1"){
					if(mainVm.showNotification) {
						mainVm.hideNotification();
					}	
					isPositve = Duration.parse(data.time_left)._nanoseconds > 0;
					if(isPositve) {
						cqVm.initTimer(data.time_left)
					} else {
						cqVm.finalTimeLeft = {
							hours: 0,
							minutes: 0,
							seconds: 0,
						}
						cqVm.stopQuiz();
					}
				}
			}, function(err){
				mainVm.initNotification();
				// if(err.status == 0) {
				// 	mainVm.timeoutModal();
				// 	cqVm.stopQuiz();
				// }
			})
		}

		cqVm.timerObj.getTime = setInterval(function(){
			cqVm.getTime();
		}, 3000);

		function calcTimeTaken(){
			if(!cqVm.finalTimeLeft) {
				return
			}
			var quizTime = JSON.parse(localStorage.getItem("candidate_info")).duration;

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

      if(cqVm.finalTimeLeft.hours + cqVm.finalTimeLeft.minutes + cqVm.finalTimeLeft.seconds == 0) {
      	var elem =  document.querySelector('#time');
      	if(elem) {
      		manipulateTime(0, elem)
      	}
      }
		}

		function submitFeedback() {
			if(!cqVm.feedback) {
				SNACKBAR({
					message: "Please write your feedback",
					messageType: "error",
				})
				return
			}

			var requestData = {
				feedback: escape(cqVm.feedback),
			};
			candidateService.sendFeedback(requestData)
			.then(function(data){
				console.log(data);
				cqVm.feedbackSubmitted = true;
			}, function(err){
				console.log(err);
			});
		}

	}

	// CANDIDATE QUIZ
	var candidateQuizDependency = [
			"$scope",
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