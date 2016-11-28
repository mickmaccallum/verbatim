$(function() {
  $('#logout-button').click(function(event) {
    if (!confirm('Are you sure you want to log out?')) {
      return;
    }

    $(this).closest('form').submit();
  });
});

function alertError(error) {
  alert('Error: ' + error);
};

function alertAjaxFailure(xhr, status, error) {
  var message = '';
  if (xhr.responseText != null) {
    message = xhr.responseText;
  } else {
    message = error;
  }

  alertError(message);
};
