(function(){

	function inviteController($scope, $rootScope, $stateParams, $state, quizService, inviteService) {
		inviteVm = this;

		inviteVm.newInvite = {};

		// FUNCTION DECLARATION
		inviteVm.inviteCandidate = inviteCandidate;
		inviteVm.removeSelectedQuiz = removeSelectedQuiz;

		quizService.getAllQuizes().then(function(data){
			var data = JSON.parse(data);
			inviteVm.allQuizes = data.debug[0].quiz;
		}, function(err){
			console.log(err);
		})

		// FUNCTION DEFINITION

		function inviteCandidate() {
			if(!inviteVm.newInvite.name) {
				SNACKBAR({
					message: "Please Enter Valid Name",
					messageType: "error",
				})
				return
			}
			if(!isValidEmail(inviteVm.newInvite.email)) {
				SNACKBAR({
					message: "Please Enter Valid Email",
					messageType: "error",
				})
				return
			}
			if(!inviteVm.newInvite.dates) {
				SNACKBAR({
					message: "Please Enter Valid Date",
					messageType: "error",
				})
				return
			}

			var dateTime = formatDate(inviteVm.newInvite.dates);
			inviteVm.newInvite.quiz_id = inviteVm.newInvite.quiz._uid_;
			inviteVm.newInvite.validity = dateTime;
			console.log(inviteVm.newInvite);
			inviteService.inviteCandidate(inviteVm.newInvite).then(function(data){
				console.log(data);
				SNACKBAR({
					message: data.Message,
					messageType: "success",
				});
				if(data.Success) {
					$state.transitionTo("invite.dashboard", {
						quizID: inviteVm.newInvite.quiz_id,
					})
					inviteVm.newInvite = {}
				}
			}, function(err){
				console.log(err)
			});
		}

		function removeSelectedQuiz(){
			delete inviteVm.newInvite.quiz;
		}
		$(document).ready(function(){
			$('#datePicker').val(new Date().toDateInputValue());
		})
	}

	function editInviteController($rootScope, $stateParams, $state, quizService, inviteService) {
		editInviteVm = this;
		var candidateUID = $stateParams.candidateID;
		var quizID = $stateParams.quizID;

		//Function Declation
		editInviteVm.editInvite = editInvite;

		if(!candidateUID) {
			SNACKBAR({
				message: "Not a valid candidate",
				messageType: "error",
			})
			$state.transitionTo("invite.add");
		}

		inviteService.getCandidate(candidateUID)
		.then(function(data){
			editInviteVm.candidateBak = data['quiz.candidate'][0];
			editInviteVm.candidate = angular.copy(editInviteVm.candidateBak);

			editInviteVm.candidate.dates = new Date(getDate(editInviteVm.candidate.validity));
		}, function(err) {
			console.log(err)
		});

		function editInvite() {
			editInviteVm.candidate.id = candidateUID;
			editInviteVm.candidate.quiz_id = quizID;
			editInviteVm.candidate.old_quiz_id = quizID;
			editInviteVm.candidate.validity = formatDate(editInviteVm.candidate.dates);

			requestData = angular.copy(editInviteVm.candidate);

			console.log(requestData);
			inviteService.editInvite(editInviteVm.candidate)
			.then(function(data){
				SNACKBAR({
					message: data.Message,
					messageType: "success",
				})
				$state.transitionTo("invite.dashboard", {
					quizID:  requestData.quiz_id,
				})
			}, function(err){
				console.log(err)
			})
		}
	}
	
	function candidatesController($rootScope, $stateParams, $state, inviteService) {
			candidatesVm = this;

			candidatesVm.quizID = $stateParams.quizID;

			if(!candidatesVm.quizID) {
				SNACKBAR({
					message: "Not a valid Quiz",
					messageType: "error",
				});
				$state.transitionTo("invite.add");
			}
			console.log(candidatesVm.quizID);
			inviteService.getInvitedCandidates(candidatesVm.quizID).then(function(data){
				candidatesVm.quizCandidates = data.quiz[0]["quiz.candidate"];

				if(!candidatesVm.quizCandidates) {
					SNACKBAR({
						message: "Invite Candidate first to see all candidate",
						messageType: "error",
					});
					$state.transitionTo("invite.add");
				}
			}, function(err){
				console.log(err);
			});
		}


	var candidatesDependency = [
	    "$rootScope",
	    "$stateParams",
	    "$state",
	    "inviteService",
	    candidatesController
	];
	angular.module('GruiApp').controller('candidatesController', candidatesDependency);

	var editInviteDependency = [
	    "$rootScope",
	    "$stateParams",
	    "$state",
	    "quizService",
	    "inviteService",
	    editInviteController
	];
	angular.module('GruiApp').controller('editInviteController', editInviteDependency);

	var inviteDependency = [
	    "$scope",
	    "$rootScope",
	    "$stateParams",
	    "$state",
	    "quizService",
	    "inviteService",
	    inviteController
	];
	angular.module('GruiApp').controller('inviteController', inviteDependency);

})();