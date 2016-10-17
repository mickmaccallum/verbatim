addEncoder = function(encoder) {
  if (encoder == null || encoder == undefined) {
    return false;
  }

  var body = $('#encoder-selection-table > tbody');
  var count = body.children().length;

  var row = $('<tr></tr>');
  row.append('<th scope=row>' + (count + 1) + '</th>');
  row.append('<td>' + encoder.IPAddress + '</td>');
  row.append('<td>' + encoder.Port + '</td>');
  row.append('<td>' + encoder.Name + '</td>');
  row.append('<td>' + encoder.Status + '</td>');

  body.append(row);
  return true;
};

$(function() {
  // new Vue({
  //   el: '#add-encoder-form',
  //   data: {
  //     newWorkorder: {
  //       name: '',
  //       area: '',
  //       areaNumber: '',
  //       location: '',
  //       detail: ''
  //     },
  //     workorders: []
  //   },
  //   ready: function() {
  //       // this.fetchWorkorders();
  //   },
  //   methods: {
  //     addworkOrder: function(e) {
  //       e.preventDefault();
  //       this.newWorkorder.push(this.newWorkorder);
  //     },
  //   }
  // });

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
      error: function () {

      }
    });
  });

  $('#encoder-selection-table > tbody > tr').click(function(event) {
    var that = $(this);
    var encoderId = that.attr('data-encoder-id');
    console.log('The encoder ID of the clicked row is: ' + encoderId);

    $.ajax({
      url: '/encoder/' + encoderId,
      type: 'DELETE',
      dataType: 'json',
      success: function(encoder) {
        that.remove();
      },
      error: function() {
        alert("Failed to remove encoder from list.")
      }
    });
  });
});
