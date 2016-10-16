(function(){

	function inviteController($scope, $rootScope, $stateParams, $state, quizService, inviteService) {
		inviteVm = this;

		inviteVm.newInvite = {};
		mainVm.pageName = "invite-page"

		// FUNCTION DECLARATION
		inviteVm.inviteCandidate = inviteCandidate;
		inviteVm.removeSelectedQuiz = removeSelectedQuiz;
		inviteVm.setMinDate = setMinDate;

		quizService.getAllQuizes().then(function(data){
			var data = JSON.parse(data);
			inviteVm.allQuizes = data.debug[0].quiz;
		}, function(err){
			console.log(err);
		})

		function setMinDate() {
			setTimeout(function() {
				$("#datePicker").attr("min", formatDate(new Date()));
			}, 100);
		}

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
			inviteService.inviteCandidate(inviteVm.newInvite).then(function(data){
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
		editInviteVm.initAllQuiz = initAllQuiz;
		editInviteVm.selectedQuiz = selectedQuiz;
		editInviteVm.removeSelectedQuiz = removeSelectedQuiz;
		editInviteVm.onQuizSelect = onQuizSelect;

		inviteVm.setMinDate();

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

			editInviteVm.initAllQuiz();
		}, function(err) {
			console.log(err)
		});

		function editInvite() {
			editInviteVm.candidate.id = candidateUID;
			editInviteVm.candidate.quiz_id = "";
			editInviteVm.candidate.old_quiz_id = "";
			editInviteVm.candidate.validity = formatDate(editInviteVm.candidate.dates);

			if(editInviteVm.candidate['candidate.quiz'][0].is_delete) {
				editInviteVm.candidate.quiz_id = editInviteVm.candidate.quiz._uid_;
				editInviteVm.candidate.old_quiz_id = quizID;
			}

			requestData = angular.copy(editInviteVm.candidate);

			inviteService.editInvite(editInviteVm.candidate)
			.then(function(data){
				SNACKBAR({
					message: data.Message,
					messageType: "success",
				})
				$state.transitionTo("invite.dashboard", {
					quizID:  quizID,
				})
			}, function(err){
				console.log(err)
			})
		}

		function initAllQuiz() {
			setTimeout(function() {
				editInviteVm.allQuizes = angular.copy(inviteVm.allQuizes);
				$rootScope.updgradeMDL();
			}, 100);
		}

		function selectedQuiz(quiz) {
			var oldQuiz = editInviteVm.candidate['candidate.quiz'][0];
			isSelected = oldQuiz._uid_ == quiz._uid_;
			if(isSelected) {
				if(!oldQuiz.is_delete) {
					editInviteVm.candidate.quiz = quiz;
					return true;
				} else {
					return false;
				}
			} else {
				var currentQuiz = editInviteVm.candidate.quiz;
				if(currentQuiz && quiz._uid_ == currentQuiz._uid_) {
					return true;
				}
			}
		}

		function removeSelectedQuiz() {
			var oldQuiz = editInviteVm.candidate['candidate.quiz'][0];
			var isOld = oldQuiz._uid_ == editInviteVm.candidate.quiz._uid_;
			if(isOld){
				oldQuiz.is_delete = true;
				delete editInviteVm.candidate.quiz
			}
		}

		function onQuizSelect() {
			var quiz = editInviteVm.candidate.quiz;
			var oldQuiz = editInviteVm.candidate['candidate.quiz'][0];
			if(!quiz) {
				oldQuiz.is_delete = true;
			} else {
				if(oldQuiz._uid_ == quiz._uid_) {
					oldQuiz.is_delete = false;
				} else {
					oldQuiz.is_delete = true;
				}
			}
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

	function addCandidatesController($state) {
		inviteVm.setMinDate();
	}

	var candidatesDependency = [
	    "$rootScope",
	    "$stateParams",
	    "$state",
	    "inviteService",
	    candidatesController
	];
	angular.module('GruiApp').controller('candidatesController', candidatesDependency);

	var addCandidatesDependency = [
	    "$state",
	    addCandidatesController
	];
	angular.module('GruiApp').controller('addCandidatesController', addCandidatesDependency);

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