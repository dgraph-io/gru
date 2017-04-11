(function() {
  function profileService($q, $http, $rootScope, MainService) {
    var services = {}; //Object to return

    services.getProfile = function() {
      var deferred = $q.defer();
      var query = "{\
        info(id: root) {\
          company.name \
          company.email \
          company.invite_email \
          company.reject_email \
          company.reject \
          backup \
          backup_days \
        }\
      }";

      MainService.proxy(query).then(function(data) {
        return deferred.resolve(data);
      });
      return deferred.promise;
    };

    services.updateProfile = function(data) {
      var deferred = $q.defer();

      // TODO - Abstract this out into a library so that its easier to add mutations
      // and values are escaped easily.
      var mutation = 'mutation {\n\
    set {\n\
      <root> <company.name> "' +
        data.name +
        '" . \n\
      <root> <company.email> "' +
        data.email +
        '" . \n\
      <root> <backup> "' +
        data.backup +
        '" . \n\
      <root> <backup_days> "' +
        data.backup_days +
        '" . \n';

      if (data.invite_email != "") {
        mutation += '<root> <company.invite_email> "' +
          data.invite_email +
          '" . \n';
      }
      if (data.reject_email != "") {
        mutation += '<root> <company.reject_email> "' +
          data.reject_email +
          '" . \n';
      }

      mutation += '<root> <company.reject> "' +
        (data.reject === true ? "true" : "false") +
        '" . \n\
      }\n\
    }';

      MainService.proxy(mutation).then(function(data) {
        if (data.code != "Success") {
          return deferred.resolve(false);
        }
        return deferred.resolve(true);
      });
      return deferred.promise;
    };

    return services;
  }

  var profileServiceArray = [
    "$q",
    "$http",
    "$rootScope",
    "MainService",
    profileService
  ];

  angular.module("GruiApp").service("profileService", profileServiceArray);
})();
