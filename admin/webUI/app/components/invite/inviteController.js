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

    inviteVm.getAllQuizzes = getAllQuizzes;
    inviteVm.inviteCandidate = inviteCandidate;
    inviteVm.removeSelectedQuiz = removeSelectedQuiz;
    inviteVm.setMinDate = setMinDate;
    inviteVm.resetForm = resetForm;
    inviteVm.preSelectQuiz = preSelectQuiz;

    function getAllQuizzes(quizId) {
      if (!inviteVm.allQuizes) {
        quizService.getAllQuizzes().then(
          function(quizzes) {
            inviteVm.allQuizes = quizzes;
            preSelectQuiz(quizId);
          },
          function(err) {
            console.error(err);
          }
        );
      } else {
        preSelectQuiz(quizId);
      }
    }

    function preSelectQuiz(quizId) {
      if (!quizId) {
        return;
      }
      var qLen = inviteVm.allQuizes.length;
      var q = inviteVm.allQuizes.find(function(quizz) {
        return quizz.uid == quizId;
      })
      if (q) {
        inviteVm.newInvite.quiz = q;
      }
    }

    function setMinDate() {
      setTimeout(
        function() {
          $datePicker = $("#datePicker");
          $datePicker.attr("min", new Date().toISOString());
          inviteVm.newInvite.validity = sevenDaysFromNow();
        },
        100
      );
    }

    function inviteCandidate() {
      var validationResult = inviteVm.validateInvite(inviteVm.newInvite);

      if (validationResult) {
        SNACKBAR({
          message: validationResult,
          messageType: "error"
        });
        return;
      }

      inviteVm.newInvite.quiz_id = inviteVm.newInvite.quiz.uid;

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
                    quizId: inviteVm.newInvite.quiz_id
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

    inviteVm.validateInvite = function(invite) {
      const badEmail = invite.emails.find(email => !isValidEmail(email));
      if (badEmail) {
        return `${badEmail} isn't a valid email.`;
      }

      if (!invite.validity || invite.validity < new Date()) {
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
    var quizId = $state.params.quizId;

    inviteVm.setMinDate();
    inviteVm.getAllQuizzes(quizId);
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
    editInviteVm.quizId = $stateParams.quizId;

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

        editInviteVm.candidate.validity =
            new Date(editInviteVm.candidate.validity);

        editInviteVm.initAllQuiz();
      });

    function valid(input) {
      if (!isValidEmail(input.email)) {
        return input.email + " isn't a valid email.";
      }
      if (!input.validity) {
        return "Please Enter Valid Date";
      }
      return true;
    }

    function editInvite() {
      editInviteVm.candidate.id = candidateUID;
      editInviteVm.candidate.quiz_id = "";
      editInviteVm.candidate.old_quiz_id = "";

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
        editInviteVm.candidate.old_quiz_id = editInviteVm.quizId;
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
              quizId: editInviteVm.quizId
            });
          },
          function(err) {
            console.error(err);
          }
        );
      }

      // If either the email or the quiz changes, we want to validate that the email
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
        quizId: editInviteVm.quizId
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
    var candidatesVm = this;
    candidatesVm.sortType = "quiz_start";
    candidatesVm.sortReverse = true;

    candidatesVm.expires = expires;
    candidatesVm.showCancelModal = showCancelModal;
    candidatesVm.initiateCancel = initiateCancel;
    candidatesVm.showDeleteModal = showDeleteModal;
    candidatesVm.initiateDelete = initiateDelete;
    candidatesVm.deleteCandFromArray = deleteFromArray;
    candidatesVm.cancel = cancel;
    candidatesVm.delete = deleteCand;
    candidatesVm.percentile = percentile;

    candidatesVm.quizId = $stateParams.quizId;

    if (!candidatesVm.quizId) {
      SNACKBAR({
        message: "Not a valid Quiz",
        messageType: "error"
      });
      $state.transitionTo("invite.add");
    }
    inviteService.getInvitedCandidates(candidatesVm.quizId).then(
      function(data) {
        var quizCandidates = data.data.quiz[0]["quiz.candidate"];

        if (!quizCandidates) {
          SNACKBAR({
            message: "Invite Candidate first to see all candidate",
            messageType: "error"
          });
          $state.transitionTo("invite.add", {
            quizId: candidatesVm.quizId
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

    candidatesVm.copyInviteLink = function(candidate) {
      console.log(candidate)
    }

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
        .cancelInvite(candidate, candidatesVm.quizId)
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
            quizId: candidatesVm.quizId
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
            quizId: candidatesVm.quizId
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

    candidatesVm.resend = function resend(candidate) {
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
          quizId: candidatesVm.quizId
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

angular.module("GruiApp").controller("candidateReportController", [
  "$scope",
  "$rootScope",
  "$stateParams",
  "$state",
  "inviteService",
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
      .getFatReport(cReportVm.candidateID)
      .then(function(fatReport) {
          console.log(fatReport)

          cReportVm.questionStats = parseFatReport(fatReport.data.fatReport[0]["quiz.candidate"])

          return inviteService.getReport(cReportVm.candidateID)
      })
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

        var tagScores = {};

        for (var i = 0; i < questions.length; i++) {
          qn = questions[i];
          d = difficulty(qn.tags);
          if (d != "") {
            statistics[d].total++;
            correct(qn) && statistics[d].correct++;
          }

          qn.tags.forEach(function(tag) {
            var ts = tagScores[tag] = tagScores[tag] || {
              count: 0,
              correct: 0,
              totalPts: 0,
              score: 0,
            };

            ts.count++;
            ts.correct += correct(qn) ? 1 : 0;

            ts.totalPts += maxScore(qn)
            ts.score += candidateScore(qn)
          });

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
        cReportVm.tagScores = Object.entries(tagScores);
        cReportVm.tagScores.sort(function(a,b) {
          return b[1].count - a[1].count;
        })


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

    function candidateScore(q) {
      if (!q.Answered) { return 0; }
      let res = 0
      for (let ans of q.answers) {
        res += q.correct.indexOf(ans) >= 0 ? q.positive : -q.negative
      }
      return res;
    }

    function maxScore(q) {
      return q.positive * q.correct.length
    }

    function parseFatReport(candidates) {
      candidates = candidates.filter(x => x.complete && !x.deleted)

      console.log('fat report for ', candidates.length)

      var qnMap = {}

      function getScore(answers, correct, positive, negative) {
        let res = 0
        for (let ans of answers) {
          res += correct.indexOf(ans) >= 0 ? positive : -negative
        }
        return res;
      }

      for (let k of candidates) {
        for (let qRec of k["candidate.question"]) {
          if (!qRec["candidate.answer"]) {
            continue
          }

          let q = qRec.question[0]
          const correct = q.correct.map(x => x.uid)

          const answers = qRec["candidate.answer"].split(',')
          const skipped = !answers.length || answers[0] === "skip"

          const score = skipped ? 0 : getScore(answers, correct, q.positive, q.negative)

          const curQ = qnMap[q.uid] = qnMap[q.uid] || {
            uid: q.uid,
            name: q.name,
            answerCount: 0,
            skippedCount: 0,
            maxScore: correct.length * q.positive,
            sumScores: 0,
            sumScoresSquared: 0,
          }

          curQ.answerCount ++;
          curQ.skippedCount ++;
          curQ.sumScores += score;
          curQ.sumScoresSquared += score * score;
        }
      }

      for (let q of Object.values(qnMap)) {
        let mean = 0;
        let std = q.maxScore / 2;
        const N = q.answerCount;
        if (N > 2) {
          mean = q.sumScores / N
          std = Math.sqrt(q.sumScoresSquared / (N - 1) - q.sumScores * q.sumScores / N / (N - 1))
        }

        q.mean = mean
        q.std = std
      }

      return qnMap
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
        var direction = 0;
        switch (e.keyCode) {
          case 38:
            // Up arrow key
            direction = -1
            break;
          case 40:
            // Down arrow key
            direction = +1
            break;
          default:
            // Unknown key. Ignore.
            return;
        }

        var nextQuestion = direction + (cReportVm.scrolledQuestion || 0);

        var $question = $("#question" + nextQuestion);
        if (!$question.length && direction < 0) {
          nextQuestion = 0;
          $question = $("#question0");
        }
        if ($question.length) {
          scrollTo($question);
          cReportVm.scrolledQuestion = nextQuestion;
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

    $(".mdl-layout__content").scroll(function() {
      cReportVm.pageScrolled = this.scrollTop > 100
      $scope.$digest()
    });
  }]);

  angular.module("GruiApp").controller("candidatesController", [
    "$scope",
    "$rootScope",
    "$stateParams",
    "$state",
    "$timeout",
    "$templateCache",
    "inviteService",
    candidatesController
  ]);

  angular.module("GruiApp").controller("addCandidatesController", [
    "$state",
    addCandidatesController
  ]);

  angular.module("GruiApp")
    .controller("editInviteController",   [
      "$rootScope",
      "$stateParams",
      "$state",
      "quizService",
      "inviteService",
      editInviteController
    ]);

  angular.module("GruiApp").controller("inviteController", [
    "$scope",
    "$rootScope",
    "$stateParams",
    "$state",
    "quizService",
    "inviteService",
    inviteController
  ]);
})();
