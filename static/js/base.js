$(function() {
  $('#logout-button').click(function(event) {
    $.ajax({
      url: '/logout',
      type: 'POST',
    });
  });
});