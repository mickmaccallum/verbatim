function addHandleChangeListener() {
  $('#submit-handle-change').click(function(event) {
    $.ajax({
      url: '/account/handle',
      type: 'POST',
      data: $('#admin-handle-form').serialize(),
    }).done(function(response) {
      // $('#network-form-port').val('');
      // $('#network-form-name').val('');
    }).fail(alertAjaxFailure);
  });
};

function validatePasswords(password, confirm) {
  if (password !== confirm) {
    return false;
  }

  if (password == null || password == undefined) {
    return false;
  }

  return password.length >= 8 && password.length <= 255
};

function validateNewPasswordForm() {
  var errors = [];
  var oldPasswordField = $('#admin-form-old-password');
  var newpasswordField = $('#admin-form-password');
  var confirmNewPasswordField = $('#admin-form-confirm-password');

  if (oldPasswordField.val() == null || oldPasswordField.val().length === 0) {
    errors.push('Old password is missing');
  } else if (oldPasswordField.val().length > 255) {
    errors.push('Old password is too long. Must be less than 255 characters');
  }

  var newPassword = newpasswordField.val();
  var confirmedNew = confirmNewPasswordField.val();

  if () {

  }
};

function addPasswordChangeListener() {
  $('#submit-password-change').click(function(event) {
    var oldPasswordField = $('#admin-form-old-password');
    var passwordField = $('#admin-form-password');
    var confirmPasswordField = $('#admin-form-confirm-password');
    console.log('++++++++++++++++++');
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
    }).fail(alertAjaxFailure);
  });	
};

function addDeleteAdminListener() {
  $('.delete-button').click(function(event) {
    event.stopPropagation();
    event.preventDefault();

    var row = $(this).closest('tr');
    var adminId = row.attr('data-admin-id');
    var adminHandle = row.attr('data-admin-handle');

    if (!confirm('Are you sure you want to delete the admin: ' + adminHandle)) {
      return;
    }
    console.log($('#delete-admin-form').serialize());
    $.ajax({
      url: '/account/delete/' + adminId,
      type: 'POST',
      data: $('#delete-admin-form').serialize()
    }).done(function() {
      var count = $('#admin-list-wrapper > table > tbody').children('tr').count;
      if (count <= 1) {
        $('#admin-list-wrapper').hide('400', function() {
          row.remove();          
        });
      }
    }).fail(function() {
      alert("Failed to remove network from list.");
    });
  });
};

function addAddAdminListener() { // that's hard to say...
  $('#submit-add-admin').click(function(event) {
    event.preventDefault();

    $.ajax({
      url: '/account/add',
      type: 'POST',
      dataType: 'json',
      data: $('#add-admin-form').serialize()
    }).done(function(response) {
      console.log(response);
      $('#add-admin-form-handle').val('');
      $('#add-admin-form-password').val('');
      $('#add-admin-form-confirm-password').val('');

      var count = $('#admin-list-wrapper > table > tbody').children('tr').count;
      if (condition) {

      }
    }).fail(function() {
      console.log("error");
    });    
  });
};

$(function () {
  addHandleChangeListener();
  addPasswordChangeListener();
  addDeleteAdminListener();
  addAddAdminListener();
});