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
  $('#network-selection-table > tbody > tr').click(function(e) {
    var id = $(e.currentTarget).attr('data-network-id');

    if (id != null) {
      e.preventDefault();
      window.location.href = 'networks/' + id;
      return true;
    }

    return false;
  });
};

function addNetworkCreationListener() {
  $('#submit-network').click(function (e) {
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
      success: function(network) {
        if (addNetwork(network)) {
          addNetworkListListeners();
        } else {
          // Maybe prompt to refresh? IDK
        }

        $('#network-form-port').val('');
        $('#network-form-name').val('');
      },
      error: function (xhr, ajaxOptions, thrownError) {
        console.log('++++++++++++++++++++++++++++++++');
        console.log(xhr);
        console.log(ajaxOptions);
        console.log(thrownError);
        console.log('--------------------------------');
      }
    });
  });
};

function startWebSocket() {
  socketRocket.start(socketURL).then(function(webSocket) {
    webSocket.onNewMessage = function(message) {
      console.log('Got new message');
      console.log(message);
      console.log(JSON.stringify(message));
    };

    webSocket.onerror = function(event) {
      console.log("ERROR: " + event.data);
    };

    // setTimeout(function() {
    //   console.log("sending message");
    //   var payload = {
    //     "message": "Hello, Servar."
    //   };
    //
    //   socketRocket.send(payload, function(reply) {
    //     console.log(reply);
    //   });
    // }, 2000);

  }).catch(function(event) {
    console.log(event);
  });
};

$(function () {
  addNetworkListListeners();
  addNetworkCreationListener();
  startWebSocket();

  $('#edit-network').click(function (e) {

    $.ajax({
      url: '/network/add',
      type: 'POST',
      dataType: 'json',
      success: function(network) {

      },
      error: function () {

      }
    });
  });

  $('#network-selection-table > tbody > tr').click(function(event) {
    var that = $(this);
    var networkId = that.attr('data-network-id');
    console.log('The network ID of the clicked row is: ' + encoderId);

    $.ajax({
      url: '/network/' + networkId,
      type: 'DELETE',
      dataType: 'json',
      success: function(network) {
        that.remove();
      },
      error: function() {
        alert("Failed to remove network from list.")
      }
    });
  });
});
