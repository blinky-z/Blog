$(document).ready(function () {
    var content = document.getElementById("content");

    function addNewPost(post) {
        console.log(post);
        var data = {postHeader: '', postAuthor: '', postCreationTime: '', postSnippet: '', readMoreLink: ''};
        data.postHeader = post.title;
        data.postAuthor = 'Dmitry';
        data.postCreationTime = post.date;
        if (post.content.length < 160) {
            data.postSnippet = post.content;
        } else {
            data.postSnippet = post.content.substr(0, 160);
        }
        data.readMoreLink = `posts/${post.id}`;

        nunjucks.configure('.', {autoescape: true});
        nunjucks.render('index.html', data);
    }

    var postsPage = new window.URLSearchParams(window.location.search).get('page');
    if (postsPage == null) {
        postsPage = '1';
    }
    $.ajax(
        {
            url: `/posts?page=${postsPage}`,
            type: 'GET',
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                var posts = response.body;

                for (var currentPost = 0; currentPost < 1; currentPost++) {
                    addNewPost(posts[currentPost]);
                }
            },
            error: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                console.log(response.error)
            }
        }
    );
});