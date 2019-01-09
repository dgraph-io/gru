angular.module("GruiApp").service("profileService", [
  "$http",
  "$rootScope",
  "MainService",
  function profileService($http, $rootScope, MainService) {
    return {
      getProfile: function() {
        var query = "{\
          info(func: has(is_company_info)) {\
            uid \
            company.name \
            company.email \
            company.invite_email \
            company.reject_email \
            company.reject \
            backup \
            backup_days \
          }\
        }";

        return MainService.proxy(query).then(function(data) {
          return data.data;
        });
      },

      updateProfile: function(data) {
        var uid = data.uid || "_:co"
        var mutation = '{\n\
          set {\n\
            <' + uid + '> <is_company_info> "' + data.name + '" . \n\
            <' + uid + '> <company.name> "' + data.name + '" . \n\
            <' + uid + '> <company.email> "' + data.email + '" . \n\
            <' + uid + '> <backup> "' + data.backup + '" . \n\
            <' + uid + '> <backup_days> "' + data.backup_days + '" . \n';

        if (data.invite_email != "") {
          mutation += '<' + uid + '> <company.invite_email> "' +
            data.invite_email +
            '" . \n';
        }
        if (data.reject_email != "") {
          mutation += '<' + uid + '> <company.reject_email> "' +
            data.reject_email +
            '" . \n';
        }

        mutation += '<' + uid + '> <company.reject> "' +
          (data.reject ? "true" : "false") + '" . \n\
          }\n\
        }';

        return MainService.mutateProxy(mutation).then(function(data) {
          return data.code == "Success";
        });
      },
    };
  },
]);
