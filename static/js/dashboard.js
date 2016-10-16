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
      console.log("RESPONSE: " + JSON.stringify(message));
    webSocket.onNewMessage = function(message) {
    };

    webSocket.onerror = function(event) {
      console.log("ERROR: " + event.data);
    };
  }).catch(function(event) {
    console.log(event);
  });
});

// var payload = {
//   "message": "Hello, Servar."
// };
//
// wsSend(payload, function(reply) {
//   console.log(reply);
// });
