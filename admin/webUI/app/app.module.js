//Main Config file for Angular.js
//Initialize Angular app, Declare Controller files, Main Controller.
componentHandler.upgradeAllRegistered();
angular.module('GruiApp', ['ngRoute', 'ui.router', "oc.lazyLoad", "door3.css", 'ngSanitize', 'ui.select', 'ui.codemirror',])
    .run(["$rootScope", "$state", "$stateParams", '$window', '$templateCache',
        function ($rootScope, $state, $stateParams, $window, $templateCache) {
            // Set reference to access them from any scope
            $rootScope.SEO = {};
            $rootScope.SEO.title = "Gru";

        }]);
angular.module('GruiApp').config(function(uiSelectConfig) {
  uiSelectConfig.theme = 'select2';
});

var hostname = "http://localhost:2020"

//Run After view has been loaded 
angular.module('GruiApp').run(function($rootScope, $location, $timeout, $state) {
    //Run After view has been loaded 
    $rootScope.$on('$viewContentLoaded', function() {
        componentHandler.upgradeAllRegistered();
        $timeout(function() {
            componentHandler.upgradeAllRegistered();
            componentHandler.upgradeDom();
        }, 1000); 
    });

    $rootScope.$on("$locationChangeStart", function(e, currentLocation, previousLocation){
      // window.currentLocation = currentLocation;
      // window.previousLocation = previousLocation;
      // $rootScope.is_direct_url = (currentLocation == previousLocation);
      // isAuthenticated = window.localStorage.getItem("username");

      // if($rootScope.is_direct_url) {
      //     // console.log("Hola!");
      // }
    });

    var stateChangeStartHandler = function(e, toState, toParams, fromState, fromParams) {
      if(toState.authenticate || toState.name == "login") {
        mainVm.base_url = hostname + "/api/admin";
      } else {
        mainVm.base_url = hostname + "/api";
      }
      if(toState.authenticate && !mainVm.isLoggedIn()) {
        $state.transitionTo("login");
        e.preventDefault();
      }
      if(toState.name == "login" && mainVm.isLoggedIn()) {
        $state.transitionTo("root");
        e.preventDefault();
      }
      (function(){
        setTimeout(function() {
          $mdl_input = $(".mdl-textfield__input")
          for(var i=0; i < $mdl_input.length; i++) {
            var this_field = $($mdl_input[i]);
            this_field.removeClass("is-invalid");

            if(this_field.attr('type') == "date"){
              this_field.parent().addClass("is-focused");
            }
          }
        }, 700);
      })();
    };
    $rootScope.$on('$stateChangeStart', stateChangeStartHandler);


    $rootScope.updgradeMDL = function(){
        $timeout(function() {
            componentHandler.upgradeAllRegistered();
        }, 1000);
    }

    //Run After ng-include has been loaded
    $rootScope.$on("$includeContentLoaded", function(event, templateName){
        componentHandler.upgradeAllRegistered();
    });
})



// LAZY LOAD CONFIGURATION
angular.module('GruiApp').config(['$ocLazyLoadProvider','$httpProvider', 'APP_REQUIRES', function ($ocLazyLoadProvider,$httpProvider, APP_REQUIRES) {
    'use strict';

    $httpProvider.defaults.useXDomain = true;
    delete $httpProvider.defaults.headers.common['X-Requested-With'];

    // Lazy Load modules configuration
    $ocLazyLoadProvider.config({
        debug: false,
        events: true,
        modules: APP_REQUIRES.modules,
    });

}]);

// SCRIPT NAME CONFIG For OCLAZYLOAD
angular.module('GruiApp').constant('APP_REQUIRES', {
    // jQuery based/Cusomt/standalone scripts
    scripts: {
      'homeController': ['app/components/home/homeController.js'],
      'loginController': ['app/components/login/loginController.js'],
      'loginService': ['app/components/login/loginService.js'],
      'questionController': ['app/components/question/questionController.js?v=20161027-1'],
      'questionServices': ['app/components/question/questionServices.js'],
      'quizController': ['app/components/quiz/quizController.js'],
      'quizServices': ['app/components/quiz/quizServices.js'],
      'inviteController': ['app/components/invite/inviteController.js?v=20161018-1'],
      'inviteService': ['app/components/invite/inviteService.js?v=20161018-1'],
      'quizLandingController': ['app/components/candidate/quizLandingController.js?v=20161027-1'],
      'quizLandingService': ['app/components/candidate/quizLandingService.js?v=20161027-1'],
      'candidateController': ['app/components/candidate/candidateController.js?v=20161025-1'],
      'candidateService': ['app/components/candidate/candidateService.js'],
      'angular-select': ['assets/lib/js/angular-select.min.js'],
      'codeMirror': ['assets/lib/js/codemirror.js'],
      'javascript': ['assets/lib/js/javascript.js'],
      'marked': ['https://cdnjs.cloudflare.com/ajax/libs/marked/0.3.2/marked.min.js'],
      'highlight': ['assets/lib/js/highlight.pack.js'],
    },
});

/**=========================================================
 * Module: helpers.js
 * Provides helper functions for routes definition
 =========================================================*/
angular.module('GruiApp').provider('RouteHelpers', ['APP_REQUIRES', function (appRequires) {
    "use strict";
    // Generates a resolve object by passing script names
    // previously configured in constant.APP_REQUIRES
    this.resolveFor = function () {
      var _args = arguments;
      return {
        deps: ['$ocLazyLoad', '$q', function ($ocLL, $q) {
          // Creates a promise chain for each argument
          var promise = $q.when(1); // empty promise
          for (var i = 0, len = _args.length; i < len; i++) {
            promise = andThen(_args[i]);
          }
          return promise;

          // creates promise to chain dynamically
          function andThen(_arg) {
            // also support a function that returns a promise
            if (typeof _arg == 'function'){
              return promise.then(_arg);
            }
            else {
              return promise.then(function () {
                // if is a module, pass the name. If not, pass the array
                var whatToLoad = getRequired(_arg);
                // simple error check
                if (!whatToLoad) return $.error('Route resolve: Bad resource name [' + _arg + ']');
                // finally, return a promise
                return $ocLL.load(whatToLoad);
              });
            }
          }

          function getRequired(name) {
            if (appRequires.modules)
                for (var m in appRequires.modules)
                    if (appRequires.modules[m].name && appRequires.modules[m].name === name)
                        return appRequires.modules[m];
            return appRequires.scripts && appRequires.scripts[name];
          }
        }]
      };
    }; // resolveFor

    // not necessary, only used in config block for routes
    this.$get = function () {
    };

}]);


// GENERAL CONTROLLER, SERVICE, DIRECTIVE,FILTER
(function(){
    
// CONTROLLERs, SERVICEs, DIRECTIVES DECLARATION
    
    // MAIN CONTROLLER declaration
    var MainDependency = [
      "$scope",
      "$rootScope",
      "$state",
      "$stateParams",
      "$http",
      "$q",
      "MainService",
      MainController,
    ];
    angular.module('GruiApp').controller("MainController", MainDependency);

    var MainServiceDependency = [
        "$http",
        "$q",
        "$state",
        MainService,
    ];
    angular.module('GruiApp').service("MainService", MainServiceDependency);

// CONTROLLERS, SERVICES FUNCTION DEFINITION

    // MAIN CONTROLLER
    function MainController($scope,$rootScope,$state,$stateParams, $http, $q, MainService){
      //ViewModal binding using this, instead of $scope
      //Must be use with ControllerAs syntax in view
      mainVm = this; // $Scope aliase
      mainVm.timerObj;
      mainVm.admin_url = hostname + "/api/admin";
      mainVm.candidate_url = hostname + "/api";
      mainVm.showModal = false;

      //General Methods

      mainVm.getNumber = getNumber;
      mainVm.indexOfObject = indexOfObject;
      mainVm.hasKey = hasKey;
      mainVm.isObject = isObject;
      mainVm.goTo = goTo;
      mainVm.unescapeText = unescapeText;
      mainVm.objLen = objLen;

      mainVm.isLoggedIn = isLoggedIn;
      mainVm.logout = logout;
      mainVm.isValidCandidate = isValidCandidate;
      mainVm.markDownFormat = markDownFormat;
      mainVm.parseGoTime = parseGoTime;
      mainVm.openModal = openModal;
      mainVm.hideModal = hideModal;
      mainVm.timeoutModal = timeoutModal;
      mainVm.initNotification = initNotification;
      mainVm.hideNotification = hideNotification;

      mainVm.getAllTags = getAllTags;

      // General Functions
      function markDownFormat(content) {
        if(!content) {
          return
        }
        return marked(content, {
          gfm: true,
        });
      }

      function getNumber(num) {
        return new Array(num);   
      }

      function indexOfObject(arr, obj){
        if(!arr) {
          return -1;
        }
        for(var i = 0; i < arr.length; i++){
            if(angular.equals(arr[i], obj)){
                return i;
            }
        };
        return -1;
      }

      function objLen(object) {
        return Object.keys(object).length;
      }


      function unescapeText(question) {
        return unescape(question);
      }

      function isLoggedIn() {
        if(localStorage.getItem('token')) {
          return true;
        }
        return false
      }

      function logout(){
        localStorage.removeItem('token');
        $state.transitionTo("login");
      }

      function openModal(setting) {
        if(!setting.template) {
          return
        }
        mainVm.modal = {
          template : setting.template
        }
        mainVm.showModal = true;
      }

      function hideModal() {
        mainVm.modal = {};
        mainVm.showModal =  false;
      }

      function timeoutModal() {
        var modalContent = "Sorry to inform you but we are facing some severe problems.";
        modalContent += "<div>Please send us a email on - contact@dgraph.io</div>";
        mainVm.openModal({
          // template: "./app/shared/_server_crash.html",
          template: modalContent,
        });
        $rootScope.$broadcast("endQuiz", {
          message: modalContent,
        });
      }

      function isValidCandidate() {
        var quizToken = localStorage.getItem("quiz_token");
        MainService.post("/validate/" + quizToken)
        .then(function(data){
          return data;
        }, function(err){
          if(err.status == 401) {
            SNACKBAR({
              message: err.data.Message,
              messageType: "error",
            })
          }
        })
      }

      function hasKey(obj, key){
        if(!obj) {
          return false;
        }
        return (key in obj)
      }

      function isObject(obj) {
        return Object.prototype.toString.call(obj) == "[object Object]"
      }

      function goTo(state) {
        $state.transitionTo(state);
      }

      function getAllTags(){
        return MainService.get("/get-all-tags")
      }

      function parseGoTime(time) {
        var duration = Duration.parse(time);
        var totalSec = duration.seconds();

        return {
          hours: Math.floor(totalSec / 3600),
          minutes: parseInt((totalSec / 60) % 60, 10),
          seconds: parseInt(totalSec % 60, 10),
        }
      }

      function initNotification() {
        mainVm.consecutiveError = mainVm.consecutiveError || 0;
        mainVm.consecutiveError += 1;

        if(mainVm.consecutiveError < 2) {
          return
        }
        if(mainVm.consecutiveError >= 20) {
          mainVm.timeoutModal();
        }
        mainVm.notification = {};
        mainVm.showNotification = true;

        mainVm.notification.class = "notification-error";
        mainVm.notification.message = "Can't connect to the server. We will be back up in a bit...";
      }

      function hideNotification(rejected) {
        if(!rejected) {
          mainVm.notification.class = "notification-success";
          mainVm.notification.message = "Connected to Server";
        }
        mainVm.consecutiveError = 0;
        setTimeout(function() {
          mainVm.showNotification = false;
          mainVm.notification.class = "";
          $scope.$apply();
        }, 2000);
      }

    }

    // MAIN Service
    function MainService($http, $q, $state){
      var services = {}; //Object to return
      services.post = post;
      services.get = get;
      services.put = put;

      function post(url, data, hideLoader){
        var deferred = $q.defer();

        var req = {
          method: 'POST',
          url: mainVm.base_url + url,
          data: data,
          timeout: 30000,
        }

        if(url == "/login") {
          $http.defaults.headers.common['Authorization'] = 'Basic ' + btoa(data.user + ':' + data.password);
          delete req.data;
        } else {
          // req.url = mainVm.candidate_url + url;
          candidateToken = JSON.parse(localStorage.getItem('candidate_info'));

          if(window.location.hash.indexOf("admin") == -1 && candidateToken && candidateToken.token) {
            setAuth('Bearer ' + candidateToken.token);
          } else {
            // req.url = mainVm.admin_url + url;
            setAuth('Bearer ' + localStorage.getItem('token'));
          }
        }

        if(!hideLoader) {
          mainVm.showAjaxLoader = true;
        }
        $http(req)
        .then(function(data) { 
            mainVm.showAjaxLoader = false;
            deferred.resolve(data.data);
          },
          function(response, code) {
            if(!hideLoader) {
              mainVm.showAjaxLoader = false;
            }
            deferred.reject(response);
          }
        );

        return deferred.promise;
      }

      function get(url) {
        var deferred = $q.defer();
        var req = {
          method: 'GET',
          url: mainVm.admin_url + url,
        }

        candidateToken = JSON.parse(localStorage.getItem('candidate_info'));

        if(window.location.hash.indexOf("admin") == -1 && candidateToken && candidateToken.token) {
          setAuth('Bearer ' + candidateToken.token);
        } else {
          // req.url = mainVm.admin_url + url;
          setAuth('Bearer ' + localStorage.getItem('token'));
        }

        mainVm.showAjaxLoader = true;
        $http(req)
        .then(function(data) { 
            mainVm.showAjaxLoader = false;
            deferred.resolve(data.data);
          },
          function(response, code) {
            mainVm.showAjaxLoader = false;
            deferred.reject(response);
          }
        );

        return deferred.promise;
      }
      
      function put(url, data) {
        var deferred = $q.defer();
        var auth_token = 'Bearer ' + localStorage.getItem('token')

        setAuth(auth_token);
        var req = {
          method: 'PUT',
          url: mainVm.admin_url + url,
          data: data,
        }
        mainVm.showAjaxLoader = true;
        $http(req)
        .then(function(data) { 
            mainVm.showAjaxLoader = false;
            deferred.resolve(data.data);
          },
          function(response, code) {
            mainVm.showAjaxLoader = false;
            deferred.reject(response);
          }
        );

        return deferred.promise;
      }

      function setAuth(auth) {
        $http.defaults.headers.common['Authorization'] = auth;
      }

      return services;
    }


})();
