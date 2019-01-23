componentHandler.upgradeAllRegistered();
angular
  .module("GruiApp", [
    "ngRoute",
    "ui.router",
    "oc.lazyLoad",
    "door3.css",
    "ngSanitize",
    "ui.select",
    "ui.codemirror"
  ])
  .run([
    "$rootScope",
    "$state",
    "$stateParams",
    "$window",
    "$templateCache",
    function($rootScope, $state, $stateParams, $window, $templateCache) {
      // Set reference to access them from any scope
      $rootScope.SEO = {};
      $rootScope.SEO.title = "Gru";
    }
  ]);
angular.module("GruiApp").config(function(uiSelectConfig) {
  uiSelectConfig.theme = "select2";
});

angular
  .module("GruiApp")
  .run(function($rootScope, $location, $timeout, $state) {
    //Run After view has been loaded
    $rootScope.$on("$viewContentLoaded", function() {
      componentHandler.upgradeAllRegistered();
      $timeout(function() {
        componentHandler.upgradeAllRegistered();
        componentHandler.upgradeDom();
      }, 1000);
    });

    $rootScope.$on("$stateChangeStart", function(
      e,
      toState,
      toParams,
      fromState,
      fromParams
    ) {
      if (toState.authenticate || toState.name == "login") {
        mainVm.base_url = "/api/admin";
      } else {
        mainVm.base_url = "/api";
      }
      if (toState.authenticate && !mainVm.isLoggedIn()) {
        $state.transitionTo("login");
      } else if (toState.name == "login" && mainVm.isLoggedIn()) {
        $state.transitionTo("root");
      }

      setTimeout(function() {
        $mdl_input = $(".mdl-textfield__input");
        for (var i = 0; i < $mdl_input.length; i++) {
          var this_field = $($mdl_input[i]);
          this_field.removeClass("is-invalid");

          if (this_field.attr("type") == "date") {
            this_field.parent().addClass("is-focused");
          }
        }
      }, 700);
    });

    $rootScope.upgradeMDL = function() {
      $timeout(function() {
        componentHandler.upgradeAllRegistered();
      }, 1000);
    };

    //Run After ng-include has been loaded
    $rootScope.$on("$includeContentLoaded", function(event, templateName) {
      componentHandler.upgradeAllRegistered();
    });
  });

// LAZY LOAD CONFIGURATION
angular.module("GruiApp").config([
  "$ocLazyLoadProvider",
  "$httpProvider",
  "APP_REQUIRES",
  function($ocLazyLoadProvider, $httpProvider, APP_REQUIRES) {
    "use strict";
    $httpProvider.defaults.useXDomain = true;
    delete $httpProvider.defaults.headers.common["X-Requested-With"];

    // Lazy Load modules configuration
    $ocLazyLoadProvider.config({
      debug: false,
      events: true,
      modules: APP_REQUIRES.modules
    });
  }
]);

// SCRIPT NAME CONFIG For OCLAZYLOAD
angular.module("GruiApp").constant("APP_REQUIRES", {
  scripts: {
    homeController: ["app/components/home/homeController.js"],
    loginController: ["app/components/login/loginController.js"],
    loginService: ["app/components/login/loginService.js"],
    questionController: ["app/components/question/questionController.js"],
    addQuestionController: ["app/components/question/addQuestionController.js"],
    editQuestionController: ["app/components/question/editQuestionController.js"],
    questionService: ["app/components/question/questionService.js"],
    quizController: ["app/components/quiz/quizController.js"],
    quizServices: ["app/components/quiz/quizServices.js"],
    inviteController: [
      "app/components/invite/inviteController.js"
    ],
    inviteService: ["app/components/invite/inviteService.js"],
    quizLandingController: [
      "app/components/candidate/quizLandingController.js"
    ],
    quizLandingService: [
      "app/components/candidate/quizLandingService.js"
    ],
    candidateController: [
      "app/components/candidate/candidateController.js"
    ],
    candidateService: ["app/components/candidate/candidateService.js"],
    profileController: [
      "app/components/profile/profileController.js"
    ],
    profileService: ["app/components/profile/profileService.js"],
    "angular-select": ["assets/lib/js/angular-select.min.js"],
    codeMirror: ["assets/lib/js/codemirror.js"],
    javascript: ["assets/lib/js/javascript.js"],
    marked: ["assets/lib/js/marked.min.js"],
    highlight: ["assets/lib/js/highlight.pack.js"]
  }
});

angular.module("GruiApp").provider("RouteHelpers", [
  "APP_REQUIRES",
  function(appRequires) {
    "use strict";
    // Generates a resolve object by passing script names
    // previously configured in constant.APP_REQUIRES
    this.resolveFor = function() {
      var _args = arguments;
      return {
        deps: [
          "$ocLazyLoad",
          "$q",
          function($ocLL, $q) {
            // Creates a promise chain for each argument
            var promise = $q.when(1); // empty promise
            for (var i = 0, len = _args.length; i < len; i++) {
              promise = andThen(_args[i]);
            }
            return promise;

            // creates promise to chain dynamically
            function andThen(_arg) {
              // also support a function that returns a promise
              if (typeof _arg == "function") {
                return promise.then(_arg);
              } else {
                return promise.then(function() {
                  // if is a module, pass the name. If not, pass the array
                  var whatToLoad = getRequired(_arg);
                  // simple error check
                  if (!whatToLoad)
                    return $.error(
                      "Route resolve: Bad resource name [" + _arg + "]"
                    );
                  // finally, return a promise
                  return $ocLL.load(whatToLoad);
                });
              }
            }

            function getRequired(name) {
              if (appRequires.modules) {
                for (var m in appRequires.modules) {
                  if (appRequires.modules[m].name === name) {
                    return appRequires.modules[m];
                  }
                }
              }
              return appRequires.scripts && appRequires.scripts[name];
            }
          }
        ]
      };
    }; // resolveFor

    // not necessary, only used in config block for routes
    this.$get = function() {};
  }
]);

angular.module("GruiApp").controller("MainController", [
  "$scope",
  "$rootScope",
  "$state",
  "$stateParams",
  "$sce",
  "$parse",
  "$http",
  function MainController(
    $scope,
    $rootScope,
    $state,
    $stateParams,
    $sce,
    $parse,
    $http,
  ) {
    mainVm = this;
    mainVm.timerObj;
    mainVm.admin_url = "/api/admin";
    mainVm.candidate_url = "/api";
    mainVm.showModal = false;

    mainVm.getNumber = function(num) {
      return new Array(num);
    };

    mainVm.markDownFormat = function(content) {
      return marked(content || "", {
        gfm: true
      });
    }

    mainVm.indexOfObject = indexOfObject;
    mainVm.isObject = isObject;
    mainVm.goTo = function(state, data) {
      $state.transitionTo(state, data);
    }
    mainVm.objLen = function(object) {
      return !object ? 0 : Object.keys(object).length;
    }

    mainVm.isLoggedIn = function isLoggedIn() {
      return !!localStorage.getItem("token")
    }

    mainVm.logout = function logout() {
      localStorage.removeItem("token");
      $state.transitionTo("login");
    }

    mainVm.parseGoTime = parseGoTime;
    mainVm.openModal = openModal;
    mainVm.hideModal = hideModal;
    mainVm.timeoutModal = timeoutModal;
    mainVm.initNotification = initNotification;
    mainVm.hideNotification = hideNotification;

    function indexOfObject(arr, obj) {
      if (!arr) {
        return -1;
      }
      for (var i = 0; i < arr.length; i++) {
        if (angular.equals(arr[i], obj)) {
          return i;
        }
      }
      return -1;
    }

    function openModal(setting) {
      if (!setting.template) {
        return;
      }
      mainVm.modal = {};

      // CHECK IF TEMPLATE IS STRING OR URL
      mainVm.modal.isString = !!setting.isString;
      if (mainVm.modal.isString) {
        $(".modal-wrapper").html(
          $parse($sce.trustAsHtml(setting.template))($scope)
        );
      }

      mainVm.modal.template = setting.template;
      mainVm.modal.class = setting.class || "";
      mainVm.modal.hideClose = setting.hideClose || false;
      mainVm.modal.showYes = setting.showYes || false;

      mainVm.showModal = true;
    }

    function hideModal() {
      mainVm.modal = {};
      mainVm.showModal = false;
    }

    function timeoutModal() {
      var modalContent =
        "We are sorry but we are facing some problems.";
      modalContent +=
        "Please send us a email on <a href='mailto:contact@dgraph.io'>contact@dgraph.io</a>";
      mainVm.openModal({
        template: modalContent,
        isString: true
      });
      $rootScope.$broadcast("endQuiz", {
        message: modalContent
      });
    }

    function isObject(obj) {
      return Object.prototype.toString.call(obj) == "[object Object]";
    }

    function parseGoTime(time) {
      var duration = Duration.parse(time);
      var totalSec = duration.seconds();

      return {
        minutes:
          Math.floor(totalSec / 3600) * 60 + parseInt((totalSec / 60) % 60, 10),
        seconds: parseInt(totalSec % 60, 10)
      };
    }

    function initNotification(message) {
      mainVm.consecutiveError = mainVm.consecutiveError || 0;
      mainVm.consecutiveError += 1;

      if (mainVm.consecutiveError < 2) {
        return;
      }
      if (mainVm.consecutiveError >= 20) {
        mainVm.timeoutModal();
      }
      mainVm.notification = {};
      mainVm.showNotification = true;

      mainVm.notification.class = "notification-error";
      if (message) {
        mainVm.notification.message = message;
        mainVm.consecutiveError = 1;
      } else {
        mainVm.notification.message =
          "Can't connect to the server. We will be back up in a bit...";
      }
    }

    function hideNotification(rejected) {
      if (!rejected) {
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
  },
]);


angular.module("GruiApp").service("MainService", [
  "$http",
  "$state",
  "$location",
  function MainService($http, $state, $location) {
    function setAuth(auth) {
      $http.defaults.headers.common["Authorization"] = auth;
    }

    function redirectIfUnautorized(response) {
      if (response.status !== 401) {
        return;
      }
      localStorage.removeItem("token");
      SNACKBAR({
        message: "You must login to access"
      });
      $state.transitionTo("login");
    }

    return mainService = {
      dgraphSuccess: "Success",
      proxy: function proxy(data) {
        return mainService.post("/proxy", data);
      },
      mutateProxy: function mutateProxy(data) {
        return mainService.post("/mutateProxy", data);
      },
      post: function post(url, data, hideLoader) {
        var req = {
          method: "POST",
          url: mainVm.base_url + url,
          data: data,
          timeout: 30000
        };

        if (url == "/login") {
          $http.defaults.headers.common["Authorization"] =
            "Basic " + btoa(data.user + ":" + data.password);
          delete req.data;
        } else {
          var candidateInfo = JSON.parse(localStorage.getItem("candidate_info"));
          if (mainVm.base_url.indexOf("admin") == -1 &&
            candidateInfo && candidateInfo.token
          ) {
            setAuth("Bearer " + candidateInfo.token);
          } else {
            setAuth("Bearer " + localStorage.getItem("token"));
          }
        }

        if (!hideLoader) {
          mainVm.showAjaxLoader = true;
        }
        return $http(req).then(
          function(data) {
            mainVm.showAjaxLoader = false;
            return data.data;
          },
          function(response, code) {
            mainVm.showAjaxLoader = false;
            // TODO - Remove this dirty edge case handling for login. We have it because otherwise if you login
            // with incorrect creds, it show You must login and then shows the actual error.
            url !== "/login" && redirectIfUnautorized(response);
            throw response;
          }
        );
      },
      get: function get(url) {
        var req = {
          method: "GET",
          url: mainVm.base_url + url
        };
        var candidateInfo = JSON.parse(localStorage.getItem("candidate_info"));
        if (mainVm.base_url.indexOf("admin") == -1 &&
          candidateInfo && candidateInfo.token
        ) {
          setAuth("Bearer " + candidateInfo.token);
        } else {
          setAuth("Bearer " + localStorage.getItem("token"));
        }

        mainVm.showAjaxLoader = true;
        return $http(req).then(
          function(data) {
            mainVm.showAjaxLoader = false;
            return data.data;
          },
          function(response) {
            mainVm.showAjaxLoader = false;
            redirectIfUnautorized(response);
            throw response;
          }
        );
      },
      put: function put(url, data) {
        setAuth("Bearer " + localStorage.getItem("token"));
        mainVm.showAjaxLoader = true;
        return $http({
          method: "PUT",
          url: mainVm.admin_url + url,
          data: data
        }).then(
          function(data) {
            mainVm.showAjaxLoader = false;
            return data.data;
          },
          function(response) {
            mainVm.showAjaxLoader = false;
            redirectIfUnautorized(response);
            throw response;
          }
        );
      },
    };
  },
]);
