//Angular Routes defined here.

// Setting Module
(function(){
    angular.module('GruiRoutes', ['GruiApp',]);
    
    // Template, dirctives, js/css urls

    var  homeTemplate = 'app/components/home/home.html';

    var  questionTemplate = 'app/components/question/index.html';
    var  allQuestionTemplate = 'app/components/question/all-question.html';
    var  addQuestionTemplate = 'app/components/question/add-question.html';

    var  quizTemplate = 'app/components/quiz/index.html';
    var  allQuizTemplate = 'app/components/quiz/all-quiz.html';
    var  addQuizTemplate = 'app/components/quiz/add-quiz.html';

    // CSS for View/Directives
    var select2CSS = "assets/lib/css/select2.min.css";
    var angularSelectCSS = "assets/lib/css/angular-select.min.css";

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
          })
          .state('question', {
            url: '/question',
            abstract: true,
            templateUrl: questionTemplate,
            css: [angularSelectCSS],
            resolve: helper.resolveFor('questionController', 'questionServices', 'angular-select'),
          })
            .state('question.all', {
              url: '/all-questions',
              parent: 'question',
              templateUrl: allQuestionTemplate,
              css: [angularSelectCSS],
            })
            .state('question.add', {
              url: '/add-question?:index?:qid',
              parent: 'question',
              templateUrl: addQuestionTemplate,
              css: [angularSelectCSS],
            })
          .state('quiz', {
            url: '/quiz',
            abstract: true,
            templateUrl: quizTemplate,
            resolve: helper.resolveFor('quizController', 'quizServices', 'questionServices'),
          })
            .state('quiz.all', {
              url: '/all-quiz',
              parent: 'quiz',
              templateUrl: allQuizTemplate,
            })
            .state('quiz.add', {
              url: '/add-quiz?:index?:qid',
              parent: 'quiz',
              templateUrl: addQuizTemplate,
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
