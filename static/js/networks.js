
$(function() {
  addAddEncoderHandler();
  addEditEncoderHandler();
  addDeleteEncoderHandler();
  configureEditing();
  startWebSocket();
  autoStopWebSocket();
  addMuteCaptionerListners();
  addUnmuteCaptionerListners();
});

function changeEncoderState(encoder, encoderState) {
  var row = $('.encoder-row[data-encoder-id=\'' + encoderState.encoderId + '\']');

  if (encoderState === 0) { // connected
    addEncoder(encoder);

  } else if (encoderState === 1) { // disconnected
    var wrapper = $('#encoder-list-wrapper');
    if (!wrapper.is(':hidden')) {
      $('#encoder-list-header').hide('slow');
      wrapper.animate({ height: '0px' }, 'slow', function() {
        row.remove();
      });
    }
  } else if (encoderState === 2) { // auth failure

  } else if (encoderState === 3) { // writes failing

  } else { // assuming disconnected
    // NOOP
  }
};

function encoderStateToString(state) {
  if (!Number.isInteger(state)) {
    return 'Disconnected';
  }

  if (state === 0) {
    return 'Connected';
  } else if (state === 1) {
    return 'Connecting';
  } else if (state === 2) {
    return 'Authentication Failed';
  } else if (state === 3) {
    return 'Writes Failing';
  } else {
    return 'Disconnected';
  }
};

function captionerStateToString(state) {
  if (!Number.isInteger(state)) {
    return 'Disconnected';
  }

  if (state === 0) {
    return 'Connected';
  } else if (state === 1) {
    return 'Disconnected';
  } else if (state === 2) {
    return 'Muted';
  } else if (state === 3) {
    return 'Unmuted';
  } else {
    return 'Disconnected';
  }
};

function makeMuteButton() {
  return '<p data-placement="top" data-toggle="tooltip" title="Mute">' + 
            '<button class="btn btn-danger btn-xs mute-captioner-button">' +
              '<span class="glyphicon glyphicon-volume-off"></span>' + 
            '</button>' + 
          '</p>';
};

function makeUnmuteButton() {
  return '<p data-placement="top" data-toggle="tooltip" title="Unmute">' + 
            '<button class="btn btn-danger btn-xs unmute-captioner-button">' + 
              '<span class="glyphicon glyphicon-volume-up"></span>' + 
            '</button>' + 
          '</p>';
};

function addCaptioner(captioner, tableId, state) {
  var openRow = '<tr class="captioner-row" id="' + tableId + '" ' +
    'data-captioner-ip="' + captioner.IPAddr + '" ' +
    'data-captioner-num-conn="' + captioner.NumConn + '" ' +
    'data-captioner-network-id="' + captioner.NetworkID + '">';

  var headers = '<th class="col-xl-1 col-lg-1 col-md-1 row-number" scope=row>0</th>' + 
    '<td class="col-xl-2 col-lg-2 col-md-3">' + captioner.IPAddr + '</td>' + 
    '<td class="col-xl-2 col-lg-2 col-md-3">' + captioner.NumConn + '</td>' + 
    '<td class="col-xl-2 col-lg-2 col-md-3 state-row">' + captionerStateToString(state) + '</td>';
  // state will always be 0 here. Default to disconnectable/mutable
 
  var muteColumn = '<td class="col-xl-1 col-lg-1 col-md-1 mute-row">' + makeMuteButton() + '</td>';
  var disconnect = '<td class="col-xl-1 col-lg-1 col-md-1 disconnect-row">' +
                     '<p data-placement="top" data-toggle="tooltip" title="Disconnect">' +
                       '<button class="btn btn-danger btn-xs disconnect-captioner-button">' +
                         '<span class="glyphicon glyphicon-ban-circle"></span>' +
                       '</button>' +
                     '</p>' +
                   '</td>';
 
  var endRow = '</tr>';
  var row = $(openRow + headers + muteColumn + disconnect + endRow);
  $('#captioner-selection-table > tbody').prepend(row);
  
  recountCaptioners();
  addDisconnectCaptionerListeners();
  addMuteCaptionerListners();

  var wrapper = $('#captioner-list-wrapper');
  if (wrapper.is(':hidden')) {
    $('#captioner-list-header').show('fast');
    wrapper.show('fast');
  }
};

function changeCaptionerState(captioner, state) {
  var tableId = [
    captioner.IPAddr, 
    captioner.NumConn, 
    captioner.NetworkID
  ].join(':');

  if (state === 0) { // connected
    addCaptioner(captioner, tableId, state);

  } else if (state === 1) { // disconnected
    var row = $(document.getElementById(tableId));
    row.hide('slow', function() {
      this.remove();
    });
  } else if (state === 2) { // muted
    var row = $(document.getElementById(tableId));

    row.children('.state-row').text(captionerStateToString(state));
    row.children('.mute-row').children().replaceWith(makeUnmuteButton());
    addUnmuteCaptionerListners();

  } else if (state === 3) { // unmuted
    var row = $(document.getElementById(tableId));

    row.children('.state-row').text(captionerStateToString(state));
    row.children('.mute-row').children().replaceWith(makeMuteButton());
    addMuteCaptionerListners();
  } else {
    // NOOP
  }
};

function startWebSocket() {
  socketRocket.start(socketURL).then(function(webSocket) {
    webSocket.onNewMessage = function(message) {
      var encoderState = message['encoderState'];
      var captionerState = message['captionerState'];

      if (typeof encoderState !== 'undefined') {
        changeEncoderState(encoderState.encoderId, encoderState.state);
      } else if (typeof captionerState !== 'undefined') {
        changeCaptionerState(captionerState.captionerId, captionerState.state);
      }      
    };

    webSocket.onerror = function(event) {
      console.log('ERROR: ' + event.data);
    };
  }).catch(function(event) {
    console.log(event);
  });
};

function autoStopWebSocket() {
  $(window).on('beforeunload', function() {
    socketRocket.stop();
  });
};

function recountCaptioners() {
  $('#captioner-selection-table > tbody').children('tr').each(function(index, el) {
    $(el).children('.row-number').text((index + 1) + '');
  });
}

function recountEncoders() {
  var body = $('#encoder-selection-table > tbody');
  var rows = body.children('tr');

  $('#encoder-count').text(rows.length);
  rows.each(function(index, el) {
    $(el).children('.row-number').text((index + 1) + '');
  });
};

function addEncoder(encoder) {
  if (encoder == null || encoder == undefined) {
    return false;
  }

  var wrapper = $('#encoder-list-wrapper');
  if (wrapper.is(':hidden')) {
    $('#encoder-list-header').show();
    wrapper.show();
  }

  var body = $('#encoder-selection-table > tbody');
  var deleteItem = '<td class="col-md-1">' +
      '<p data-placement="top" data-toggle="tooltip" title="Delete">' +
        '<button class="btn btn-danger btn-xs pull-right delete-encoder-button">' +
          '<span class="glyphicon glyphicon-trash"></span>' +
        '</button>' +
      '</p>' +
    '</td>';

  var count = body.children().length;

  var row = $('<tr class="encoder-row" data-encoder-id="' + encoder.ID + '"></tr>');

  row.append('<th scope=row>' + (count + 1) + '</th>');
  row.append('<td>' + encoder.Name + '</td>');
  row.append('<td>' + encoder.IPAddress + '</td>');
  row.append('<td>' + encoder.Port + '</td>');
  row.append('<td>' + encoder.Handle + '</td>');
  row.append('<td>' + encoder.Password + '</td>');
  row.append('<td>' + encoderStateToString(encoder.Status) + '</td>');
  row.append(deleteItem);

  body.append(row);

  return true;
};

function validateNewEncoderForm() {
  var errors = [];
  var ip = $('#encoder-form-ip').val();
  var name = $('#encoder-form-name').val();
  var port = $('#encoder-form-port').val();
  var handle = $('#encoder-form-handle').val();
  var password = $('#encoder-form-password').val();

  // validate IP address field
  if (ip == null || ip.length === 0) {
    errors.push('IP address missing');
  } else {
    if (ip.length < 3) {
      errors.push('IP address too short to be valid');
    } else if (ip.length > 45) {
      errors.push('IP address too long to be valid');
    } else {
      var ipv4Pattern = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
      var ipv6Pattern = /(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))/;

      if (!ipv4Pattern.test(ip) && !ipv6Pattern.test(ip)) {
        errors.push('IP Address is Invalid');
      }      
    }
  }

  // validate name field
  if ((name != null || name.length !== 0) && name.length > 255) {
    errors.push('Name is too Long');
  }

  // validate port field
  if (port == null || port.length === 0) {
    errors.push('Missing port');
  } else {
    var intPort = parseInt(port, 10);
    if (isNaN(intPort)) {
      errors.push('Port is not a Number');
    } else {
      if (intPort < 1 || intPort > 65535) {
        errors.push('Invalid Port. Must be in range [1, 65535].');
      }
    }
  }

  // validate handle field
  if (handle == null || handle.length === 0) {
    errors.push('Missing handle');
  } else {
    if (handle.length > 255) {
      errors.push('Handle too long. Must contain less than 255 characters');
    }
  }

  // validate password field
  if (password == null || password.length === 0) {
    errors.push('Missing password');
  } else {
    if (password.length > 255) {
      errors.push('Password too long. Must contain less than 255 characters');
    }
  }

  return errors;
};

function displayNewEncoderErrors(errors) {
  var container = $('#encoder-form-error-container');
  container.text(errors.join(',\t\t'));
  if (container.is(':hidden')) {
    container.show('fast');
  }
};

function hideNewEncoderErrors() {
  var container = $('#encoder-form-error-container');
  if (!container.is(':hidden')) {
    container.hide('fast');
  }
};

function addAddEncoderHandler() {
  $('#submit-encoder').click(function(event) {
    event.preventDefault();

    var encoderErrors = validateNewEncoderForm();
    if (encoderErrors.length > 0) {
      displayNewEncoderErrors(encoderErrors);
      return;
    }

    hideNewEncoderErrors();

    var form = $('#add-encoder-form');
    var data = form.serializeArray();
    data.push({
      name: 'network_id',
      value: form.attr('data-network-id')
    });

    $.ajax({
      url: '/encoder/add',
      type: 'POST',
      dataType: 'json',
      data: $.param(data),
    }).done(function(encoder) {
      if (addEncoder(encoder)) {
        addDeleteEncoderHandler();
        configureEditing();
        recountEncoders();
        form.find('input.form-control').val('');
      } else {
        alertError('Failed to show new encoder');
      }
    }).fail(alertAjaxFailure);
  });
};

function addEditEncoderHandler() {
  $('#edit-encoder').click(function (e) {
    $.ajax({
      url: '/encoder/' + id,
      type: 'POST',
      dataType: 'json',
      // data: {param1: 'value1'},
    }).done(function() {
      console.log('success');
    }).fail(alertAjaxFailure);
  });
};

function addDeleteEncoderHandler() {
  $('.delete-encoder-button').click(function(event) {
    var row = $(this).closest('tr');
    var encoderId = row.attr('data-encoder-id');

    if (!confirm('Are you sure you want to delete this encoder?')) {
      return;
    }

    $.ajax({
      url: '/encoder/delete/' + encoderId,
      type: 'POST',
      data: $('#delete-encoder-form').serialize()
    }).done(function() {
      row.remove();
      recountEncoders();
    }).fail(alertAjaxFailure);    
  });
};

function getCaptionerData(button) {
  var row = $(button).closest('tr');
  var data = $('#toggle-captioner-mute-form').serializeArray();
  data.push({name: 'ipAddress', value: row.attr('data-captioner-ip')});
  data.push({name: 'numConn', value: row.attr('data-captioner-num-conn')});
  data.push({name: 'networkId', value: row.attr('data-captioner-network-id')});
  return data;
};

function addMuteCaptionerListners() {
  $('.mute-captioner-button').click(function(event) {
    event.preventDefault();

    $.ajax({
      url: '/captioner/mute',
      type: 'POST',
      data: $.param(getCaptionerData(this))
    }).fail(alertAjaxFailure);
  });
};

function addUnmuteCaptionerListners() {
  $('.unmute-captioner-button').click(function(event) {
    event.preventDefault();

    $.ajax({
      url: '/captioner/unmute',
      type: 'POST',
      data: $.param(getCaptionerData(this))
    }).fail(alertAjaxFailure);
  });
};

function addDisconnectCaptionerListeners() {
  $('.disconnect-captioner-button').click(function(event) {
    event.preventDefault();

    $.ajax({
      url: '/captioner/disconnect',
      type: 'POST',
      data: $.param(getCaptionerData(this))
    }).fail(alertAjaxFailure);
  });  
};

function configureEditing() {
  $.fn.editable.defaults.mode = 'inline';

  $('.page-header > h1,h2,h3 > span').editable({
    mode: 'popup',
    placement: 'right',
    url: function(event) {
      var d = new $.Deferred();
      var id = $('#editing-page-header').attr('data-network-id');

      var data = $('#edit-network-form').serializeArray();
      $('.page-header > h1,h2,h3 > span').each(function(index, el) {
        var obj = $(el);

        if (event.name == obj.attr('data-name')) {
          data.push({
            name: obj.attr('name'),
            value: event.value
          });
        } else {
          data.push({
            name: obj.attr('name'),
            value: obj.text()
          });          
        }
      });

      console.log(data);
      if (event.value === 'abc') {
        return d.reject('error message');
      } else {
        $.ajax({
          url: '/network/' + id,
          type: 'POST',
          data: $.param(data),
        }).done(function() {
          d.resolve(this);
        }).fail(alertAjaxFailure);

        return d.promise();
      }
    }
  });

  $('#encoder-selection-table > tbody td').editable({
    mode: 'inline'
  });
};
