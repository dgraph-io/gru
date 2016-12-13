(function() {

  function inviteService($q, $http, $rootScope, MainService) {

    var services = {}; //Object to return

    services.inviteCandidate = function(data) {
      return MainService.post('/candidate', data);
    }

    services.getInvitedCandidates = function(data) {
      return MainService.get('/candidates?quiz_id=' + data);
    }

    services.getCandidate = function(data) {
      return MainService.get('/candidate/' + data);
    }

    services.editInvite = function(data) {
      return MainService.put('/candidate/' + data.id, data);
    }

    services.getReport = function(candidateID) {
      return MainService.get('/candidate/report/' + candidateID);
    }

    services.getResume = function(candidateID) {
      $http.defaults.headers.common['Authorization'] = 'Bearer ' + localStorage.getItem('token')
      mainVm.showAjaxLoader = true;
      $http({
        method: 'GET',
        url: mainVm.admin_url + '/candidate/resume/' + candidateID,
        responseType: 'arraybuffer'
      }).success(function(data, status, headers) {
        mainVm.showAjaxLoader = false;
        headers = headers();

        // So much stuff to download the file.
        var filename = headers['x-filename'];
        var contentType = headers['content-type'];

        var linkElement = document.createElement('a');
        try {
          var blob = new Blob([data], { type: contentType });
          var url = window.URL.createObjectURL(blob);

          linkElement.setAttribute('href', url);
          linkElement.setAttribute("download", filename);

          var clickEvent = new MouseEvent("click", {
            "view": window,
            "bubbles": true,
            "cancelable": false
          });
          linkElement.dispatchEvent(clickEvent);
        } catch (ex) {
          console.log(ex);
        }
      }).error(function(data) {
        mainVm.showAjaxLoader = true;
        console.log(data);
      });
    }

    services.alreadyInvited = function(quizId, emails) {
      var deferred = $q.defer();
      // TODO - User filter on email after incorporating Dgraph schema.
      var query = "{\
                  quiz(_uid_: " + quizId + ") {\
                          quiz.candidate {\
                                  email\
                          }\
                  }\
          }"

       MainService.proxy(query).then(function(data) {
        var candidates = data.quiz[0]["quiz.candidate"];
        for (var j = 0; j < emails.length; j++) {
          email = emails[j]
          for (var i = 0; i < candidates.length; i++) {
            if (candidates[i].email === email) {
              return deferred.resolve(email);
            }
          }
        }
        return deferred.resolve("");
      });
      return deferred.promise;
    }

    services.resendInvite = function(candidate) {
      var deferred = $q.defer();
      // We update the validity to be 7 days from now on resending the invite.
      var new_validity = new Date();
      new_validity.setDate(new_validity.getDate() + 7)
      var date = new_validity.getDate(),
        month = new_validity.getMonth() + 1,
        year = new_validity.getFullYear();

      var val = year + "-" + month + "-" + date + " 00:00:00 +0000 UTC";

      var mutation = "mutation {\n\
            set {\n\
              <_uid_:" + candidate._uid_ + "> <validity> \"" + val + "\" .\n\
            }\n\
          }"

      MainService.proxy(mutation).then(function(res) {
        if (res.code != "ErrorOk") {
          return deferred.resolve({
            success: false,
            message: "Validity couldn't be extended."
          })
        }
      })
      candidate.validity = val

      var payload = {
        "email": candidate.email,
        "token": candidate.token,
        "validity": candidate.validity
      }

      MainService.post('/candidate/invite/' + candidate._uid_, payload).then(function(data) {
        return deferred.resolve({
          sucess: true,
          message: data.Message
        })
      })
      return deferred.promise;
    }

    services.cancelInvite = function(candidate, quizId) {
      var deferred = $q.defer();

      // TODO - Abstract this out into a library so that its easier to add mutations
      // and values are escaped easily.
      var mutation = "mutation {\n\
    delete {\n\
      <_uid_:" + candidate._uid_ + "> <email> \"" + candidate.email + "\" . \n\
      <_uid_:" + candidate._uid_ + "> <invite_sent> \"" + candidate.invite_sent + "\" . \n\
      <_uid_:" + candidate._uid_ + "> <token> \"" + candidate.token + "\" . \n\
      <_uid_:" + candidate._uid_ + "> <validity> \"" + candidate.validity + "\" . \n\
      <_uid_:" + candidate._uid_ + "> <complete> \"" + candidate.complete + "\" . \n\
      <_uid_:" + candidate._uid_ + "> <candidate.quiz> <_uid_:" + quizId + "> . \n\
      <_uid_:" + quizId + "> <quiz.candidate> <_uid_:" + candidate._uid_ + "> .\n\
      }\n\
    }"
      MainService.proxy(mutation).then(function(data) {
        if (data.code == "ErrorOk") {
          return deferred.resolve(true);
        }
        return deferred.resolve(false);
      });
      return deferred.promise;
    }

    services.deleteCand = function(candidateId) {
      var deferred = $q.defer();

      // TODO - Abstract this out into a library so that its easier to add mutations
      // and values are escaped easily.
      var mutation = "mutation {\n\
    set {\n\
      <_uid_:" + candidateId + "> <deleted> \"true\" . \n\
    }\n\
      }"
      MainService.proxy(mutation).then(function(data) {
        if (data.code == "ErrorOk") {
          return deferred.resolve(true);
        }
        console.log(data)
        return deferred.resolve(false);
      });
      return deferred.promise;
    }

    return services;

  }

  var inviteServiceArray = [
    "$q",
    "$http",
    "$rootScope",
    "MainService",
    inviteService
  ];

  angular.module('GruiApp').service('inviteService', inviteServiceArray);

})();
