$(function() {
  $('#logout-button').click(function(event) {
    $(this).closest('form').submit();
  });
});

function alertError(error) {
  alert('Error: ' + error);
};