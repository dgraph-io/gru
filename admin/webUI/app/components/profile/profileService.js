(function() {

  function profileService($q, $http, $rootScope, MainService) {

    var services = {}; //Object to return

    services.getProfile = function() {
      var deferred = $q.defer();
      var query = "{\
                  info(_xid_: root) {\
                          company.name \
                          company.email \
                  }\
          }"

      MainService.proxy(query).then(function(data) {
        return deferred.resolve(data)
      });
      return deferred.promise;
    }

    services.updateProfile = function(data) {
      var deferred = $q.defer();

      // TODO - Abstract this out into a library so that its easier to add mutations
      // and values are escaped easily.
      var mutation = "mutation {\n\
    set {\n\
      <root> <company.name> \"" + data.name + "\" . \n\
      <root> <company.email> \"" + data.email + "\" . \n\
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

    return services;
  }

  var profileServiceArray = [
    "$q",
    "$http",
    "$rootScope",
    "MainService",
    profileService
  ];

  angular.module('GruiApp').service('profileService', profileServiceArray);

})();
