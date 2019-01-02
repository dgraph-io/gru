angular.module("GruiApp").service("inviteService", [
  "$q",
  "$http",
  "$rootScope",
  "MainService",
  function inviteService($q, $http, $rootScope, MainService) {
    var services = {}; //Object to return

    services.inviteCandidate = function(data) {
      return MainService.post("/candidate", data);
    };

    services.getInvitedCandidates = function(data) {
      return MainService.get("/candidates?quiz_id=" + data);
    };

    services.getCandidate = function(data) {
      return MainService.get("/candidate/" + data);
    };

    services.editInvite = function(data) {
      return MainService.put("/candidate/" + data.id, data);
    };

    services.getReport = function(candidateID) {
      return MainService.get("/candidate/report/" + candidateID);
    };

    services.getResume = function(candidateID) {
      $http.defaults.headers.common["Authorization"] =
        "Bearer " + localStorage.getItem("token");
      mainVm.showAjaxLoader = true;
      $http({
        method: "GET",
        url: mainVm.admin_url + "/candidate/resume/" + candidateID,
        responseType: "arraybuffer"
      })
        .success(function(data, status, headers) {
          mainVm.showAjaxLoader = false;
          headers = headers();

          // So much stuff to download the file.
          var filename = headers["x-filename"];
          var contentType = headers["content-type"];

          var linkElement = document.createElement("a");
          try {
            var blob = new Blob([data], { type: contentType });
            var url = window.URL.createObjectURL(blob);

            linkElement.setAttribute("href", url);
            linkElement.setAttribute("download", filename);

            var clickEvent = new MouseEvent("click", {
              view: window,
              bubbles: true,
              cancelable: false
            });
            linkElement.dispatchEvent(clickEvent);
          } catch (ex) {
            console.log(ex);
          }
        })
        .error(function(data) {
          mainVm.showAjaxLoader = true;
          console.log(data);
        });
    };

    services.alreadyInvited = function(quizId, emails) {
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
        console.log("got proxied ", data);
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
    };

    services.resendInvite = function(candidate) {
      var deferred = $q.defer();
      // We update the validity to be 7 days from now on resending the invite.
      var new_validity = new Date();
      new_validity.setDate(new_validity.getDate() + 7);
      var date = new_validity.getDate(),
        month = new_validity.getMonth() + 1,
        year = new_validity.getFullYear();

      if (date < 10) {
        date = "0" + date;
      }
      if (month < 10) {
        month = "0" + month;
      }

      var val = year + "-" + month + "-" + date + " 00:00:00 +0000 UTC";

      var mutation =
        "mutation {\n\
            set {\n\
              <" +
        candidate.uid +
        '> <validity> "' +
        val +
        '" .\n\
            }\n\
          }';

      MainService.proxy(mutation).then(function(res) {
        if (res.code != MainService.dgraphSuccess) {
          return deferred.resolve({
            success: false,
            message: "Validity couldn't be extended."
          });
        }
      });
      candidate.validity = val;

      var payload = {
        email: candidate.email,
        token: candidate.token,
        validity: candidate.validity
      };

      MainService.post(
        "/candidate/invite/" + candidate.uid,
        payload
      ).then(function(data) {
        return deferred.resolve({
          sucess: true,
          message: data.Message
        });
      });
      return deferred.promise;
    };

    services.cancelInvite = function(candidate, quizId) {
      var deferred = $q.defer();

      // TODO - Abstract this out into a library so that its easier to add mutations
      // and values are escaped easily.
      var mutation =
        "mutation {\n\
    delete {\n\
      <" +
        candidate.uid +
        '> <email> "' +
        candidate.email +
        '" . \n\
      <' +
        candidate.uid +
        '> <invite_sent> "' +
        candidate.invite_sent +
        '" . \n\
      <' +
        candidate.uid +
        '> <token> "' +
        candidate.token +
        '" . \n\
      <' +
        candidate.uid +
        '> <validity> "' +
        candidate.validity +
        '" . \n\
      <' +
        candidate.uid +
        '> <complete> "' +
        candidate.complete +
        '" . \n\
      <' +
        candidate.uid +
        "> <candidate.quiz> <" +
        quizId +
        "> . \n\
      <" +
        quizId +
        "> <quiz.candidate> <" +
        candidate.uid +
        "> .\n\
      }\n\
    }";
      MainService.proxy(mutation).then(function(data) {
        if (data.code == MainService.dgraphSuccess) {
          return deferred.resolve(true);
        }
        return deferred.resolve(false);
      });
      return deferred.promise;
    };

    services.deleteCand = function(candidateId) {
      var deferred = $q.defer();

      // TODO - Abstract this out into a library so that its easier to add mutations
      // and values are escaped easily.
      var mutation =
        "mutation {\n\
    set {\n\
      <" +
        candidateId +
        '> <deleted> "true" . \n\
    }\n\
      }';
      MainService.proxy(mutation).then(function(data) {
        if (data.code == MainService.dgraphSuccess) {
          return deferred.resolve(true);
        }
        console.log(data);
        return deferred.resolve(false);
      });
      return deferred.promise;
    };

    return services;
  }
]);
