$(function() {
  $('#logout-button').click(function(event) {
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
