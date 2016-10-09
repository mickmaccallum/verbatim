$(function () {
  $('#network-selection-table > tbody > tr').click(function(e) {
    var id = $(e.currentTarget).attr('data-network-id');

    if (id != null) {
      e.preventDefault();
      window.location.href = 'networks/' + id;
      return true;
    }

    return false;
  });
});
