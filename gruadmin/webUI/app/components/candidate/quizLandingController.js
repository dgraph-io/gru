(function(){

	function quizLandingController($scope, $state, $stateParams, $q, $http, $interpolate) {

	// VARIABLE DECLARATION
		qlVm = this;
		qlVm.invalidUser = false;
		mainVm.pageName = "quiz-landing";

		if(!$stateParams.quiz_token) {
			console.log("Not a valid CANDIDATE");
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

      		if(token) {
      			localStorage.setItem('candidate_token', token);
      			// $state.transitionTo("candidate.landing");
      			qlVm.validated = true;
      			$scope.time = mainVm.parseGoTime(data.data.duration);
      			localStorage.setItem('quiz_time', JSON.stringify($scope.time));
      			initInstructions();
      		} else {
      			qlVm.invalidUser = true;
      		}
        },
        function(response, code) {
      		qlVm.invalidUser = true;
        }
      );
		}

		function initInstructions() {
			qlVm.info = {
				General: [
					"By taking this quiz, you agree not to discuss/post the questions shown here.",
					$interpolate("The duration of the quiz is <span class='bold text-red'> \
						<span ng-if='time.hours'>{{time.hours}} hours, </span> \
						<span ng-if='time.minutes'>{{time.minutes}} minutes, </span> \
							<span ng-if='time.seconds'>{{time.seconds}} seconds </span> \
						</span>. Timing would be clearly shown.")($scope),
					"Once you start the quiz, the timer would not stop, irrespective of any client side issues",
					"Questions can have single or multiple correct answers. They will be shown accordingly.",
					"Your total score and the time left at any point in the quiz would be displayed on the top.",
					"The scoring for each attempt of a question, would be visible to you in a separate section.",
				],
				Score: [
					"There is NEGATIVE scoring for wrong answers. So, please DO NOT GUESS.",
					"If you skip a question, the score awarded is always ZERO.",
					"Scoring would be clearly marked in the question on the right hand side box.",
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