// APP START'angular.embed.timepicker'
// -----------------------------------
angular.module('GruiApp', ['ngRoute', 'ui.router', "oc.lazyLoad"])
    .run(["$rootScope", "$state", "$stateParams", '$window', '$templateCache',
        function ($rootScope, $state, $stateParams, $window, $templateCache) {
            // Set reference to access them from any scope

        }]);


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

    $rootScope.updgradeMDL = function(){
        $timeout(function() {
            componentHandler.upgradeAllRegistered();
        }, 100);
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
        'questionController': ['app/components/question/questionController.js'],
        'questionServices': ['app/components/question/questionServices.js'],
        'quizController': ['app/components/quiz/quizController.js'],
        'quizServices': ['app/components/quiz/quizServices.js'],
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
                    if (typeof _arg == 'function')
                        return promise.then(_arg);
                    else
                        return promise.then(function () {
                            // if is a module, pass the name. If not, pass the array
                            var whatToLoad = getRequired(_arg);
                            // simple error check
                            if (!whatToLoad) return $.error('Route resolve: Bad resource name [' + _arg + ']');
                            // finally, return a promise
                            return $ocLL.load(whatToLoad);
                        });
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
        "$window",
        "$compile",
        "$timeout",
        "$state",
        "$location",
        "$http",
        MainController,
    ];
    angular.module('GruiApp').controller("MainController", MainDependency);

// CONTROLLERS, SERVICES FUNCTION DEFINITION

    // MAIN CONTROLLER
    function MainController($scope, $rootScope, $window, $compile,$timeout,$state,$location,$http){
        //ViewModal binding using this, instead of $scope
        //Must be use with ControllerAs syntax in view
        mainVm = this; // $Scope aliase
        mainVm.timerObj;

        //General Methods

        mainVm.startTimer = startTimer;
        mainVm.stopTimer = stopTimer;


        // General Functions for Timer
        function start(duration, display) {
            var timer = duration, minutes, seconds;

            mainVm.timerObj = setInterval(function () {
                minutes = parseInt(timer / 60, 10);
                seconds = parseInt(timer % 60, 10);

                minutes = minutes < 10 ? "0" + minutes : minutes;
                seconds = seconds < 10 ? "0" + seconds : seconds;

                display.textContent = minutes + ":" + seconds;

                if (--timer < 0) {
                    mainVm.stopTimer();
                    $scope.$broadcast ('endQuiz');
                }
            }, 1000);
        }

        function startTimer(totalTime) {
            var minute = parseInt(totalTime.split("m")[0]);

            display = document.querySelector('#time');
            start(minute * 60, display);
        };

        function stopTimer() {
            clearInterval(mainVm.timerObj);
        }


    }


})();
