$(document).ready(function () {
    function generatePostsList(posts) {
        var postsList = document.getElementById("posts-list");
        var blogPostTemplate = Handlebars.compile(document.getElementById("blog-post-template").innerHTML);
        postsList.innerHTML = '';

        if (posts == null) {
            return;
        }

        for (var currentPostNum = 0; currentPostNum < posts.length && currentPostNum < 10; currentPostNum++) {
            var post = posts[currentPostNum];
            var data = {postHeader: '', postAuthor: '', postCreationTime: '', postSnippet: '', readMoreLink: ''};
            data.postHeader = post.title;
            data.postAuthor = 'Dmitry';
            data.postCreationTime = post.date;
            if (post.content.length < 160) {
                data.postSnippet = post.content;
            } else {
                data.postSnippet = post.content.substr(0, 160);
            }
            data.readMoreLink = `/posts/${post.id}`;

            var blogPostHTML = blogPostTemplate(data);

            postsList.innerHTML += blogPostHTML;
        }
    }

    function generatePageSelector(currentPage, posts) {
        var pageSelector = document.getElementById("blog-page-selector");
        var pageSelectorTemplate = Handlebars.compile(document.getElementById("blog-page-selector-template").innerHTML);

        var data = {newerPostsLink: '', olderPostsLink: ''};

        if (posts == null) {
            pageSelector.className = 'has-no-posts';
            pageSelector.innerHTML = pageSelectorTemplate(data);
            return;
        } else {
            data.newerPostsLink = `/?page=${currentPage - 1}`;
            data.olderPostsLink = `/?page=${currentPage + 1}`;
        }

        pageSelector.innerHTML = pageSelectorTemplate(data);
        if (currentPage === 0) {
            document.getElementById("blog-page-selector-newer-posts").className = 'has-no-posts';
        }
        if (posts.length <= 10) {
            document.getElementById("blog-page-selector-older-posts").className = 'has-no-posts';
        }
    }

    function generateIndexContent(posts) {
        generatePostsList(posts);

        generatePageSelector(parseInt(postsPage), posts);
    }

    var postsPage = new window.URLSearchParams(window.location.search).get('page');
    if (postsPage == null) {
        postsPage = '0';
    }

    $.ajax(
        {
            url: `/api/posts?page=${postsPage}&posts-per-page=11`,
            type: 'GET',
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                var posts = response.body;

                generateIndexContent(posts);
            },
            error: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                console.log(response.error)
            }
        }
    );
});