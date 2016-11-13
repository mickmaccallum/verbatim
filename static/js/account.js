function addHandleChangeListener() {
  $('#submit-handle-change').click(function(event) {
    $.ajax({
      url: '/account/handle',
      type: 'POST',
      data: $('#admin-handle-form').serialize(),
    }).done(function(response) {
      // $('#network-form-port').val('');
      // $('#network-form-name').val('');
    }).fail(function() {
      console.log("error");
      console.log(this);
    });
  });
};

function addPasswordChangeListener() {
	
};

$(function () {
  addHandleChangeListener();
  addPasswordChangeListener();
});