//Main Config file for Angular.js
//Initialize Angular app, Declare Controller files, Main Controller.
componentHandler.upgradeAllRegistered();
angular.module('GruiApp', ['ngRoute', 'ui.router', "oc.lazyLoad", "door3.css", 'ngSanitize', 'ui.select', 'ui.codemirror',])
    .run(["$rootScope", "$state", "$stateParams", '$window', '$templateCache',
        function ($rootScope, $state, $stateParams, $window, $templateCache) {
            // Set reference to access them from any scope

        }]);
angular.module('GruiApp').config(function(uiSelectConfig) {
  uiSelectConfig.theme = 'select2';
});

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
      window.currentLocation = currentLocation;
      window.previousLocation = previousLocation;
      $rootScope.is_direct_url = (currentLocation == previousLocation);
      isAuthenticated = window.localStorage.getItem("username");

      if($rootScope.is_direct_url) {
          console.log("Hola!");
      }
    });

    var stateChangeStartHandler = function(e, toState, toParams, fromState, fromParams) {
      if(toState.authenticate && !mainVm.isLoggedIn()) {
        $state.transitionTo("login");
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
      'questionController': ['app/components/question/questionController.js'],
      'questionServices': ['app/components/question/questionServices.js'],
      'quizController': ['app/components/quiz/quizController.js'],
      'quizServices': ['app/components/quiz/quizServices.js'],
      'inviteController': ['app/components/invite/inviteController.js'],
      'inviteService': ['app/components/invite/inviteService.js'],
      'quizLandingController': ['app/components/candidate/quizLandingController.js'],
      'candidateController': ['app/components/candidate/candidateController.js'],
      'candidateService': ['app/components/candidate/candidateService.js'],
      'angular-select': ['assets/lib/js/angular-select.min.js'],
      'codeMirror': ['assets/lib/js/codemirror.js'],
      'javascript': ['assets/lib/js/javascript.js'],
      'marked': ['https://cdnjs.cloudflare.com/ajax/libs/marked/0.3.2/marked.min.js'],
      // 'highlight': ['https://cdnjs.cloudflare.com/ajax/libs/highlight.js/8.4/highlight.min.js'],
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
        MainService,
    ];
    angular.module('GruiApp').service("MainService", MainServiceDependency);

// CONTROLLERS, SERVICES FUNCTION DEFINITION

    // MAIN CONTROLLER
    function MainController($rootScope, $state,$stateParams, $http, $q, MainService){
      //ViewModal binding using this, instead of $scope
      //Must be use with ControllerAs syntax in view
      mainVm = this; // $Scope aliase
      mainVm.timerObj;
      mainVm.admin_url = "http://localhost:8082/admin";
      mainVm.candidate_url = "http://localhost:8082";

      //General Methods

      mainVm.getNumber = getNumber;
      mainVm.indexOfObject = indexOfObject;
      mainVm.hasKey = hasKey;
      mainVm.isObject = isObject;
      mainVm.goTo = goTo;
      mainVm.unescapeText = unescapeText;

      mainVm.isLoggedIn = isLoggedIn;
      mainVm.logout = logout;
      mainVm.isValidCandidate = isValidCandidate;
      mainVm.markDownFormat = markDownFormat;

      mainVm.getAllTags = getAllTags;

      // General Functions
      function markDownFormat(content) {
        return marked(content);
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

      function isValidCandidate() {
        if(localStorage.getItem('ctoken')) {
          return true;
        }
        return false
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
        // MainService("/get-all-tags")
        var deferred = $q.defer();

        $http.defaults.headers.common['Authorization'] = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.eb0qBs-z-zbhRorx4PZxakiDfSC_HyY741ZES0hOVdU';

        console.log($http.defaults.headers.common['Authorization']);
        var req = {
          method: 'GET',
          url: mainVm.admin_url + '/get-all-tags',
        }

        $http(req)
        .then(function(data) { 
            deferred.resolve(data.data);
          },
          function(response, code) {
            mainVm.showAjaxLoader = false;
            deferred.reject(response);
            SNACKBAR({
              message: "Something went wrong",
              messageType: "error"
            })
          }
        );

        return deferred.promise;
      }
    }

    // MAIN Service
    function MainService($http, $q){
      var services = {}; //Object to return
      services.post = post;
      services.get = get;
      services.put = put;

      function post(url, data, hideLoader){
        var deferred = $q.defer();

        var req = {
          method: 'POST',
          url: mainVm.admin_url + url,
          data: data,
        }
        if(!mainVm.isLoggedIn()) {
          req.url = mainVm.candidate_url + url;
        }

        if(url == "/login") {
          $http.defaults.headers.common['Authorization'] = 'Basic ' + btoa(data.user + ':' + data.password);
          delete req.data;
        } else {
          candidateToken = localStorage.getItem('candidate_token');
          if(candidateToken) {
            setAuth('Bearer ' + candidateToken);
          } else {
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
            mainVm.showAjaxLoader = false;
            deferred.reject(response);
          }
        );

        return deferred.promise;
      }

      function get(url) {
        var deferred = $q.defer();
        setAuth('Bearer ' + localStorage.getItem('token'));
        var req = {
          method: 'GET',
          url: mainVm.admin_url + url,
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
        console.log(auth_token)
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
