var socketRocket = (function() {
  var webSocket;
  var reference = 0;
  var waitQueue = {};
  var exports = {
    onComplete: null
  };

  function loadCallback(message) {
    if (message == null) {
      return null;
    }

    var ref = message['Reference'];
    if (ref == null) {
      return null;
    }

    return waitQueue[ref];
  };

  function removeCallback(message) {
    var ref = message['Reference'];

    waitQueue[ref] = null;
    delete waitQueue[reference];
  };

  function chainMessageSend(payload, completion) {
    return new Promise(function(resolve, reject) {
      var message = {
        'Reference': ++reference,
        'Payload': payload
      };

      waitQueue[reference] = completion;
      webSocket.send(JSON.stringify(message));

      resolve(message);
    });
  };

  exports.start = function(url) {
    if (webSocket != null) {
      return Promise.resolve(webSocket);
    }

    webSocket = new WebSocket(url);

    return new Promise(function(resolve, reject) {
      webSocket.onopen = function(event) {
        resolve(webSocket);
      };

      webSocket.onmessage = function(event) {
        var message = JSON.parse(event.data);
        var handler = loadCallback(message);
        var payload = message['Payload'];

        if (handler == null) {
          webSocket.onNewMessage(payload)
        } else {
          handler(payload);
          removeCallback(message);
        }
      }

      webSocket.onclose = function(event) {
        webSocket = null;

        if (event.code === 1005 && exports.onComplete != null && typeof(exports.onComplete) === typeof(Function)) {
          exports.onComplete();
          exports.onComplete = null;
          return;
        }

        reject(event);
      };
    });
  };

  exports.stop = function(completion) {
    if (webSocket == null) {
      if (completion != null && typeof(completion) === typeof(Function)) {
        completion();
      }
      
      return;
    }

    exports.onComplete = completion;
    webSocket.close();
  };

  exports.send = function(payload, completion) {
    if (webSocket != null) {
      return chainMessageSend(payload, completion);
    }

    return wsStart().then(function(value) {
      return chainMessageSend(payload, completion);
    }, function(error) {
      return Promise.reject(error);
    });
  };

  return exports;
})();
