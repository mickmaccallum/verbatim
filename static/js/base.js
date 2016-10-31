$(function() {
  $('#logout-button').click(function(event) {
  	$(this).closest('form').submit();
  });
});