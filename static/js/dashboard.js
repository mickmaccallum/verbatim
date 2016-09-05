$('#network-selection-table > tbody > tr').click(function(e) {
  console.log(e.currentTarget.sectionRowIndex);
  console.log(e);
  window.location.href = 'network.html';
});
