function addHandleChangeListener() {
  $('#submit-handle-change').click(function(event) {

    var data = $('#admin-handle-form').serialize();
    console.log($('#admin-handle-form'));
    console.log(data);

    $.ajax({
      url: '/account/handle',
      type: 'POST',
      dataType: 'json',
      // contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
      data: data,
    }).done(function(network) {
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