$(document).on('focus.autoExpand', 'textarea.autoExpand', function () {
    if (!this.baseHeightIsSet) {
        this.baseHeightIsSet = true;
        this.baseScrollHeight = this.scrollHeight;
        console.log("Setting base height event fired");
        console.log("baseH: ");
        console.log(this.baseScrollHeight);
    }
}).on('input.autoExpand', 'textarea.autoExpand', function () {
    console.log("Setting new height event fired");
    var minRows = this.getAttribute('data-min-rows') | 0;
    var fontSize = parseInt($(this).css("font-size"));
    this.rows = minRows;
    var rows = Math.ceil((this.scrollHeight - this.baseScrollHeight) / fontSize);
    this.rows = minRows + rows;
});