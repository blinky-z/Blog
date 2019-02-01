$(document).ready(function () {
    function generateEditPage(post) {
        var inputFields = document.getElementsByTagName('textarea');

        var titleInput = inputFields[0];
        titleInput.value = post.title;

        var contentInput = inputFields[1];
        contentInput.value = post.content;
    }

    var postID = new window.URLSearchParams(window.location.search).get('id');

    $.ajax(
        {
            url: `/api/posts/${postID}`,
            type: 'GET',
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                generateEditPage(response.body);
            },
            error: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                console.log(response.error)
            }
        }
    );
});