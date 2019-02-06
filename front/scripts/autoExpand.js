$(document).on('focus.autoExpand', 'textarea.autoExpand', function () {
    if (!this.baseHeightIsSet) {
        this.baseHeightIsSet = true;
        this.baseScrollHeight = this.scrollHeight;
    }
}).on('input.autoExpand', 'textarea.autoExpand', function () {
    var minRows = this.getAttribute('data-min-rows') | 0;
    var fontSize = parseInt($(this).css("font-size"));
    this.rows = minRows;
    var rows = Math.ceil((this.scrollHeight - this.baseScrollHeight) / fontSize);
    this.rows = minRows + rows;
});