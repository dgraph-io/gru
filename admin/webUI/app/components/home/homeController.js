angular.module('GruiApp').controller('homeController', [
  "$scope",
  "$rootScope",
  function homeController($scope, $rootScope, questionService) {
    homeVm = this;
    mainVm.pageName = "home"

    marked.setOptions({
      renderer: new marked.Renderer(),
      gfm: true,
      tables: true,
      breaks: false,
      pedantic: false,
      sanitize: false, // if false -> allow plain old HTML ;)
      smartLists: true,
      smartypants: false,
      highlight: function(code, lang) {
        // in case, there is code without language specified
        if (lang) {
          return hljs.highlight(lang, code).value;
        } else {
          return hljs.highlightAuto(code).value;
        }
      }
    });

    mainVm.markDownFormat = function(content) {
      return marked(content || "", {
        gfm: true
      });
    }
  }
]);
