angular.module("GruiApp").service("inviteService", [
  "$q",
  "$http",
  "$rootScope",
  "MainService",
  function inviteService($q, $http, $rootScope, MainService) {
    return {
      inviteCandidate: function(data) {
        return MainService.post("/candidate", data);
      },

      getInvitedCandidates: function(data) {
        return MainService.get("/candidates?quiz_id=" + data);
      },

      getCandidate: function(data) {
        return MainService.get("/candidate/" + data);
      },

      editInvite: function(data) {
        return MainService.put("/candidate/" + data.id, data);
      },

      getReport: function(candidateID) {
        return MainService.get("/candidate/report/" + candidateID);
      },

      alreadyInvited: function(quizId, emails) {
        // TODO - User filter on email after incorporating Dgraph schema.
        var query =
          "{\
            quiz(func: uid(" + quizId + ")) {\
              uid \
              quiz.candidate {\
                email\
              }\
            }\
          }";

        return MainService.proxy(query).then(function(data) {
          if (!data || !data.data) {
            return "";
          }
          var candidates = data.data.quiz[0]["quiz.candidate"];
          if (candidates === undefined) {
            return "";
          }
          for (var j = 0; j < emails.length; j++) {
            email = emails[j];
            for (var i = 0; i < candidates.length; i++) {
              if (candidates[i].email === email) {
                return email;
              }
            }
          }
          return "";
        });
      },

      resendInvite: function(candidate) {
        // We update the validity to be 7 days from now on resending the invite.
        var validity = new Date(Date.now() + 3600 * 24 * 7 * 1000).toISOString();
        var mutation = "{\n\
          set {\n\
            <" + candidate.uid + '> <validity> "' +  validity + '" .\n\
          }}';

        return MainService.mutateProxy(mutation).then(function(res) {
          if (!res.data || res.data.code != MainService.dgraphSuccess) {
            throw {
              success: false,
              message: "Validity couldn't be extended."
            }
          }
        }).then(function() {
          return MainService.post(
            "/candidate/invite/" + candidate.uid, {
              email: candidate.email,
              token: candidate.token,
              validity: validity,
            })
        }).then(function(res) {
          return {
            sucess: res.Success,
            message: res.Message + " " + res.Error
          }
        });
      },

      cancelInvite: function(candidate, quizId) {
        var mutation =
          "{\n\
            delete {\n\
              <" + candidate.uid + "> * * .\n\
              <" + quizId + "> <quiz.candidate> <" + candidate.uid + "> .\n\
            }\n\
          }";
        return MainService.mutateProxy(mutation).then(function(data) {
          return (data.data && data.data.code == MainService.dgraphSuccess);
        });
      },

      deleteCand: function(candidateId) {
        var mutation =
          "{\n\
            set {\n\
              <" + candidateId + '> <deleted> "true" . \n\
            }}';
        return MainService.mutateProxy(mutation).then(function(data) {
          return (data.code == MainService.dgraphSuccess);
        });
      },
    }
  },
]);
