(function() {

  $(document).ready(function() {
    componentHandler.upgradeAllRegistered();

    window.formatDate = formatDate;
    window.getDocHeight = getDocHeight;
    window.scrollTo = scrollTo;
    window.scrollNavInit = scrollNavInit;


    function getDocHeight(pageID) {
      var D = document;
      return Math.max(
        document.getElementById(pageID).scrollHeight,
        document.getElementById(pageID).offsetHeight,
        document.getElementById(pageID).clientHeight
      );
    }

    (function() {
      setTimeout(function() {
        $mdl_input = $(".mdl-textfield__input")
        for (var i = 0; i < $mdl_input.length; i++) {
          var this_field = $($mdl_input[i]);
          this_field.removeClass("is-invalid");

          if (this_field.attr('type') == "date") {
            this_field.parent().addClass("is-focused");
          }
        }
      }, 1000);
    })();

    window.getDate = function(date) {
      var now = new Date(date);

      var day = ("0" + now.getDate()).slice(-2);
      var month = ("0" + (now.getMonth() + 1)).slice(-2);

      var today = now.getFullYear() + "-/" + (month) + "/" + (day);


      return today;
    }

    window.isValidEmail = function(email) {
      var re;
      re = /^([\w-]+(?:\.[\w-]+)*)@((?:[\w-]+\.)*\w[\w-]{0,66})\.([a-z]{2,6}(?:\.[a-z]{2})?)$/i;
      return re.test(email);
    };

    window.lsSupported = (function() {
      return (typeof Storage !== "undefined") ? true : false;
    })();

    Date.prototype.toDateInputValue = (function() {
      var local = new Date(this);
      local.setMinutes(this.getMinutes() - this.getTimezoneOffset());
      return local.toJSON().slice(0, 10);
    });

    function formatDate(date) {
      var d = new Date(date),
        month = '' + (d.getMonth() + 1),
        day = '' + d.getDate(),
        year = d.getFullYear();

      if (month.length < 2) month = '0' + month;
      if (day.length < 2) day = '0' + day;

      return [year, month, day].join('-');
    }

    $(document).on("click", ".slide-wrapper .slide-link", function() {
      $this = $(this);
      if (!$this.hasClass("is-active")) {
        $(".slide-content:visible").slideUp("swing", function() {
          // Animation complete.
        });
      };
      $parent = $this.closest(".slide-wrapper");

      $thisContent = $(".slide-content", $parent);

      $thisContent.stop(true, true).slideToggle("swing", function() {
        // Animation complete.
        $(".slide-link").removeClass("is-active");
        if ($thisContent.is(':visible')) {
          $this.addClass("is-active");
        }
      });
    });

    // So that clicking on the question name in the side nav should bring the
    // question description for this question to the top.
    $(document).on("click", ".qn-nav .side-tabs", function() {
      $container = $(".qn-desc")
      $qnDesc = $("#tab0-panel")
      $container.scrollTop(
        $qnDesc.offset().top - $container.offset().top + $container.scrollTop()
      );
    });


    // General Function for Scroll to particular element

    function scrollTo(selector) {
      $container = $(".mdl-layout__content")

      $container.scrollTop(
        $(selector).offset().top - $container.offset().top + $container.scrollTop()
      );
    }

    // Function for Scroll to particular element, based on data-target

    function scrollNavInit() {
      $headerHeight = $(".mdl-layout__header-row").height();
      var sidebarLinks = $(".sticky-sidebar-link"); // find the sidebar link
      if (sidebarLinks) {
        var aArray = []; // create the empty aArray
        var sidebarLen = sidebarLinks.length;
        for (var i = 0; i < sidebarLen; i++) {
          var aChild = sidebarLinks[i];
          var ahref = $(aChild).data('scrollto');
          aArray.push(ahref);
        }
      }

      $window = $(window);
      $(".mdl-layout__content").scroll(function() {
        var windowPos = $window.scrollTop(); // get the offset of the window from the top of page
        var windowHeight = $window.height(); // get the height of the window
        var docHeight = $(document).height();

        for (var i = 0; i < aArray.length; i++) {
          var theID = aArray[i];
          var divPos = $(theID).offset().top - $headerHeight; // get the offset of the div from the top of page
          var divHeight = $(theID).height(); // get the height of the div in question
          if (windowPos >= divPos && windowPos < (divPos + divHeight)) {
            $(".sticky-sidebar-link[data-scrollto='" + theID + "']").addClass("selected-sidebar");
          } else {
            $(".sticky-sidebar-link[data-scrollto='" + theID + "']").removeClass("selected-sidebar");
          }
        }

        if (windowPos + windowHeight == docHeight) {
          if (!$(".sticky-sidebar-link:last-child a").hasClass("selected-sidebar")) {
            var navActiveCurrent = $(".selected-sidebar").attr("href");
            $(".sticky-sidebar-link[data-scrollto='" + navActiveCurrent + "']").removeClass("selected-sidebar");
            $(".sticky-sidebar-link:last-child a").addClass("selected-sidebar");
          }
        }

        var $selectedSideBar = $(".selected-sidebar");
        if ($selectedSideBar.length) {
          if (!isScrolledIntoView($selectedSideBar)) {
            $(".sticky-sidebar-content").scrollTop(
              $selectedSideBar.offset().top
            );
          }
        }
      });

      function isScrolledIntoView(elem) {
        var docViewTop = $(window).scrollTop();
        var docViewBottom = docViewTop + $(window).height();

        var elemTop = elem.offset().top;
        var elemBottom = elemTop + elem.height();

        return ((elemBottom <= docViewBottom) && (elemTop >= docViewTop));
      }

    }

    $(document).on("click", "[data-scrollto]", function() {
      $this = $(this);
      var scrollToElem = $this.data("scrollto");
      if (scrollToElem == ".mdl-layout__content") {
        $(scrollToElem).scrollTop(0);
      } else {
        scrollTo(scrollToElem);
      }
    });

    var snackbarContainer = document.querySelector('#snackbar-container');
    window.SNACKBAR = function(setting) {
      if (setting.messageType) {
        $(snackbarContainer).addClass(setting.messageType);
      } else {
        $(snackbarContainer).addClass("error");
      }

      var data = {
        message: setting.message,
      };

      if (setting.timeout) {
        data.timeout = setting.timeout;
      } else {
        data.timeout = 3000;
      }
      snackbarContainer.MaterialSnackbar.showSnackbar(data);
    }

    $(document).on("click", ".reload-same-url", function() {
      // var href = this.href
      // window.location = window.location.href.split("?")[0];
      // setTimeout(function() {
      //   window.location.reload();
      // }, 10);
    });
  })
})();
