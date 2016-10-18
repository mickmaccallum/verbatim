
$(function() {
  addAddEncoderHandler();
  addEditEncoderHandler();
  addDeleteEncoderHandler();
});

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

  var row = $('<tr></tr>');
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
  $('#submit-encoder').click(function (e) {
    var ip       = $('#encoder-form-ip').val().trim();
    var name     = $('#encoder-form-name').val().trim();
    var port     = $('#encoder-form-port').val().trim();
    var handle   = $('#encoder-form-handle').val().trim();
    var password = $('#encoder-form-password').val().trim();
    var network  = $('#add-encoder-network-element').val().trim();

    var data = {
      'ip_address': ip,
      'name': name,
      'port': port,
      'handle': handle,
      'password': password,
      'network_id': network
    }

    $.ajax({
      url: '/encoder/add',
      type: 'POST',
      dataType: 'json',
      data: data,
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
      url: '/encoder/' + encoderId,
      type: 'DELETE',
      success: function(msg) {
        row.remove();
        recountEncoders();
      },
      error: function (xhr, ajaxOptions, thrownError) {
        alert(thrownError);
      }
    });
  });
};

