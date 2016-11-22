
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

function changeEncoderState(encoderState) {
  var row = $('.encoder-row[data-encoder-id=\'' + encoderState.encoderId + '\']');
  // console.log(row);
};

function encoderStateToString(state) {
  if (!Number.isInteger(state)) {
    return "Disconnected";
  }

  if (state == 0) {
    return "Connected";
  } else if (state == 1) {
    return "Connecting";
  } else if (state == 2) {
    return "Authentication Failed";
  } else if (state == 3) {
    return "Writes Failing";
  } else {
    return "Disconnected";
  }
};

function captionerStateToString(state) {
  if (state == 0) {
    return "Connected";
  } else if (state == 1) {
    return "Disconnecting";
  } else if (state == 2) {
    return "Muted";
  } else if (state == 3) {
    return "Unmuted";
  } else {
    return "Disconnected";
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

  var headers = '<th class="row-number" scope=row>0</th>' + 
    '<td>' + captioner.IPAddr + '</td>' + 
    '<td>' + captioner.NumConn + '</td>' + 
    '<td class="state-row">' + captionerStateToString(state) + '</td>';

  var muteColumn = '';
  if (state == 2 || state == 3) {
    if (state == 2) {
      muteColumn = makeUnmuteButton();
    } else {
      muteColumn = makeMuteButton();
    }
  } else {
    // TODO: handle this.
  }

  var endRow = '</tr>';

  var row = $(openRow + headers + muteColumn + endRow);
  $('#captioner-selection-table > tbody').prepend(row);
  recountCaptioners();
};

function changeCaptionerState(captioner, state) {
  var tableId = [captioner.IPAddr, captioner.NumConn, captioner.NetworkID].join(":");

  if (state == 0) { // connected
    addCaptioner(captioner, tableId, state);

  } else if (state == 1) { // disconnected
    $(document.getElementById(tableId)).hide('slow', function(){ 
      this.remove(); 
    });

  } else if (state == 2) { // muted
    var row = $(document.getElementById(tableId));

    row.children('.state-row').text(captionerStateToString(state));
    row.children('.mute-row').children().replaceWith(makeUnmuteButton());
    addUnmuteCaptionerListners();

  } else if (state == 3) { // unmuted
    var row = $(document.getElementById(tableId));

    row.children('.state-row').text(captionerStateToString(state));
    row.children('.mute-row').children().replaceWith(makeMuteButton());
    addMuteCaptionerListners();

  } else {

  }
};

function startWebSocket() {
  socketRocket.start(socketURL).then(function(webSocket) {
    webSocket.onNewMessage = function(message) {
      var encoderState = message['encoderState'];
      var captionerState = message['captionerState'];

      if (typeof encoderState !== 'undefined') {
        changeEncoderState(encoderState);
      } else if (typeof captionerState !== 'undefined') {
        changeCaptionerState(captionerState.captionerId, captionerState.state);
      }      
    };

    webSocket.onerror = function(event) {
      console.log("ERROR: " + event.data);
    };
  }).catch(function(event) {
    console.log(event);
  });
};

function autoStopWebSocket() {
  $(window).on("beforeunload", function() {
    socketRocket.stop();
  });
};

function recountCaptioners() {
  $('#captioner-selection-table > tbody').children('tr').each(function(index, el) {
    $(el).children('.row-number').text((index + 1) + "");
  });
}

function recountEncoders() {
  var body = $('#encoder-selection-table > tbody');
  var rows = body.children('tr');

  $('#encoder-count').text(rows.length);
  rows.each(function(index, el) {
    $(el).children('.row-number').text((index + 1) + "");
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

function addAddEncoderHandler() {
  $('#submit-encoder').click(function(event) {
    event.preventDefault();
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

        // form.children('.form-control').val('');
      } else {
        alertError('Failed to show new encoder');
      }
    }).fail(function(xhr, status, error) {
      var message = '';
      if (xhr.responseText != null) {
        message = xhr.responseText;
      } else {
        message = error;
      }

      alertError(message);
    });
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
      console.log("success");
    }).fail(function() {
      console.log("error");
    });
  });
};

function addDeleteEncoderHandler() {
  $('.delete-encoder-button').click(function(event) {
    var row = $(this).closest('tr');
    var encoderId = row.attr('data-encoder-id');

    if (!confirm("Are you sure you want to delete this encoder?")) {
      return;
    }

    $.ajax({
      url: '/encoder/delete/' + encoderId,
      type: 'POST',
      data: $('#delete-encoder-form').serialize()
    }).done(function() {
      row.remove();
      recountEncoders();
    }).fail(function(error) {
      console.log('error');
      console.log(error);
    });    
  });
};

function getCaptionerData(button) {
  var row = $(button).closest('tr');
  var data = $('#toggle-captioner-mute-form').serializeArray();
  data.push({name: "ipAddress", value: row.attr('data-captioner-ip')});
  data.push({name: "numConn", value: row.attr('data-captioner-num-conn')});
  data.push({name: "networkId", value: row.attr('data-captioner-network-id')});
  return data;
};

function addMuteCaptionerListners() {
  $('.mute-captioner-button').click(function(event) {
    event.preventDefault();

    $.ajax({
      url: '/captioner/mute',
      type: 'POST',
      data: $.param(getCaptionerData(this))
    }).fail(function() {
      console.log('error');
      console.log(this);
    });
  });
};

function addUnmuteCaptionerListners() {
  $('.unmute-captioner-button').click(function(event) {
    event.preventDefault();

    $.ajax({
      url: '/captioner/unmute',
      type: 'POST',
      data: $.param(getCaptionerData(this))
    }).fail(function() {
      console.log('error');
      console.log(this);
    });
  });
};

function addDisconnectCaptionerListeners() {
  $('.disconnect-captioner-button').click(function(event) {
    event.preventDefault();

    $.ajax({
      url: '/captioner/disconnect',
      type: 'POST',
      data: $.param(getCaptionerData(this))
    }).fail(function() {
      console.log('error');
      console.log(this);
    });
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
        }).fail(function() {
          d.reject(this);
        });

        return d.promise();
      }
    }
  });

  $('#encoder-selection-table > tbody td').editable({
    mode: 'inline'
  });
};
