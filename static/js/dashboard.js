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
}

$(function () {
  addNetworkListListeners();

  wsStart().then(function(webSocket) {
    webSocket.onNewMessage = function(message) {
      console.log('Got new message');
      console.log(message);
      console.log(JSON.stringify(message));
    };

    webSocket.onerror = function(event) {
      console.log("ERROR: " + event.data);
    };
  }).catch(function(event) {
    console.log(event);
  });

  $('#submit-network').click(function (e) {
    var port = $('#network-form-port').val().trim();
    var name = $('#network-form-name').val().trim();

    var data = {
      'port': port,
      'name': name
    }

    $.ajax({
      url: '/network/add',
      type: 'POST',
      dataType: 'json',
      data: data,
      success: function(network) {
        if (addNetwork(network)) {

        } else {

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

});

// var payload = {
//   "message": "Hello, Servar."
// };
//
// wsSend(payload, function(reply) {
//   console.log(reply);
// });
