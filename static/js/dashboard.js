function addNetwork(network) {
  if (network == null || network == undefined) {
    return false;
  }

  var body = $('#network-selection-table > tbody');
  
  var deleteItem = '<td class="col-xl-1 col-lg-1 col-md-1">' +
                      '<p data-placement="top" data-toggle="tooltip" title="Delete">' +
                        '<button class="btn btn-danger btn-xs pull-right delete-button" ' +
                          'data-title="Delete" data-toggle="modal" ' +
                          'data-target="#delete">' +
                            '<span class="glyphicon glyphicon-trash"></span>' +
                        '</button>' +
                      '</p>' +
                    '</td>';


  var count = body.children().length;

  var row = $('<tr></tr>');

  row.attr('data-network-id', network.ID + '');
  row.attr('data-network-name', network.Name);

  row.append('<th class="col-xl-1 col-lg-1 col-md-1 row-number" scope=row>' + (count + 1) + '</th>');
  row.append('<td class="col-xl-3 col-lg-3 col-md-3">' + network.Name + '</td>');
  row.append('<td class="col-xl-2 col-lg-2 col-md-2">' + network.ListeningPort + '</td>');
  row.append('<td class="col-xl-3 col-lg-3 col-md-3">' + network.Timeout + '</td>');
  row.append('<td class="col-xl-2 col-lg-2 col-md-2 state-row">' + network.State + '</td>');
  row.append(deleteItem);

  body.append(row);
  
  return true;
};

function addNetworkListListeners() {
  $('#network-selection-table > tbody > tr').click(function(event) {
    var id = $(event.currentTarget).attr('data-network-id');

    if (id != null) {
      event.preventDefault();
      window.location.href = '/network/' + id;
      return true;
    }

    return false;
  });
};

function deleteNetworkListListeners() {
  $('.delete-button').click(function(event) {
    event.stopPropagation();
    event.preventDefault();

    var row = $(this).closest('tr');
    var networkId = row.attr('data-network-id');
    var networkName = row.attr('data-network-name');

    if (!confirm('Are you sure you want to delete the network: ' + networkName)) {
      return;
    }

    $.ajax({
      url: '/network/delete/' + networkId,
      type: 'POST',
      data: $('#delete-network-form').serialize()
    }).done(function() {
      row.remove();
    }).fail(alertAjaxFailure);
  });
}

function validateNewNetworkForm() {
  var errors = [];
  var name = $('#network-form-name').val();
  var port = $('#network-form-port').val();
  var timeout = $('#network-form-timeout').val();


  // validate name field
  if (name == null || name.length === 0) {
    errors.push('Missing network name');
  } else {
    if (name.length > 255) {
      errors.push('Network name too long. Must contain less than 255 characters');
    }
  }

  // validate port field
  if (port == null || port.length === 0) {
    errors.push('Missing port');
  } else {
    var intPort = parseInt(port, 10);
    if (isNaN(intPort)) {
      errors.push('Port is not a number');
    } else {
      if (intPort < 1 || intPort > 65535) {
        errors.push('Invalid Port. Must be in range [1, 65535].');
      }
    }
  }

  // validate timeout field
  if (timeout == null || timeout.length === 0) {
    errors.push('Missing timeout');
  } else {
    var intTimeout = parseInt(timeout, 10);
    if (isNaN(intTimeout)) {
      errors.push('Timeout is not a number');
    } else {
      if (intTimeout < 10) {
        errors.push('Timeout too short. Must be at least 10 seconds');
      } else if (intTimeout > 1800) {
        errors.push('Timeout too long. Must be less than 1800 seconds (30 minutes).')
      }
    }
  }

  return errors;
};

function displayNewNetworkErrors(errors) {
  var container = $('#network-form-error-container');
  container.text(errors.join(',\t\t'));
  if (container.is(':hidden')) {
    container.show('fast');
  }
};

function hideNewNetworkErrors() {
  var container = $('#network-form-error-container');
  if (!container.is(':hidden')) {
    container.hide('fast');
  }
};

function addNetworkCreationListener() {
  $('#submit-network').click(function(event) {
    event.preventDefault();

    var networkErrors = validateNewNetworkForm();
    if (networkErrors.length > 0) {
      displayNewNetworkErrors(networkErrors);
      return;
    }

    hideNewNetworkErrors();
    var form = $(this).closest('form');

    $.ajax({
      url: '/network/add',
      type: 'POST',
      dataType: 'json',
      data: form.serialize(),
    }).done(function(network) {
      if (addNetwork(network)) {
        addNetworkListListeners();
        deleteNetworkListListeners();
      }

      form.find('input.form-control').val('');
    }).fail(alertAjaxFailure);
  });
};

function networkStateToString(state) {
  if (!Number.isInteger(state)) {
    return 'Disconnected';
  }

  if (state == 0) {
    return 'Connected'
  } else if (state == 1) {
    return 'Listening'
  } else if (state == 2) {
    return 'Listening Failed'
  } else if (state == 3) {
    return 'Closed'
  } else if (state == 4) {
    return 'Deleted'
  } else {
    return 'Disconnected'
  }
};

function changeNetworkState(network, state) {
  var row = $('tr[data-network-id=' + network.ID + ']');

  if (state === 4) { // deleted
    row.remove();
  } else { // connecting, disconnected, listening, listening failed, close
    row.children('.state-row').text(networkStateToString(state));
  }
};

function startWebSocket() {
  socketRocket.start(socketURL).then(function(webSocket) {
    webSocket.onNewMessage = function(message) {
      var networkState = message['networkState'];

      if (typeof networkState !== 'undefined') {
        changeNetworkState(networkState.network, networkState.state);
      }
    };

    webSocket.onerror = function(event) {
      console.log('Web Socket Error: ' + event.data);
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

$(function () {
  addNetworkListListeners();
  deleteNetworkListListeners();
  addNetworkCreationListener();
  startWebSocket();
  autoStopWebSocket();
});
