addEncoder = function(encoder) {
  if (encoder == null || encoder == undefined) {
    return false;
  }

  var row = $('<tr></tr>');
  row.append('<th scope=row></th>');
  row.append('<td>' + encoder.IPAddress + '</td>');
  row.append('<td>' + encoder.Port + '</td>');
  row.append('<td>' + encoder.Name + '</td>');
  row.append('<td>' + encoder.Status + '</td>');

  $('#encoder-selection-table > tbody').append(row);
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
    var ip      = $('#encoder-form-ip').val().trim();
    var port    = $('#encoder-form-port').val().trim();
    var name    = $('#encoder-form-name').val().trim();
    var network = $('#add-encoder-network-element').val().trim();

    var data = {
      'ip': ip,
      'port': port,
      'name': name,
      'network': network
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
