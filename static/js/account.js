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

function validatePasswords(password, confirm) {
  if (password !== confirm) {
    return false;
  }

  return password.length >= 8 && password.length <= 255
};

function addPasswordChangeListener() {
  $('#submit-password-change').click(function(event) {
    var passwordField = $('#admin-form-password');
    var confirmPasswordField = $('#admin-form-confirm-password');

    if (!validatePasswords(passwordField.val(), confirmPasswordField.val())) {
      console.log('password isn\'t valid.');
      return;
    }

    $.ajax({
      url: '/account/password',
      type: 'POST',
      data: $('#admin-password-form').serialize(),
    }).done(function(response) {
      passwordField.val('');
      confirmPasswordField.val('');
    }).fail(function() {
      console.log("error");
      console.log(this);
    });
  });	
};

function addDeleteAdminListener() {
  
};

$(function () {
  addHandleChangeListener();
  addPasswordChangeListener();
  addDeleteAdminListener();
});