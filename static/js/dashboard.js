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

  row.attr('data-network-id', network.ID + "");
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
      window.location.href = 'networks/' + id;
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
    }).fail(function() {
      alert("Failed to remove network from list.");
    });
  });
}

function addNetworkCreationListener() {
  $('#submit-network').click(function(event) {
    $.ajax({
      url: '/network/add',
      type: 'POST',
      dataType: 'json',
      data: $(this).closest('form').serialize(),
    }).done(function(network) {
      if (addNetwork(network)) {
        addNetworkListListeners();
        deleteNetworkListListeners();
      } else {
        // Maybe prompt to refresh? IDK
      }

      $('#network-form-port').val('');
      $('#network-form-name').val('');
    }).fail(function() {
      console.log("error");
      console.log(this);
    });
  });
};

function networkStateToString(state) {
  if (!Number.isInteger(state)) {
    return "Disconnected";
  }

  if (state == 0) {
    return "Connected"
  } else if (state == 1) {
    return "Listening"
  } else if (state == 2) {
    return "Listening Failed"
  } else if (state == 3) {
    return "Closed"
  } else if (state == 4) {
    return "Deleted"
  } else {
    return "Disconnected"
  }
};

function changeNetworkState(network, state) {
  if (state == 0) { // connecting
    
  } else if (state == 1 || state == 2 || state == 3) { // listening, listening failed, close
    var row = $('tr[data-network-id=' + network.ID + ']');
    row.children('.state-row').text(networkStateToString(state));
  } else if (state == 4) { // deleted
    
  } else { // disconnected

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
      console.log("ERROR: " + event.data);
    };

    // setTimeout(function() {
    //   console.log("sending message");
    //   var payload = {
    //     "message": "Hello, Servar."
    //   };
    
    //   socketRocket.send(payload, function(reply) {
    //     console.log(reply);

    //     // socketRocket.stop(function() {
    //     //   console.log("finished closing.");
    //     // });
    //   });
    // }, 2000);

  }).catch(function(event) {
    console.log(event);
  });
};

function autoStopWebSocket() {
  $(window).on("beforeunload", function() {
    socketRocket.stop(function() {
      console.log("finished closing.");
    });
  });
};

$(function () {
  addNetworkListListeners();
  deleteNetworkListListeners();
  addNetworkCreationListener();
  startWebSocket();
  autoStopWebSocket();
});
