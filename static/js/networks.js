new Vue({
    el: '#workorder',
    data: {

        newWorkorder: {
            name: '',
            area: '',
            areaNumber: '',
            location: '',
            detail: ''
        },
        workorders: []
    },
    ready: function(){
        this.fetchWorkorders();
    },
    methods: {
         addworkOrder: function(e) {
         e.preventDefault();
         this.newWorkorder.push(this.newWorkorder);
 },
    }
});
