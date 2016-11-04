(function(){

	function quizLandingController($scope, $state, $stateParams, $http, $interpolate, quizLandingService) {

	// VARIABLE DECLARATION
		qlVm = this;
		qlVm.candidate = {name: ""};
		qlVm.invalidUser = false;
		qlVm.saveName = saveName;
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
				$scope.time = data.data.duration;
				time = mainVm.parseGoTime(data.data.duration)
				$scope.time_minutes = time.hours * 60 + time.minutes
				data.data.duration = time;
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
						{{time_minutes}} minutes</span>. Timing would be clearly shown.")($scope),
					"We recommend using a pen and paper to help visualize some of the questions.",
					"Once you start the quiz, the timer would not stop, irrespective of any client side issues.",
					"Questions can have single or multiple correct answers. They will be shown accordingly.",
					"Your total score and the time left at any point in the quiz would be displayed on the top.",
				],
				Score: [
					"There is NEGATIVE scoring for wrong answers. So, please DO NOT GUESS.",
					"If you skip a question, the score awarded is always ZERO.",
					"If you skip a question, you can't go back and answer it again.",
					"Scoring for a question would be clearly marked on the right hand side box.",
					"Questions with multiple correct answers have partial scoring",
					"For questions with multiple correct answers, positive and negative score would indicate the score for each correct and wrong answer respectively.",
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

			return (checkedInput == totalInput) && qlVm.candidate.name != "" ? false : true;
		}

		function saveName() {
			var requestData = {
				name: qlVm.candidate.name,
			};

			var ctoken = JSON.parse(localStorage.getItem("candidate_info"));
			ctoken.Name = qlVm.candidate.name
			localStorage.setItem('candidate_info', JSON.stringify(ctoken));

			quizLandingService.addName(requestData).then(function(data){
				console.log(data);
				mainVm.goTo('candidate.quiz')
			}, function(err){
				console.log(err);
			});
		}
	}

	// CANDIDATE QUIZ
	var quizLandingDependency = [
		"$scope",
		"$state",
		"$stateParams",
		"$http",
		"$interpolate",
		"quizLandingService",
		quizLandingController
	];
	angular.module('GruiApp').controller('quizLandingController', quizLandingDependency);
})();