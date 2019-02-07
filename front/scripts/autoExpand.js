$(document).on('focus.autoExpand', 'textarea.autoExpand', function () {
    if (!this.baseHeightIsSet) {
        this.baseHeightIsSet = true;
        this.baseScrollHeight = this.scrollHeight;
    }
}).on('input.autoExpand', 'textarea.autoExpand', function () {
    var minRows = this.getAttribute('data-min-rows') | 0;
    var lineHeight = parseInt($(this).css("line-height"));
    this.rows = minRows;
    var rows = Math.ceil((this.scrollHeight - this.baseScrollHeight) / lineHeight);
    this.rows = minRows + rows;
});