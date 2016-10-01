// MATERIAL DESIGN SNACKBAR
(function(){
  
  $(document).ready(function() {
    componentHandler.upgradeAllRegistered();

    (function(){
      $mdl_input = $(".mdl-textfield__input")
      for(var i=0; i < $mdl_input.length; i++) {
        var this_field = $mdl_input[i];
        this_field.removeClass("is-invalid");
      }
    })();

    $(document).on("click", ".slide-wrapper .slide-link", function(){
        $this = $(this);
        if(!$this.hasClass("is-active")) { 
          $(".slide-content:visible").slideUp( "swing", function() {
            // Animation complete.
          });
        };
        $parent = $this.closest(".slide-wrapper");

        $thisContent = $(".slide-content", $parent);

        $thisContent.stop(true, true).slideToggle( "swing", function() {
            // Animation complete.
            $(".slide-link").removeClass("is-active");
            if($thisContent.is(':visible')){
              $this.addClass("is-active");
            }
        });
    });
    
    var snackbarContainer = document.querySelector('#snackbar-container');
    window.SNACKBAR = function(setting) {
      if(setting.messageType == "error") {
          $(snackbarContainer).addClass("error");
      } else {
          $(snackbarContainer).removeClass("success");
      }

      var data = {
        message: setting.message,
      };

      if(setting.timeout){
          data.timeout = setting.timeout;
      } else {
          data.timeout = 3000;
      }
      snackbarContainer.MaterialSnackbar.showSnackbar(data);
    }

    $(document).on("click", ".reload-same-url", function(){
      var href = this.href
      window.location = window.location.href.split("?")[0];
      setTimeout(function() {
        window.location.reload();
      }, 10);
    });
  })  
})();