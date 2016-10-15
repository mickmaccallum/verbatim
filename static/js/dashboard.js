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

var ws;

function startWebSocket() {
  if (ws != null) {
    return;
  }

  ws = new WebSocket(socketURL);

  ws.onopen = function(event) {
    console.log("OPEN");

    console.log("sending message");

    var payload = {
      "message": "Hello, Servar."
    };

    ws.send(JSON.stringify(payload));
  }

  ws.onclose = function(event) {
    console.log("CLOSE");
    ws = null;
  }

  ws.onmessage = function(event) {
    var msg = JSON.parse(event.data);

    console.log("RESPONSE: " + JSON.stringify(msg));
  }

  ws.onerror = function(event) {
    console.log("ERROR: " + event.data);
  }
}

$(function () {
  addNetworkListListeners();
  startWebSocket();
});
