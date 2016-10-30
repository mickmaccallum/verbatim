function addNetwork(network) {
  if (network == null || network == undefined) {
    return false;
  }

  var body = $('#network-selection-table > tbody');
  var count = body.children().length;

  var row = $('<tr></tr>');
  row.append('<th scope=row>' + (count + 1) + '</th>');
  row.append('<td>' + network.Name + '</td>');
  row.append('<td>' + network.ListeningPort + '</td>');
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

    var row = $(this).closest('tr');
    var networkId = row.attr('data-network-id');
    var networkName = row.attr('data-network-name');

    if (!confirm('Are you sure you want to delete the network: ' + networkName)) {
      return;
    }

    $.ajax({
      url: '/network/' + networkId,
      type: 'DELETE',
    }).done(function() {
      row.remove();
    }).fail(function() {
      alert("Failed to remove network from list.");
    });
  });
}

function addNetworkCreationListener() {
  $('#submit-network').click(function (event) {
    var port = $('#network-form-port').val().trim();
    var name = $('#network-form-name').val().trim();

    var data = {
      'listening_port': port,
      'name': name
    }

    $.ajax({
      url: '/network/add',
      type: 'POST',
      dataType: 'json',
      data: data,
    }).done(function() {
      if (addNetwork(network)) {
        addNetworkListListeners();
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

function startWebSocket() {
  socketRocket.start(socketURL).then(function(webSocket) {
    webSocket.onNewMessage = function(message) {
      console.log('Got new message');
      console.log(message);
    };

    webSocket.onerror = function(event) {
      console.log("ERROR: " + event.data);
    };

    setTimeout(function() {
      console.log("sending message");
      var payload = {
        "message": "Hello, Servar."
      };
    
      socketRocket.send(payload, function(reply) {
        console.log(reply);

        // socketRocket.stop(function() {
        //   console.log("finished closing.");
        // });
      });
    }, 2000);

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
