$('#network-selection-table > tbody > tr').click(function(e) {
  var id = $(e.currentTarget).attr('data-network-id');

  if (id != null) {
    e.preventDefault();
    window.location.href = 'network.html?network=' + id;
    return true;
  }

  return false;
});
