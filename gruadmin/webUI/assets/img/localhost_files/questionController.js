(function(){

	function questionController($scope, $rootScope, $http, $q, $state) {

	// VARIABLE DECLARATION
		questionVm = this;
		mainVm.pageName = "question"

	// FUNCTION DECLARATION
		questionVm.userAuthentication = userAuthentication;

	// FUNCTION DEFINITION

		// Check if user is authorized
		function userAuthentication(testId) {
		}
	}
	var questionDependency = [
	    "$scope",
	    "$rootScope",
	    "$http",
	    "$q",
	    "$state",
	    questionController
	];
	angular.module('GruiApp').controller('questionController', questionDependency);

})();