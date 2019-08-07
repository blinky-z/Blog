var viewer;

$(document).ready(function () {
    renderMD(document.getElementById('content').value)

});

function renderMD(text) {
    viewer = new tui.Editor.factory({
        el: document.querySelector('#content'),
        viewer: true,
        height: 'auto',
        usageStatistics: false,
        initialValue: text
    });
}