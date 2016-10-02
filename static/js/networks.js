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

  $.ajaxSetup({
    xhrFields: {
      withCredentials: true
    }
  });

  $('#submit-encoder').click(function (e) {
    var ip   = $('#encoder-form-ip').val().trim();
    var port = $('#encoder-form-port').val().trim();
    var name = $('#encoder-form-name').val().trim();

    var data = {
      'ip': ip,
      'port': port,
      'name': name
    }

    $.ajax({
      url: '/encoder/add',
      type: 'POST',
      dataType: 'json',
      data: data,
      crossDomain: true,
      xhrFields: {
        withCredentials: true
      },
      // beforeSend: function(xhr) {
      //   console.log(xhr);
      // },
      success: function(data) {
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
