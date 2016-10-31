
$(function() {
  addAddEncoderHandler();
  addEditEncoderHandler();
  addDeleteEncoderHandler();
  configureEditing();
  startWebSocket();
  autoStopWebSocket();
});

function changeEncoderState(encoderState) {
  var row = $('.encoder-row[data-encoder-id=\'' + encoderState.encoderId + '\']');
  
}

function startWebSocket() {
  socketRocket.start(socketURL).then(function(webSocket) {
    webSocket.onNewMessage = function(message) {
      var encoderState = message['encoderState'];
      if (typeof encoderState !== 'undefined') {
        changeEncoderState(encoderState);
      }
      
      console.log('Got new message');
      console.log(message);
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
    socketRocket.stop(function() {
      console.log("finished closing.");
    });
  });
};

function recountEncoders() {
  var body = $('#encoder-selection-table > tbody');
  var count = body.children().length;
  $('#encoder-count').text(count);
};

function addEncoder(encoder) {
  if (encoder == null || encoder == undefined) {
    return false;
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
  row.append('<td>' + encoder.IPAddress + '</td>');
  row.append('<td>' + encoder.Name + '</td>');
  row.append('<td>' + encoder.Port + '</td>');
  row.append('<td>' + encoder.Handle + '</td>');
  row.append('<td>' + encoder.Password + '</td>');
  row.append(deleteItem);

  body.append(row);
  return true;
};

function addAddEncoderHandler() {
  $('#submit-encoder').click(function (event) {
    $.ajax({
      url: '/encoder/add',
      type: 'POST',
      dataType: 'json',
      data: $(this).closest('form').serialize(),
      success: function(encoder) {
        if (addEncoder(encoder)) {
          recountEncoders();
        } else {
          
        }

        $('#encoder-form-ip').val('');
        $('#encoder-form-port').val('');
        $('#encoder-form-name').val('');
        $('#encoder-form-handle').val('');
        $('#encoder-form-password').val('');
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

function addEditEncoderHandler() {
  $('#edit-encoder').click(function (e) {
    // var ip      =
    // var port    =
    // var name    =
    // var network =

    $.ajax({
      url: '/encoder/add',
      type: 'POST',
      dataType: 'json',
      success: function(encoder) {

      },
      error: function (xhr, ajaxOptions, thrownError) {

      }
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
      url: '/encoder/' + encoderId + '/delete',
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

function configureEditing() {
  $.fn.editable.defaults.mode = 'inline';

  $('.page-header > h1,h2,h3 > span').editable({
    mode: 'popup',
    placement: 'right'
  });

  $('#encoder-selection-table > tbody td').editable({
    mode: 'inline'
  });
};
