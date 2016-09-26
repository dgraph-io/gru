// MATERIAL DESIGN SNACKBAR
(function(){
    var snackbarContainer = document.querySelector('#snackbar-container');
  window.SNACKBAR = function(setting) {
    if(setting.messageType == "error") {
        $(snackbarContainer).addClass("bg-fr-red");
    } else {
        $(snackbarContainer).removeClass("bg-fr-red");
    }

    var data = {
      message: setting.message,
    };

    if(setting.timeout){
        data.timeout = setting.timeout;
    } else {
        data.timeout = 2000;
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
})();