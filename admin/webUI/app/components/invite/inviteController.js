(function() {

  function inviteController($scope, $rootScope, $stateParams, $state, quizService, inviteService) {
    inviteVm = this;

    inviteVm.newInvite = {};
    mainVm.pageName = "invite-page"

    // FUNCTION DECLARATION
    inviteVm.getAllQuizes = getAllQuizes;
    inviteVm.inviteCandidate = inviteCandidate;
    inviteVm.removeSelectedQuiz = removeSelectedQuiz;
    inviteVm.setMinDate = setMinDate;
    inviteVm.resetForm = resetForm;
    inviteVm.invalidateInput = invalidateInput;
    inviteVm.preSeleteQuiz = preSeleteQuiz;


    function getAllQuizes(quizID) {
      if (!inviteVm.allQuizes) {
        quizService.getAllQuizes().then(function(data) {
          inviteVm.allQuizes = data.debug[0].quiz;

          preSeleteQuiz(quizID);
        }, function(err) {
          console.log(err);
        })
      } else {
        preSeleteQuiz(quizID);
      }
    }

    function preSeleteQuiz(quizID) {
      if (quizID) {
        var qLen = inviteVm.allQuizes.length;
        for (var i = 0; i < qLen; i++) {
          if (inviteVm.allQuizes[i]._uid_ == quizID) {
            inviteVm.newInvite.quiz = inviteVm.allQuizes[i];
            break;
          }
        }
      }
    }

    function setMinDate() {
      setTimeout(function() {
        $datePicker = $("#datePicker")
        var today = new Date();
        $datePicker.attr("min", formatDate(new Date()));


        inviteVm.newInvite.dates = new Date(today.setDate(today.getDate() + 7));
        // $datePicker.val(formatDate(inviteVm.newInvite.dates));
      }, 100);
    }

    // FUNCTION DEFINITION

    function inviteCandidate() {
      var invalidateInput = inviteVm.invalidateInput(inviteVm.newInvite);

      if (invalidateInput) {
        SNACKBAR({
          message: invalidateInput,
          messageType: "error",
        })
        return
      }

      var dateTime = formatDate(inviteVm.newInvite.dates);
      inviteVm.newInvite.quiz_id = inviteVm.newInvite.quiz._uid_;
      inviteVm.newInvite.validity = dateTime;

      inviteService.alreadyInvited(inviteVm.newInvite.quiz_id, inviteVm.newInvite.email).then(function(invited) {
        if (invited) {
          SNACKBAR({
            message: "Candidate has already been invited.",
            messageType: "error",
          })
          return
        } else {
          inviteService.inviteCandidate(inviteVm.newInvite).then(function(data) {
            SNACKBAR({
              message: data.Message,
              messageType: "success",
            });
            if (data.Success) {
              $state.transitionTo("invite.dashboard", {
                quizID: inviteVm.newInvite.quiz_id,
              })
              inviteVm.newInvite = {}
            }
          }, function(err) {
            console.log(err)
          });
        }
      })
    }

    function invalidateInput(inputs) {
      if (!isValidEmail(inputs.email)) {
        return "Please Enter Valid Email";
      }
      if (!inputs.dates) {
        return "Please Enter Valid Date";
      }
      return false
    }

    function removeSelectedQuiz() {
      delete inviteVm.newInvite.quiz;
    }
    $(document).ready(function() {
      $('#datePicker').val(new Date().toDateInputValue());
    })

    function resetForm() {
      inviteVm.removeSelectedQuiz();
    }
  }

  function addCandidatesController($state, $stateParams) {
    acVm = this;
    var quizID = $state.params.quizID;

    inviteVm.setMinDate();
    inviteVm.getAllQuizes(quizID);
  }


  function editInviteController($rootScope, $stateParams, $state, quizService, inviteService) {
    editInviteVm = this;
    var candidateUID = $stateParams.candidateID;
    editInviteVm.quizID = $stateParams.quizID;

    //Function Declation
    editInviteVm.editInvite = editInvite;
    editInviteVm.initAllQuiz = initAllQuiz;
    editInviteVm.selectedQuiz = selectedQuiz;
    editInviteVm.onQuizChange = onQuizChange;
    editInviteVm.goToDashboard = goToDashboard;

    inviteVm.setMinDate();
    inviteVm.getAllQuizes();

    if (!candidateUID) {
      SNACKBAR({
        message: "Not a valid candidate",
        messageType: "error",
      })
      $state.transitionTo("invite.add");
    }

    inviteService.getCandidate(candidateUID)
      .then(function(data) {
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

      var invalidateInput = inviteVm.invalidateInput(editInviteVm.candidate);
      if (invalidateInput) {
        SNACKBAR({
          message: invalidateInput,
          messageType: "error",
        })
        return
      }

      if (editInviteVm.candidate['candidate.quiz'][0].is_delete) {
        editInviteVm.candidate.quiz_id = editInviteVm.candidate.quiz._uid_;
        editInviteVm.candidate.old_quiz_id = editInviteVm.quizID;
      }

      var requestData = angular.copy(editInviteVm.candidate);

      function update() {
        inviteService.editInvite(requestData)
          .then(function(data) {
            SNACKBAR({
              message: data.Message,
              messageType: "success",
            })
            $state.transitionTo("invite.dashboard", {
              quizID: editInviteVm.quizID,
            })
          }, function(err) {
            console.log(err)
          })
      }

      // If either the email or the quiz changes, we wan't to validate that the email
      // shouldn't be already invited to this quiz.
      if (editInviteVm.candidateBak.email != editInviteVm.candidate.email || editInviteVm.candidate.quiz._uid_ != editInviteVm.candidateBak["candidate.quiz"][0]._uid_) {
        inviteService.alreadyInvited(editInviteVm.candidate.quiz._uid_, editInviteVm.candidate.email).then(function(invited) {
            if (invited) {
              SNACKBAR({
                message: "Candidate has already been invited.",
                messageType: "error",
              })
              return
            } else {
              // Not invited yet, update.
              update()
            }
          })
          // Both email and quiz are same so maybe validity changed, we update.
      } else {
        update()
      }
    }

    function initAllQuiz() {
      setTimeout(function() {
        editInviteVm.allQuizes = angular.copy(inviteVm.allQuizes);
        $rootScope.updgradeMDL();
        editInviteVm.selectedQuiz()
      }, 100);
    }

    function selectedQuiz() {
      var oldQuiz = editInviteVm.candidate['candidate.quiz'][0]
      var quizLen = editInviteVm.allQuizes.length;
      for (var i = 0; i < quizLen; i++) {
        var quiz = editInviteVm.allQuizes[i];
        if (oldQuiz._uid_ == quiz._uid_) {
          editInviteVm.candidate.quiz = quiz;
          break;
        }
      }
    }

    function onQuizChange(item, model) {
      var oldQuiz = editInviteVm.candidate['candidate.quiz'][0];
      var isOld = oldQuiz._uid_ == model._uid_;

      oldQuiz.is_delete = isOld ? false : true;
    }

    function goToDashboard() {
      $state.transitionTo("invite.dashboard", {
        quizID: editInviteVm.quizID,
      });
    }
  }

  function candidatesController($rootScope, $stateParams, $state, inviteService) {
    candidatesVm = this;
    candidatesVm.sortType = 'score';
    candidatesVm.sortReverse = true;

    candidatesVm.quizID = $stateParams.quizID;

    if (!candidatesVm.quizID) {
      SNACKBAR({
        message: "Not a valid Quiz",
        messageType: "error",
      });
      $state.transitionTo("invite.add");
    }
    inviteService.getInvitedCandidates(candidatesVm.quizID).then(function(data) {
      candidatesVm.quizCandidates = data.quiz[0]["quiz.candidate"];

      if (!candidatesVm.quizCandidates) {
        SNACKBAR({
          message: "Invite Candidate first to see all candidate",
          messageType: "error",
        });
        $state.transitionTo("invite.add", {
          quizID: candidatesVm.quizID,
        });
      } else {
        for (var i = 0; i < candidatesVm.quizCandidates.length; i++) {
          var cand = candidatesVm.quizCandidates[i]
            // TODO - Maybe store invite in a format that frontend directly
            // understands.
          if (cand.complete == "false") {
            cand.invite_sent = new Date(Date.parse(cand.invite_sent)) || '';
            continue;
          }
          cand.quiz_start = new Date(Date.parse(cand.quiz_start)) || '';
          var score = 0.0;
          for (var j = 0; j < cand["candidate.question"].length; j++) {
            score += parseFloat(cand["candidate.question"][j]["candidate.score"]) || 0;
          }
          cand.score = score;
        }
      }
    }, function(err) {
      console.log(err);
    });
  }

  function candidateReportController($scope, $rootScope, $stateParams, $state, inviteService) {
    cReportVm = this;
    cReportVm.candidateID = $stateParams.candidateID;

    // Function
    cReportVm.initScoreCircle = initScoreCircle;
    cReportVm.isCorrect = isCorrect;

    if (!cReportVm.candidateID) {
      cReportVm.inValidID = true;
      return
    }

    inviteService.getReport(cReportVm.candidateID)
      .then(function(data) {
        console.log(data);
        for (var i = 0; i < data.questions.length; i++) {
          if (data.questions[i].time_taken != "-1") {
            data.questions[i].parsedTime = mainVm.parseGoTime(data.questions[i].time_taken)
          }
        }
        cReportVm.info = data;
        cReportVm.timeTaken = mainVm.parseGoTime(cReportVm.info.time_taken);
        cReportVm.info.feedback = unescape(cReportVm.info.feedback).replace(/\n/, "<br>");

        cReportVm.initScoreCircle();
      }, function(error) {
        console.log(error);
      }).then(function() {
        var correct = [];
        var skipped = [];
        var incorrect = [];
        var questions = cReportVm.info.questions;
        quesLen = questions.length

        for (var i = 0; i < questions.length; i++) {
          qn = questions[i]

          qn.answerArray = [];
          for (var j = 0; j < qn.answers.length; j++) {
            var answerObj = {
              _uid_: qn.answers[j]
            }
            answerObj.is_correct = (qn.correct.indexOf(qn.answers[j]) > -1)
            qn.answerArray.push(answerObj);
          }
          if ((qn.answers.length < qn.correct.length)) {
            qn.notAnswered = qn.correct.length - qn.answers.length;
          }

          if (qn.answers.length === qn.correct.length && angular.equals(qn.answers.sort(),
              qn.correct.sort())) {
            correct.push(qn)
          } else if (qn.score === 0 && qn.answers.length === 1) {
            skipped.push(qn)
            qn.isSkip = true
          } else {
            incorrect.push(qn)
          }
        }

        cReportVm.info.questions = [].concat([], incorrect, skipped, correct);

        setTimeout(function() {
          scrollNavInit();
        }, 0);
      });

    function initScoreCircle() {
      var circleWidth = 2 * Math.PI * 30;

      var percentage = (cReportVm.info.total_score * 100) / cReportVm.info.max_score;

      var circlePercentage = (circleWidth * percentage) / 100;

      var circleProgressWidth = circleWidth - circlePercentage;

      $progressBar = $(".prograss-circle");
      if (cReportVm.info.total_score != 0) {
        $progressBar.css({ 'display': 'block' });
        if (cReportVm.info.total_score < 0) {
          $progressBar.css({ 'stroke': 'red' });
        }
      }
      setTimeout(function() {
        $progressBar.css({ 'stroke-dashoffset': circleProgressWidth });
      }, 100);
    }

    function isCorrect(option, correct_options) {
      var uid = option._uid_;
      if (!correct_options) {
        return false
      }
      var optLength = correct_options.length;

      for (var i = 0; i < optLength; i++) {
        if (correct_options[i] == uid) {
          // option.is_correct = true
          return true
        }
      }
      return false
    }

    // var mdlContent = $(".mdl-layout__content");
    $(".mdl-layout__content").scroll(function() {
      if (this.scrollTop >= 100) {
        cReportVm.pageScrolled = true;
      } else {
        cReportVm.pageScrolled = false;
      }
      $scope.$digest();
    });
  }

  var candidateReportDependency = [
    "$scope",
    "$rootScope",
    "$stateParams",
    "$state",
    "inviteService",
    candidateReportController
  ];
  angular.module('GruiApp').controller('candidateReportController', candidateReportDependency);

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
