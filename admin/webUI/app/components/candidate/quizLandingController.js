(function(){

	function quizLandingController($scope, $state, $stateParams, $q, $http, $interpolate) {

	// VARIABLE DECLARATION
		qlVm = this;
		qlVm.invalidUser = false;
		mainVm.pageName = "quiz-landing";

		if(!$stateParams.quiz_token) {
			console.log("Not a valid CANDIDATE");
			qlVm.invalidUser = true
		} else {
			localStorage.setItem("quiz_token", $stateParams.quiz_token);
		}

	// FUNCTION DECLARATION
		qlVm.validateQuiz = validateQuiz;
		qlVm.checkedInfo = checkedInfo;

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
      		if(data.data.quiz_started) {
      			$state.transitionTo("candidate.quiz");
      		} else {
      			if(token) {
	      			// $state.transitionTo("candidate.landing");
	      			qlVm.validated = true;
	      			$scope.time = mainVm.parseGoTime(data.data.duration);
							data.data.duration = $scope.time;
	      			localStorage.setItem('candidate_info', JSON.stringify(data.data));
	      			initInstructions();
	      		} else {
	      			qlVm.invalidUser = true;
	      		}
      		}
        },
        function(response, code) {
      		qlVm.invalidUser = true;
      		if(response.data) {
      			mainVm.errorMessage = response.data.Message;
      		} else {
      			mainVm.errorMessage = "Something went wrong, mail us on contact@dgraph.io"
      		}
        }
      );
		}

		function initInstructions() {
			qlVm.info = {
				General: [
					"By taking this quiz, you agree not to discuss/post the questions shown here.",
					$interpolate("The duration of the quiz is <span class='bold text-red'> \
						<span ng-if='time.hours > 0'>{{time.hours}} hours, </span> \
						<span ng-if='time.minutes > 0'>{{time.minutes}} minutes, </span> \
							<span ng-if='time.seconds > 0'>{{time.seconds}} seconds</span> \
						</span>. Timing would be clearly shown.")($scope),
					"Once you start the quiz, the timer would not stop, irrespective of any client side issues.",
					"Once you start the quiz, do not refresh the page or else the current question would be skipped.",
					"Questions can have single or multiple correct answers. They will be shown accordingly.",
					"Your total score and the time left at any point in the quiz would be displayed on the top.",
				],
				Score: [
					"There is NEGATIVE scoring for wrong answers. So, please DO NOT GUESS.",
					"If you skip a question, the score awarded is always ZERO.",
					"If you skip a question, you can't go back and answer it again.",
					"Scoring for a question would be clearly marked on the right hand side box.",
				],
				Contact: [
					"If there are any problems or something is unclear, please DO NOT start the quiz.",
					"Send email to contact@dgraph.io and tell us the problem. So we can solve it before you take the quiz.",
				],
			}
		}

		function checkedInfo() {
			var checkedInput = $(".quiz-landing .mdl-checkbox__input:checked").length;
			var totalInput = qlVm.info.General.length + qlVm.info.Score.length + qlVm.info.Contact.length;

			return (checkedInput == totalInput) ? false : true;
		}
	}

	// CANDIDATE QUIZ
	var quizLandingDependency = [
		"$scope",
	    "$state",
	    "$stateParams",
	    "$q",
	    "$http",
	    "$interpolate",
	    quizLandingController
	];
	angular.module('GruiApp').controller('quizLandingController', quizLandingDependency);
})();