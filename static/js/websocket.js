var webSocket;

function wsStart() {
  if (webSocket != null) {
    return Promise.resolve(webSocket);
  }

  webSocket = new WebSocket(socketURL);

  return new Promise(function(resolve, reject) {
    webSocket.onopen = function(event) {
      console.log("OPEN");
      resolve(webSocket);
    };

    webSocket.onmessage = function(event) {
      var message = JSON.parse(event.data);
      var handler = loadCallback(message);
      var payload = message['payload'];

      if (handler == null) {
        webSocket.onNewMessage(payload)
      } else {
        handler(payload);
        removeCallback(message);
      }
    }

    webSocket.onclose = function(event) {
      console.log("CLOSE");
      webSocket = null;
      reject(event);
    };
  });
};

var reference = 0;
var waitQueue = {};

function wsSend(payload, completion) {
  if (webSocket != null) {
    return chainMessageSend(payload, completion);
  }

  return wsStart().then(function(value) {
    return chainMessageSend(payload, completion)
  }, function(error) {
    return Promise.reject(error);
  });
};

function loadCallback(message) {
  if (message == null) {
    return null;
  }

  var ref = message['reference'];
  if (ref == null) {
    return null;
  }

  return waitQueue[ref];
};

function removeCallback(message) {
  var ref = message['reference'];

  waitQueue[ref] = null;
  delete waitQueue[reference];
};

function chainMessageSend(payload, completion) {
  return new Promise(function(resolve, reject) {
    var message = {
      'reference': ++reference,
      'payload': payload
    };

    waitQueue[reference] = completion
    webSocket.send(JSON.stringify(message));

    resolve(message);
  });
};
