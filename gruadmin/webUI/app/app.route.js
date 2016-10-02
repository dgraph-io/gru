//Angular Routes defined here.

// Setting Module
(function(){
    angular.module('GruiRoutes', ['GruiApp',]);
    
    // Template, dirctives, js/css urls

    var  homeTemplate = 'app/components/home/home.html';
    var  indexTemplate = 'app/index.html';
    var  loginTemplate = 'app/components/login/index.html';

    var  questionTemplate = 'app/components/question/index.html';
    var  allQuestionTemplate = 'app/components/question/all-question.html';
    var  addQuestionTemplate = 'app/components/question/add-question.html';

    var  quizTemplate = 'app/components/quiz/index.html';
    var  allQuizTemplate = 'app/components/quiz/all-quiz.html';
    var  addQuizTemplate = 'app/components/quiz/add-quiz.html';

    var  inviteTemplate = 'app/components/invite/index.html';
    var  inviteDashboardTemplate = 'app/components/invite/views/invite-dashboard.html';
    var  inviteUserTemplate = 'app/components/invite/views/invite-user.html';
    var  editInviteTemplate = 'app/components/invite/views/edit-invite.html';

    // CSS for View/Directives
    var select2CSS = "assets/lib/css/select2.min.css";
    var angularSelectCSS = "assets/lib/css/angular-select.min.css";
    var codeMirrorCSS = "assets/lib/css/codemirror.css";

    function MainRoutes($stateProvider, $locationProvider, $urlRouterProvider, helper) {
      'use strict';

      // Set the following to true to enable the HTML5 Mode
      // You may have to set <base> tag in index and a routing configuration in your server
      $locationProvider.html5Mode(false);
      // $locationProvider.hashPrefix('!');

      // default route
      $urlRouterProvider.otherwise('/');

      // --------------Application Routes---------------
      $stateProvider
        .state('root', {
          url: '/',
          templateUrl: homeTemplate,
          resolve: helper.resolveFor('homeController'),
          authenticate: true,
        })
        .state('login', {
          url: '/login',
          templateUrl: loginTemplate,
          resolve: helper.resolveFor('loginController', 'loginService'),
        })
        .state('question', {
          url: '/admin/question',
          abstract: true,
          templateUrl: questionTemplate,
          css: [angularSelectCSS],
          resolve: helper.resolveFor('questionController', 'questionServices', 'angular-select', 'codeMirror', 'javascript'),
        })
          .state('question.all', {
            url: '/all-questions',
            parent: 'question',
            templateUrl: allQuestionTemplate,
            css: [angularSelectCSS],
            authenticate: true,
          })
          .state('question.add', {
            url: '/add-question',
            parent: 'question',
            templateUrl: addQuestionTemplate,
            css: [angularSelectCSS, codeMirrorCSS],
            authenticate: true,
          })
        .state('quiz', {
          url: '/admin/quiz',
          abstract: true,
          templateUrl: quizTemplate,
          resolve: helper.resolveFor('quizController', 'quizServices', 'questionServices'),
        })
          .state('quiz.all', {
            url: '/all-quiz',
            parent: 'quiz',
            templateUrl: allQuizTemplate,
            authenticate: true,
          })
          .state('quiz.add', {
            url: '/add-quiz?:index?:qid',
            parent: 'quiz',
            templateUrl: addQuizTemplate,
            authenticate: true,
          })
        .state('invite', {
          url: '/admin/invite',
          abstract: true,
          templateUrl: inviteTemplate,
          resolve: helper.resolveFor('inviteController', 'quizServices', 'inviteService'),
        })
          .state('invite.dashboard', {
            url: '/dashboard/:quizID',
            parent: 'invite',
            templateUrl: inviteDashboardTemplate,
            authenticate: true,
          })
          .state('invite.add', {
            url: '/invite-user',
            parent: 'invite',
            templateUrl: inviteUserTemplate,
            authenticate: true,
          })
          .state('invite.edit', {
            url: '/edit-invite/:quizID/:candidateID',
            parent: 'invite',
            templateUrl: editInviteTemplate,
            authenticate: true,
          })

    }

    // Dependency and rout function array
    var GruiRoutes = [
      '$stateProvider', 
      '$locationProvider', 
      '$urlRouterProvider',
      'RouteHelpersProvider',
      MainRoutes,
    ]

    // Getting module and setting routes
    angular.module('GruiRoutes').config(GruiRoutes);
})();
