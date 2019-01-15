(function() {
  function inviteController(
    $scope,
    $rootScope,
    $stateParams,
    $state,
    quizService,
    inviteService
  ) {
    inviteVm = this;

    inviteVm.newInvite = {};
    mainVm.pageName = "invite-page";

    // FUNCTION DECLARATION
    inviteVm.getAllQuizzes = getAllQuizzes;
    inviteVm.inviteCandidate = inviteCandidate;
    inviteVm.removeSelectedQuiz = removeSelectedQuiz;
    inviteVm.setMinDate = setMinDate;
    inviteVm.resetForm = resetForm;
    inviteVm.invalidateInput = invalidateInput;
    inviteVm.preSelectQuiz = preSelectQuiz;

    function getAllQuizzes(quizID) {
      if (!inviteVm.allQuizes) {
        quizService.getAllQuizzes().then(
          function(quizzes) {
            inviteVm.allQuizes = quizzes;
            preSelectQuiz(quizID);
          },
          function(err) {
            console.error(err);
          }
        );
      } else {
        preSelectQuiz(quizID);
      }
    }

    function preSelectQuiz(quizID) {
      if (!quizID) {
        return;
      }
      var qLen = inviteVm.allQuizes.length;
      var q = inviteVm.allQuizes.find(function(quizz) {
        return quizz.uid == quizID;
      })
      if (q) {
        inviteVm.newInvite.quiz = q;
      }
    }

    function setMinDate() {
      setTimeout(
        function() {
          $datePicker = $("#datePicker");
          var today = new Date();
          $datePicker.attr("min", formatDate(new Date()));

          inviteVm.newInvite.dates = new Date(
            today.setDate(today.getDate() + 7)
          );
        },
        100
      );
    }

    function inviteCandidate() {
      var invalidateInput = inviteVm.invalidateInput(inviteVm.newInvite);

      if (invalidateInput) {
        SNACKBAR({
          message: invalidateInput,
          messageType: "error"
        });
        return;
      }

      var dateTime = formatDate(inviteVm.newInvite.dates);
      inviteVm.newInvite.quiz_id = inviteVm.newInvite.quiz.uid;
      inviteVm.newInvite.validity = dateTime;

      inviteService
        .alreadyInvited(inviteVm.newInvite.quiz_id, inviteVm.newInvite.emails)
        .then(function(email) {
          if (email != "") {
            SNACKBAR({
              message: "Candidate with email " +
                email +
                " has already been invited.",
              messageType: "error"
            });
            return;
          } else {
            inviteService.inviteCandidate(inviteVm.newInvite).then(
              function(data) {
                SNACKBAR({
                  message: data.Message,
                  messageType: "success"
                });
                if (data.Success) {
                  $state.transitionTo("invite.dashboard", {
                    quizID: inviteVm.newInvite.quiz_id
                  });
                  inviteVm.newInvite = {};
                }
              },
              function(err) {
                console.error(err);
              }
            );
          }
        });
    }

    function invalidateInput(inputs) {
      for (var i = 0; i < inputs.emails.length; i++) {
        if (!isValidEmail(inputs.emails[i])) {
          return inputs.emails[i] + " isn't a valid email.";
        }
      }
      if (!inputs.dates) {
        return "Please Enter Valid Date";
      }
      return false;
    }

    function removeSelectedQuiz() {
      delete inviteVm.newInvite.quiz;
    }
    $(document).ready(function() {
      $("#datePicker").val(new Date().toDateInputValue());
    });

    function resetForm() {
      inviteVm.removeSelectedQuiz();
    }
  }

  function addCandidatesController($state, $stateParams) {
    acVm = this;
    var quizID = $state.params.quizID;

    inviteVm.setMinDate();
    inviteVm.getAllQuizzes(quizID);
  }

  function editInviteController(
    $rootScope,
    $stateParams,
    $state,
    quizService,
    inviteService
  ) {
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
    inviteVm.getAllQuizzes();

    if (!candidateUID) {
      SNACKBAR({
        message: "Not a valid candidate",
        messageType: "error"
      });
      $state.transitionTo("invite.add");
    }

    inviteService.getCandidate(candidateUID).then(
      function(data) {
        editInviteVm.candidateBak = data.data["quiz.candidate"][0];
        editInviteVm.candidate = angular.copy(editInviteVm.candidateBak);

        editInviteVm.candidate.dates = new Date(
          getDate(editInviteVm.candidate.validity)
        );

        editInviteVm.initAllQuiz();
      });

    function valid(input) {
      if (!isValidEmail(input.email)) {
        return input.email + " isn't a valid email.";
      }
      if (!input.dates) {
        return "Please Enter Valid Date";
      }
      return true;
    }

    function editInvite() {
      editInviteVm.candidate.id = candidateUID;
      editInviteVm.candidate.quiz_id = "";
      editInviteVm.candidate.old_quiz_id = "";
      editInviteVm.candidate.validity = formatDate(
        editInviteVm.candidate.dates
      );

      var validateInput = valid(editInviteVm.candidate);
      if (validateInput != true) {
        SNACKBAR({
          message: validateInput,
          messageType: "error"
        });
        return;
      }

      if (editInviteVm.candidate["candidate.quiz"][0].is_delete) {
        editInviteVm.candidate.quiz_id = editInviteVm.candidate.quiz.uid;
        editInviteVm.candidate.old_quiz_id = editInviteVm.quizID;
      }

      var requestData = angular.copy(editInviteVm.candidate);

      function update() {
        inviteService.editInvite(requestData).then(
          function(data) {
            SNACKBAR({
              message: data.Message,
              messageType: "success"
            });
            $state.transitionTo("invite.dashboard", {
              quizID: editInviteVm.quizID
            });
          },
          function(err) {
            console.error(err);
          }
        );
      }

      // If either the email or the quiz changes, we wan't to validate that the email
      // shouldn't be already invited to this quiz.
      if (
        editInviteVm.candidateBak.email != editInviteVm.candidate.email ||
        editInviteVm.candidate.quiz.uid !=
          editInviteVm.candidateBak["candidate.quiz"][0].uid
      ) {
        inviteService
          .alreadyInvited(editInviteVm.candidate.quiz.uid, [
            editInviteVm.candidate.email
          ])
          .then(function(email) {
            if (email != "") {
              SNACKBAR({
                message: "Candidate has already been invited.",
                messageType: "error"
              });
              return;
            } else {
              // Not invited yet, update.
              update();
            }
          });
        // Both email and quiz are same so maybe validity changed, we update.
      } else {
        update();
      }
    }

    function initAllQuiz() {
      setTimeout(
        function() {
          editInviteVm.allQuizes = angular.copy(inviteVm.allQuizes);
          $rootScope.upgradeMDL();
          editInviteVm.selectedQuiz();
        },
        100
      );
    }

    function selectedQuiz() {
      var oldQuiz = editInviteVm.candidate["candidate.quiz"][0];
      var quizLen = editInviteVm.allQuizes.length;
      for (var i = 0; i < quizLen; i++) {
        var quiz = editInviteVm.allQuizes[i];
        if (oldQuiz.uid == quiz.uid) {
          editInviteVm.candidate.quiz = quiz;
          break;
        }
      }
    }

    function onQuizChange(item, model) {
      var oldQuiz = editInviteVm.candidate["candidate.quiz"][0];
      var isOld = oldQuiz.uid == model.uid;

      oldQuiz.is_delete = isOld ? false : true;
    }

    function goToDashboard() {
      $state.transitionTo("invite.dashboard", {
        quizID: editInviteVm.quizID
      });
    }
  }

  function candidatesController(
    $scope,
    $rootScope,
    $stateParams,
    $state,
    $timeout,
    $templateCache,
    inviteService
  ) {
    candidatesVm = this;
    candidatesVm.sortType = "score";
    candidatesVm.sortReverse = true;

    candidatesVm.expires = expires;
    candidatesVm.showCancelModal = showCancelModal;
    candidatesVm.initiateCancel = initiateCancel;
    candidatesVm.showDeleteModal = showDeleteModal;
    candidatesVm.initiateDelete = initiateDelete;
    candidatesVm.deleteCandFromArray = deleteFromArray;
    candidatesVm.cancel = cancel;
    candidatesVm.resend = resend;
    candidatesVm.delete = deleteCand;
    candidatesVm.percentile = percentile;

    candidatesVm.quizID = $stateParams.quizID;

    if (!candidatesVm.quizID) {
      SNACKBAR({
        message: "Not a valid Quiz",
        messageType: "error"
      });
      $state.transitionTo("invite.add");
    }
    inviteService.getInvitedCandidates(candidatesVm.quizID).then(
      function(data) {
        var quizCandidates = data.data.quiz[0]["quiz.candidate"];

        if (!quizCandidates) {
          SNACKBAR({
            message: "Invite Candidate first to see all candidate",
            messageType: "error"
          });
          $state.transitionTo("invite.add", {
            quizID: candidatesVm.quizID
          });
        } else {
          completed = [];
          notCompleted = [];
          quizCandidates.forEach(function(candidate) {
            if (!candidate.complete) {
              candidate.invite_sent = new Date(
                Date.parse(candidate.invite_sent)
              );
              notCompleted.push(candidate);
            } else {
              candidate.quiz_start = new Date(
                Date.parse(candidate.quiz_start)
              );
              candidate.score = parseFloat(candidate.score) || 0.0;
              completed.push(candidate);
            }
          });

          completed.sort(function(c1, c2) {
            return c1.score - c2.score;
          });

          var lastScore = 0.0, lastIdx = 0, idx = 0, i = completed.length;
          while (i--) {
            var cand = completed[i];
            if (cand.score != lastScore) {
              cand.idx = idx;
              lastScore = cand.score;
              lastIdx = idx;
            } else {
              cand.idx = lastIdx;
            }
            idx++;
          }
          candidatesVm.completed = completed;
          candidatesVm.notCompleted = notCompleted;
          scrollToCandidate();
        }
      },
      function(err) {
        console.error(err);
      }
    );

    function showCancelModal(candidate) {
      // Timeout to let dirty checking done first then modal content get
      // updated variable text
      candidatesVm.currentCancel = {};
      candidatesVm.currentCancel = candidate;
      $timeout(
        function() {
          mainVm.openModal({
            template: "cancel-modal-template",
            showYes: true,
            hideClose: true,
            class: "cancel-invite-modal"
          });
        },
        10
      );
    }

    function initiateCancel() {
      if (candidatesVm.currentCancel) {
        candidatesVm.cancel(candidatesVm.currentCancel);
      }
    }

    function showDeleteModal(candidate) {
      candidatesVm.currentDeleteName = candidate.name;
      candidatesVm.currentDelete = candidate.uid;
      $timeout(
        function() {
          mainVm.openModal({
            template: "delete-candidate-template",
            showYes: true,
            hideClose: true,
            class: "delete-candidate-modal"
          });
        },
        10
      );
    }

    function initiateDelete() {
      if (candidatesVm.currentDelete) {
        candidatesVm.delete(candidatesVm.currentDelete);
      }
    }

    function expires(validity) {
      var validity_date = new Date(validity);
      var today = new Date();
      var diff = (validity_date - today) / (1000 * 60 * 60 * 24);
      var numDays = Math.round(diff);
      if (numDays <= 0) {
        return "Expired";
      }
      return numDays;
    }

    function deleteFromArray(candidateID, array) {
      var idx = -1;
      for (var i = 0; i < array.length; i++) {
        if (array[i].uid == candidateID) {
          idx = i;
          break;
        }
      }
      if (idx >= 0) {
        array.splice(idx, 1);
      }
    }

    function cancel(candidate) {
      inviteService
        .cancelInvite(candidate, candidatesVm.quizID)
        .then(function(cancelled) {
          if (!cancelled) {
            SNACKBAR({
              message: "Invite could not be cancelled.",
              messageType: "error",
            });
            return;
          }
          SNACKBAR({
            message: "Invite cancelled successfully.",
            messageType: "success",
          });
          deleteFromArray(candidate.uid, candidatesVm.notCompleted);
          $state.transitionTo("invite.dashboard", {
            quizID: candidatesVm.quizID
          });

          candidatesVm.currentCancel = {};
          mainVm.hideModal();
        });
    }

    function deleteCand(candidateId) {
      inviteService.deleteCand(candidateId).then(
        function(deleted) {
          if (!deleted) {
            SNACKBAR({
              message: "Candidate couldn't be deleted.",
              messageType: "error"
            });
            return;
          }
          SNACKBAR({
            message: "Candidate deleted successfully."
          });

          deleteFromArray(candidateId, candidatesVm.completed);
          $state.transitionTo("invite.dashboard", {
            quizID: candidatesVm.quizID
          });

          candidatesVm.currentDelete = "";
          mainVm.hideModal();
        },
        function(err) {
          console.error(error);
          candidatesVm.currentDelete = "";
          mainVm.hideModal();
        }
      );
    }

    function resend(candidate) {
      inviteService.resendInvite(candidate).then(function(response) {
        if (!response.success) {
          SNACKBAR({
            message: response.message,
            messageType: "error"
          });
          return;
        }
        SNACKBAR({
          message: response.message
        });
        $state.transitionTo("invite.dashboard", {
          quizID: candidatesVm.quizID
        });
      });
    }

    function scrollToCandidate() {
      // Scroll page to candidate if his/her report was viewed
      $timeout(
        function() {
          $candidateViewed = $(".report-viewed");
          if ($candidateViewed.length) {
            $(".mdl-layout__content").scrollTop(
              $candidateViewed.offset().top - 200
            );
          }
        },
        10
      );
    }

    function percentile(size, idx) {
      return (size - idx) / size * 100;
    }

    $(".mdl-layout__content").unbind("scroll");
  }

  function candidateReportController(
    $scope,
    $rootScope,
    $stateParams,
    $state,
    inviteService
  ) {
    cReportVm = this;
    cReportVm.candidateID = $stateParams.candidateID;
    cReportVm.idx = $stateParams.idx;
    cReportVm.total = $stateParams.total;
    cReportVm.resume = inviteService.getResume;
    inviteVm.reportViewed = cReportVm.candidateID;
    // Function
    cReportVm.initScoreCircle = initScoreCircle;
    cReportVm.isCorrect = isCorrect;

    if (!cReportVm.candidateID) {
      cReportVm.inValidID = true;
      return;
    }

    inviteService
      .getReport(cReportVm.candidateID)
      .then(
        function(data) {
          for (var i = 0; i < data.questions.length; i++) {
            if (data.questions[i].time_taken != "-1") {
              data.questions[i].parsedTime = mainVm.parseGoTime(
                data.questions[i].time_taken
              );
            }
          }
          cReportVm.info = data;
          cReportVm.timeTaken = mainVm.parseGoTime(cReportVm.info.time_taken);
          cReportVm.initScoreCircle();
        },
        function(error) {
          console.error(error);
        }
      )
      .then(function() {
        var questions = cReportVm.info.questions;
        var statistics = {
          easy: {
            correct: 0,
            total: 0
          },
          medium: {
            correct: 0,
            total: 0
          },
          hard: {
            correct: 0,
            total: 0
          }
        };

        for (var i = 0; i < questions.length; i++) {
          qn = questions[i];
          d = difficulty(qn.tags);
          if (d != "") {
            statistics[d].total++;
            correct(qn) && statistics[d].correct++;
          }
          qn.answerArray = [];
          for (var j = 0; j < qn.answers.length; j++) {
            var answerObj = {
              uid: qn.answers[j]
            };
            answerObj.is_correct = qn.correct.indexOf(qn.answers[j]) > -1;
            qn.answerArray.push(answerObj);
          }
          if (qn.answers.length < qn.correct.length) {
            qn.notAnswered = qn.correct.length - qn.answers.length;
          }

          if (qn.score === 0 && qn.answers.length === 1) {
            qn.isSkip = true;
          }
        }
        cReportVm.statistics = statistics;

        setTimeout(
          function() {
            scrollNavInit();
            bindHandlers();
          },
          0
        );
      });

    function difficulty(tags) {
      for (var i = 0; i < tags.length; i++) {
        if (tags[i] === "easy") {
          return "easy";
        } else if (tags[i] === "medium") {
          return "medium";
        } else if (tags[i] === "hard") {
          return "hard";
        }
      }
      return "";
    }

    function correct(question) {
      return angular.equals(question.correct.sort(), question.answers.sort());
    }

    function initScoreCircle() {
      var circleWidth = 2 * Math.PI * 30;

      var percentage = cReportVm.info.percentile;

      var circlePercentage = circleWidth * percentage / 100;

      var circleProgressWidth = circleWidth - circlePercentage;

      $progressBar = $(".prograss-circle");
      $progressBar.css({ display: "block" });

      setTimeout(
        function() {
          $progressBar.css({ "stroke-dashoffset": circleProgressWidth });
        },
        100
      );
    }

    // To scroll to next and previous question on pressing Up and Down arrow keys.
    function bindHandlers() {
      document.onkeydown = function(e) {
        switch (e.keyCode) {
          case 38:
            // Up arrow key
            selected = $(".selected-sidebar").data("scrollto");
            // Initially selected would be undefined, we don't want to do
            // anything on up key.
            if (selected === undefined) {
              break;
            }
            // Lets get the current selected question.
            qno = parseInt(selected.slice(9, 11));
            $previousQn = $("#question" + (qno - 1));
            if ($previousQn.length != 0) {
              $question = $previousQn;
            }
            $question.length != 0 && scrollTo($question);
            break;
          case 40:
            // Down arrow key
            selected = $(".selected-sidebar").data("scrollto");
            // Initially selected would be undefined, so lets select the first
            // question.
            if (selected === undefined) {
              $question = $("#question0");
            } else {
              qno = parseInt(selected.slice(9, 11));
              $nextQn = $("#question" + (qno + 1));
              if ($nextQn.length != 0) {
                $question = $nextQn;
              }
            }
            $question.length != 0 && scrollTo($question);
            break;
        }
      };
    }

    function isCorrect(option, correct_options) {
      var uid = option.uid;
      if (!correct_options) {
        return false;
      }
      var optLength = correct_options.length;

      for (var i = 0; i < optLength; i++) {
        if (correct_options[i] == uid) {
          return true;
        }
      }
      return false;
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
  angular
    .module("GruiApp")
    .controller("candidateReportController", candidateReportDependency);

  var candidatesDependency = [
    "$scope",
    "$rootScope",
    "$stateParams",
    "$state",
    "$timeout",
    "$templateCache",
    "inviteService",
    candidatesController
  ];
  angular
    .module("GruiApp")
    .controller("candidatesController", candidatesDependency);

  var addCandidatesDependency = ["$state", addCandidatesController];
  angular
    .module("GruiApp")
    .controller("addCandidatesController", addCandidatesDependency);

  var editInviteDependency = [
    "$rootScope",
    "$stateParams",
    "$state",
    "quizService",
    "inviteService",
    editInviteController
  ];
  angular
    .module("GruiApp")
    .controller("editInviteController", editInviteDependency);

  var inviteDependency = [
    "$scope",
    "$rootScope",
    "$stateParams",
    "$state",
    "quizService",
    "inviteService",
    inviteController
  ];
  angular.module("GruiApp").controller("inviteController", inviteDependency);
})();
