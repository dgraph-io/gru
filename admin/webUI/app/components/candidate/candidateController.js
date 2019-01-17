angular.module("GruiApp").controller("candidateController", [
  "$state",
  function candidateController($state) {
    candidateVm = this;
    mainVm.pageName = "candidate-page";

    candidateVm.isValidUser = candidateVm.checkValidity();

    // Check if user is authorized
    candidateVm.checkValidity = function checkValidity() {
      var candToken = localStorage.getItem("candidate_info");
      // If the candidate directly came to /quiz url to resume the quiz, then token might be null
      // in incognito window. So we show him an error.
      if (candToken == null) {
        mainVm.errorMessage = "Please use the link we emailed you to resume the quiz.";
        return false;
      }
      var ctoken = JSON.parse(candToken);

      if (ctoken.token) {
        candidateVm.info = ctoken;
        return true;
      } else {
        return false;
      }
    }
  }
]);

angular.module("GruiApp").controller("candidateQuizController", [
  "$scope",
  "$rootScope",
  "$state",
  "$interval",
  "candidateService",
  function candidateQuizController(
    $scope,
    $rootScope,
    $state,
    $interval,
    candidateService
  ) {
    cqVm = this;
    cqVm.total_score = 0;
    candidateVm.isValidUser = candidateVm.checkValidity();
    mainVm.pageName = "candidate-quiz-page";
    cqVm.timerObj = {};

    cqVm.getQuestion = getQuestion;
    cqVm.submitAnswer = submitAnswer;
    cqVm.getTime = getTime;
    cqVm.initTimer = initTimer;
    cqVm.calcTimeTaken = calcTimeTaken;
    cqVm.stopQuiz = stopQuiz;
    cqVm.submitFeedback = submitFeedback;

    if (candidateVm.isValidUser) {
      cqVm.getQuestion();
    }

    $rootScope.$on("endQuiz", function(e, data) {
      cqVm.stopQuiz();
      cqVm.apiError = data.message;
      mainVm.hideNotification(true);
    });

    function getQuestion() {
      candidateService.getQuestion().then(
        function(data) {
          cqVm.question = data;

          if (data.uid == "END") {
            //END QUIZ
            cqVm.stopQuiz();
            cqVm.total_score = data.score;
          } else {
            var seconds = Duration.parse(data.time_taken).seconds();
            $timeTakenElem = document.querySelector("#time-taken");
            // So that we call getTime() when question is fetched the first time and update
            // time.
            cqVm.getTime();
            $timeTakenElem.textContent = "00:00";
            cqVm.timerObj.time_elapsed = 0;
            startTimer(seconds, $timeTakenElem, true);

            if (cqVm.question.multiple == "true") {
              cqVm.answer = {};
            }
          }

          setTimeout(
            function() {
              componentHandler.upgradeAllRegistered();
            },
            10
          );

          scrollTo($(".candidate"));
        },
        function(err) {
          if (err.status != 0) {
            if (err.data.Message) {
              cqVm.apiError = err.data.Message;
              SNACKBAR({
                message: err.data.Message,
                messageType: "error"
              });
            }
          }
        }
      );
    }

    function stopQuiz() {
      cqVm.quizEnded = true;
      clearAllTimers();
      cqVm.calcTimeTaken();
    }

    function clearAllTimers() {
      $interval.cancel(cqVm.timerObj.time_taken);
      $interval.cancel(cqVm.timerObj.time_left);
      $interval.cancel(cqVm.timerObj.getTime);
    }

    function submitAnswer(skip) {
      cqVm.retry = true;
      if (
        !skip &&
        (!cqVm.answer || angular.equals({}, cqVm.answer) || cqVm.answer == "")
      ) {
        SNACKBAR({
          message: "Please select answer or Skip the question",
          messageType: "error"
        });
        return;
      }
      var requestData = {
        qid: cqVm.question.uid,
        cuid: cqVm.question.cuid
      };

      if (!skip) {
        // If multiple Answer
        if (mainVm.isObject(cqVm.answer)) {
          requestData.aid = "";
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
      candidateService.submitAnswer(requestData).then(
        function(data) {
          cqVm.retry = false;
          cqVm.answer = "";
          if (data.status == 200) {
            cqVm.getQuestion();
          } else {
            clearAllTimers();
          }
        },
        function(err) {
          cqVm.clearUnwatch = $scope.$watch(
            angular.bind(mainVm, function() {
              return mainVm.showNotification;
            }),
            function(newValue, oldValue) {
              //If showNotification is false, it means server connecte, and RETRY is true
              if (newValue == false && cqVm.retry) {
                cqVm.submitAnswer(skip);
              }
            }
          );
          if (err.status == 400) {
          }
        }
      );
    }

    function manipulateTime(timer, display, isTimeLeft) {
      hours = Math.floor(timer / 3600);
      minutes = hours * 60 + parseInt(timer / 60 % 60, 10);
      seconds = parseInt(timer % 60, 10);

      if (isTimeLeft) {
        cqVm.finalTimeLeft = {
          minutes: minutes,
          seconds: seconds
        };
      }

      minutes = minutes < 10 ? "0" + minutes : minutes;
      seconds = seconds < 10 ? "0" + seconds : seconds;

      display.textContent = minutes + ":" + seconds;
    }

    function startTimer(duration, display, isReverse) {
      var timer = duration, hours, minutes, seconds;

      if (isReverse) {
        $interval.cancel(cqVm.timerObj.time_taken);
        cqVm.timerObj.time_taken = $interval(
          function foo() {
            manipulateTime(timer, display);
            cqVm.timerObj.time_elapsed = timer++;
          },
          1000
        );
      } else {
        cqVm.timerObj.time_left = $interval(
          function() {
            manipulateTime(timer, display, true);
            if (--timer < 0) {
              stopTimer();
            }
          },
          1000
        );
      }

      if (cqVm.quizEnded) {
        clearAllTimers();
      }
    }

    function stopTimer() {
      $interval.cancel(cqVm.timerObj.time_left);
    }

    function initTimer(totalTime) {
      var duration = Duration.parse(totalTime);
      display = document.querySelector("#time");

      stopTimer(); //Reset Time left interval
      if (display) {
        startTimer(duration.seconds(), display);
      }
    }

    function getTime() {
      // Hit the PING api
      candidateService.getTime().then(
        function(data) {
          mainVm.consecutiveError = 0;
          if (data.time_left != "-1") {
            if (mainVm.showNotification) {
              mainVm.hideNotification();
            }
            isPositve = Duration.parse(data.time_left)._nanoseconds > 0;
            if (isPositve) {
              cqVm.initTimer(data.time_left);
            } else {
              cqVm.finalTimeLeft = {
                minutes: 0,
                seconds: 0
              };
              cqVm.stopQuiz();
            }
          }
        },
        function(err) {
          var message = err.status == 0
            ? "You internet seems to be offline, try refreshing the page after its back up..."
            : false;
          mainVm.initNotification(message);
        }
      );
    }

    cqVm.timerObj.getTime = $interval(
      function() {
        cqVm.getTime();
      },
      3000
    );

    function calcTimeTaken() {
      if (!cqVm.finalTimeLeft) {
        return;
      }
      var quizTime = JSON.parse(
        localStorage.getItem("candidate_info")
      ).duration;
      var minutes = quizTime.minutes - cqVm.finalTimeLeft.minutes;
      var seconds = quizTime.seconds - cqVm.finalTimeLeft.seconds;

      timeTakenSec = minutes * 60 + seconds;

      var timeTaken = {
        minutes: parseInt(timeTakenSec / 60 % 60, 10),
        seconds: parseInt(timeTakenSec % 60, 10)
      };

      // Adding prefix
      minutes = timeTaken.minutes < 10
        ? "0" + timeTaken.minutes
        : timeTaken.minutes;
      seconds = timeTaken.seconds < 10
        ? "0" + timeTaken.seconds
        : timeTaken.seconds;

      var text = minutes + ":" + seconds;
      $("#time-taken").text(text);

      if (cqVm.finalTimeLeft.minutes + cqVm.finalTimeLeft.seconds == 0) {
        var elem = document.querySelector("#time");
        if (elem) {
          manipulateTime(0, elem);
        }
      }
    }

    function submitFeedback() {
      if (!cqVm.feedback) {
        SNACKBAR({
          message: "Please write your feedback",
          messageType: "error"
        });
        return;
      }

      candidateService.sendFeedback({
        feedback: cqVm.feedback
      }).then(
        function(data) {
          cqVm.feedbackSubmitted = true;
        });
    }

    $scope.$on("$destroy", function() {
      if (cqVm.timerObj.getTime) {
        $interval.cancel(cqVm.timerObj.getTime);
      }
    });
  }
]);
