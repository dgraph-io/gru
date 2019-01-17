angular.module('GruiApp').service('candidateService', [
  "$q",
  "$http",
  "MainService",
  function candidateService($q, $http, MainService) {
    return {
      getQuestion: function() {
        return MainService.post('/quiz/question');
      },

      sendFeedback: function(data) {
        mainVm.showAjaxLoader = true;
        return $http({
            method: 'POST',
            url: mainVm.candidate_url + '/quiz/feedback',
            data: $.param(data),
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded'
            }
          })
          .then(function(data) {
              mainVm.showAjaxLoader = false;
              return data;
            },
            function(response, code) {
              mainVm.showAjaxLoader = false;
              throw response;
            }
          );
      },

      submitAnswer: function(requestData) {
        mainVm.showAjaxLoader = true;
        return $http({
            method: 'POST',
            url: mainVm.candidate_url + '/quiz/answer',
            data: $.param(requestData),
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded'
            }
          })
          .then(function(data) {
              mainVm.showAjaxLoader = false;
              return data;
            },
            function(response, code) {
              mainVm.showAjaxLoader = false;
              throw response;
            }
          );
      },

      getTime: function() {
        return MainService.post('/quiz/ping', "", true);
      },
    }
  }
]);
