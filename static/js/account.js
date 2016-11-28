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
  var errors = [];

  if (password == null || confirm == null) {
    errors.push('Missing password');
  } else {
    if (password === confirm) {
      if (password.length === 0) {
        errors.push('Missing password');
      } else if (password.length < 8) {
        errors.push('New password is too short. Must be at least 8 characters');
      } else if (password.length > 255) {
        errors.push('New password is too long. Must be less than 256 characters');
      }
    } else {
      errors.push('New passwords do not match');
    }
  }

  return errors;
};

function validateNewAdminForm() {
  var errors = [];
  var handleField = $('#add-admin-form-handle');
  var passwordField = $('#add-admin-form-password');
  var confirmPasswordField = $('#add-admin-form-confirm-password');

  if (handleField.val() == null || handleField.val().length === 0) {
    errors.push('Handle is missing');
  } else if (handleField.val().length > 255) {
    errors.push('Handle is too long. Must be less than 256 characters');
  }

  errors = errors.concat(
    validatePasswords(
      passwordField.val(), 
      confirmPasswordField.val()
    )
  );

  return errors;
};

function validateNewPasswordForm() {
  var errors = [];
  var oldPasswordField = $('#admin-form-old-password');
  var newpasswordField = $('#admin-form-password');
  var confirmNewPasswordField = $('#admin-form-confirm-password');

  if (oldPasswordField.val() == null || oldPasswordField.val().length === 0) {
    errors.push('Old password is missing');
  } else if (oldPasswordField.val().length > 255) {
    errors.push('Old password is too long. Must be less than 256 characters');
  }

  errors = errors.concat(
    validatePasswords(
      newpasswordField.val(), 
      confirmNewPasswordField.val()
    )
  );

  return errors;
};

function displayErrorsOnContainer(errors, container) {
  container.text(errors.join(',\t\t'));
  if (container.is(':hidden')) {
    container.show('fast');
  }
};

function hideErrorContainer(container) {
  if (!container.is(':hidden')) {
    container.hide('fast');
  }
};

function addPasswordChangeListener() {
  $('#submit-password-change').click(function(event) {
    event.preventDefault();
    var passwordErrors = validateNewPasswordForm();
    var container = $('#change-password-form-error-container');

    if (passwordErrors.length > 0) {
      displayErrorsOnContainer(passwordErrors, container);
      return;
    }

    hideErrorContainer(container);
    var form = $('#admin-password-form');

    $.ajax({
      url: '/account/password',
      type: 'POST',
      data: form.serialize(),
    }).done(function(response) {
      form.find('input.form-control').val('');
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
    }).fail(alertAjaxFailure);
  });
};

function addAddAdminListener() { // that's hard to say...
  $('#submit-add-admin').click(function(event) {
    event.preventDefault();

    var adminErrors = validateNewAdminForm();
    var container = $('#add-admin-form-error-container');

    if (adminErrors.length > 0) {
      displayErrorsOnContainer(adminErrors, container);
      return;
    }

    hideErrorContainer(container);
    var form = $('#add-admin-form');

    $.ajax({
      url: '/account/add',
      type: 'POST',
      dataType: 'json',
      data: form.serialize()
    }).done(function(admin) {
      form.find('input.form-control').val('');
      addAdmin(admin);
    }).fail(alertAjaxFailure);    
  });
};

$(function () {
  addHandleChangeListener();
  addPasswordChangeListener();
  addDeleteAdminListener();
  addAddAdminListener();
});