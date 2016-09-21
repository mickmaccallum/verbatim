$('#network-selection-table > tbody > tr').click(function(e) {
  console.log(e.currentTarget.sectionRowIndex);
  window.location.href = 'network.html';
});
