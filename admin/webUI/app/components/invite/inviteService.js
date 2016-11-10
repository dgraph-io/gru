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

    services.alreadyInvited = function(quizId, email) {
      var deferred = $q.defer();
      // TODO - User filter on email after incorporating Dgraph schema.
      var query = "{\
                  quiz(_uid_: " + quizId + ") {\
                          quiz.candidate {\
                                  email\
                          }\
                  }\
          }"

      services.proxy(query).then(function(data) {
        var candidates = data.quiz[0]["quiz.candidate"];
        if (candidates) {
          for (var i = 0; i < candidates.length; i++) {
            if (candidates[i].email === email) {
              return deferred.resolve(true);
            }
          }
        }
        return deferred.resolve(false);
      });
      return deferred.promise;
    }

    services.resendInvite = function(candidateID) {
      var deferred = $q.defer();
      // TODO - User filter on email after incorporating Dgraph schema.
      var query = "{\n\
        quiz.candidate(_uid_: " + candidateID + ") {\n\
          email\n\
          token\n\
          validity\n\
        }\n\
      }"

      services.proxy(query).then(function(data) {
        var candidate = data["quiz.candidate"][0];
        if (candidate == null) {
          return deferred.resolve({
            success: false,
            message: "No candidate found."
          });
        }
        return candidate
      }).then(function(candidate) {
        var payload = {
          "email": candidate.email,
          "token": candidate.token,
          "validity": candidate.validity
        }

        MainService.post('/candidate/invite/' + candidateID, payload).then(function(data) {
          return deferred.resolve({
            sucess: true,
            message: data.Message
          })
        })
      });
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
      <_uid_:" + quizId + "> < quiz.candidate> <_uid_:" + candidate._uid_ + "> .\n\
      }\n\
    }"
      services.proxy(mutation).then(function(data) {
        if (data.code == "ErrorOk") {
          return deferred.resolve(true);
        }
        console.log(data)
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
      services.proxy(mutation).then(function(data) {
        if (data.code == "ErrorOk") {
          return deferred.resolve(true);
        }
        console.log(data)
        return deferred.resolve(false);
      });
      return deferred.promise;
    }


    // TODO - Move to a location where other services can access this.
    services.proxy = function(data) {
      return MainService.post('/proxy', data);
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
