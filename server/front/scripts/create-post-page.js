$(document).ready(function () {
    function generatePost(post) {
        console.log("Post generating started");

        var postChild = document.getElementsByClassName('post')[0];
        var postTemplate = Handlebars.compile(document.getElementById('post-template').innerHTML);

        var data = {postHeader: '', postAuthor: '', postCreationTime: '', postContent: ''};
        data.postHeader = post.title;
        data.postAuthor = 'Dmitry';
        data.postCreationTime = post.date;
        data.postContent = post.content;

        postChild.innerHTML = postTemplate(data);

        document.title = post.title;
    }

    console.log("script running..");

    var path = window.location.pathname;
    var id = path.substr(path.lastIndexOf('/'));

    console.log('Id is: ');
    console.log(id);
    $.ajax(
        {
            url: `/api/posts/${id}`,
            type: 'GET',
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                var post = response.body;

                generatePost(post);
            },
            error: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                console.log(response.error)
            }
        }
    );
});