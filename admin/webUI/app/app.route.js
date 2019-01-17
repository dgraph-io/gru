angular.module("GruiRoutes", ["GruiApp"]);

var homeTemplate = "app/components/home/home.html";
var indexTemplate = "app/index.html";
var loginTemplate = "app/components/login/index.html";

var questionTemplate = "app/components/question/index.html";
var allQuestionTemplate = "app/components/question/all-question.html";
var addQuestionTemplate = "app/components/question/add-question.html";
var editQuestionTemplate = "app/components/question/edit-question.html";

var quizTemplate = "app/components/quiz/index.html";
var addQuizTemplate = "app/components/quiz/add-quiz.html";
var editQuizTemplate = "app/components/quiz/edit-quiz.html";
var allQuizTemplate = "app/components/quiz/all-quiz.html";

var inviteTemplate = "app/components/invite/index.html";
var inviteDashboardTemplate =
  "app/components/invite/views/invite-dashboard.html";
var inviteUserTemplate = "app/components/invite/views/invite-user.html";
var editInviteTemplate = "app/components/invite/views/edit-invite.html";
var candidateReportTemplate = "app/components/invite/views/candidate-report.html";

var candidateIndexTemplate = "app/components/candidate/index.html";
var candidateLandingTemplate = "app/components/candidate/views/landing.html";
var candidateQuizTemplate = "app/components/candidate/views/quiz.html";

var quizLandingTemplate = "app/components/candidate/views/quiz-landing.html";

var profileIndexTemplate = "app/components/profile/index.html";
var editProfileTemplate = "app/components/profile/edit.html";

// CSS for View/Directives
var angularSelectCSS = "assets/lib/css/angular-select.min.css";
var codeMirrorCSS = "assets/lib/css/codemirror.css";
var githubCSS = "assets/lib/css/github.css";

angular.module("GruiRoutes").config([
  "$stateProvider",
  "$locationProvider",
  "$urlRouterProvider",
  "RouteHelpersProvider",
  function MainRoutes(
    $stateProvider,
    $locationProvider,
    $urlRouterProvider,
    helper
  ) {
    // Set the following to true to enable the HTML5 Mode
    // You may have to set <base> tag in index and a routing configuration in your server
    $locationProvider.html5Mode(false);
    // $locationProvider.hashPrefix('!');

    // default route
    $urlRouterProvider.otherwise("/login");

    $stateProvider
      .state("root", {
        url: "/admin",
        templateUrl: homeTemplate,
        resolve: helper.resolveFor("homeController", "marked"),
        authenticate: true
      })
      .state("login", {
        url: "/login",
        templateUrl: loginTemplate,
        resolve: helper.resolveFor("loginController", "loginService")
      })
      .state("question", {
        url: "/admin/question",
        abstract: true,
        templateUrl: questionTemplate,
        css: [angularSelectCSS, githubCSS],
        resolve: helper.resolveFor(
          "questionController",
          "questionService",
          "angular-select",
          "codeMirror",
          "javascript",
          "marked",
          "highlight"
        )
      })
      .state("question.all", {
        url: "/all-questions",
        parent: "question",
        templateUrl: allQuestionTemplate,
        css: [angularSelectCSS, githubCSS],
        params: { quesID: null },
        authenticate: true
      })
      .state("question.add", {
        url: "/add-question",
        parent: "question",
        templateUrl: addQuestionTemplate,
        css: [angularSelectCSS, codeMirrorCSS, githubCSS],
        resolve: helper.resolveFor(
          "addQuestionController",
        ),
        authenticate: true
      })
      .state("question.edit", {
        url: "/edit-question/:quesID/returnQuizId/:returnQuizId",
        parent: "question",
        templateUrl: editQuestionTemplate,
        css: [angularSelectCSS, codeMirrorCSS, githubCSS],
        resolve: helper.resolveFor(
          "editQuestionController",
        ),
        authenticate: true
      })
      .state("quiz", {
        url: "/admin/quiz",
        abstract: true,
        templateUrl: quizTemplate,
        resolve: helper.resolveFor(
          "quizController",
          "quizServices",
          "questionController",
          "questionService",
        )
      })
      .state("quiz.all", {
        url: "/all-quiz",
        parent: "quiz",
        templateUrl: allQuizTemplate,
        authenticate: true
      })
      .state("quiz.add", {
        url: "/add-quiz?:index?:qid",
        parent: "quiz",
        templateUrl: addQuizTemplate,
        authenticate: true
      })
      .state("quiz.edit", {
        url: "/edit-quiz/:quizID",
        parent: "quiz",
        templateUrl: editQuizTemplate,
        authenticate: true
      })
      .state("invite", {
        url: "/admin/invite",
        abstract: true,
        templateUrl: inviteTemplate,
        resolve: helper.resolveFor(
          "inviteController",
          "quizServices",
          "inviteService",
          "marked",
          "highlight"
        )
      })
      .state("invite.dashboard", {
        url: "/dashboard/:quizID",
        parent: "invite",
        templateUrl: inviteDashboardTemplate,
        authenticate: true
      })
      .state("invite.add", {
        url: "/invite-user",
        parent: "invite",
        css: [angularSelectCSS],
        templateUrl: inviteUserTemplate,
        params: {
          quizID: null
        },
        authenticate: true
      })
      .state("invite.edit", {
        url: "/edit-invite/:quizID/:candidateID",
        parent: "invite",
        css: [angularSelectCSS],
        templateUrl: editInviteTemplate,
        authenticate: true
      })
      .state("invite.report", {
        url: "/candidate-report/:candidateID",
        parent: "invite",
        css: [githubCSS],
        templateUrl: candidateReportTemplate,
        authenticate: true
      })
      .state("quiz-landing", {
        url: "/quiz/:quiz_token",
        css: [angularSelectCSS],
        templateUrl: quizLandingTemplate,
        resolve: helper.resolveFor(
          "quizLandingController",
          "quizLandingService"
        )
      })
      .state("candidate", {
        url: "/candidate",
        abstract: true,
        templateUrl: candidateIndexTemplate,
        resolve: helper.resolveFor(
          "candidateController",
          "candidateService",
          "marked",
          "highlight"
        )
      })
      .state("candidate.landing", {
        url: "/home",
        parent: "candidate",
        templateUrl: candidateLandingTemplate
      })
      .state("candidate.quiz", {
        url: "/quiz/:quiz_token",
        parent: "candidate",
        css: [githubCSS],
        templateUrl: candidateQuizTemplate
      })
      .state("profile", {
        url: "/admin/profile",
        abstract: true,
        templateUrl: profileIndexTemplate,
        resolve: helper.resolveFor(
          "profileController",
          "profileService",
          "codeMirror",
          "marked"
        )
      })
      .state("profile.edit", {
        url: "/edit",
        authenticate: true,
        parent: "profile",
        css: [codeMirrorCSS, githubCSS],
        templateUrl: editProfileTemplate
      });
  }
]);
