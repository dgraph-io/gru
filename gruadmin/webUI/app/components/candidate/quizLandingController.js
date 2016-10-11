(function(){

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

	// CANDIDATE QUIZ
	var quizLandingDependency = [
	    "$state",
	    "$stateParams",
	    "$q",
	    "$http",
	    quizLandingController
	];
	angular.module('GruiApp').controller('quizLandingController', quizLandingDependency);
})();