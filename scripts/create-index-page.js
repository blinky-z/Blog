$(document).ready(function () {
    function generatePostsList(posts) {
        var postsList = document.getElementById("posts-list");
        var blogPostTemplate = Handlebars.compile(postsList.innerHTML);
        postsList.innerHTML = '';

        for (var currentPostNum = 0; currentPostNum < posts.length; currentPostNum++) {
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
            data.readMoreLink = `posts/${post.id}`;

            var blogPostHTML = blogPostTemplate(data);

            postsList.innerHTML += blogPostHTML;
        }
    }

    function generatePageSelector(currentPage) {
        var pageSelector = document.getElementById("blog-page-selector");
        var pageSelectorTemplate = Handlebars.compile(pageSelector.innerHTML);

        var data = {newerPostsLink: '', hasNewerPosts: '', olderPostsLink: ''};

        data.olderPostsLink = `/?page=${currentPage + 1}`;
        if (postsPage === '0') {
            data.hasNewerPosts = 'has-no-posts';
            data.newerPostsLink = '';
        } else {
            data.hasNewerPosts = '';
            data.newerPostsLink = `/?page=${currentPage - 1}`;
        }

        pageSelector.innerHTML = pageSelectorTemplate(data);
    }

    function generateIndexContent(posts) {
        generatePostsList(posts);

        generatePageSelector(parseInt(postsPage));
    }

    var postsPage = new window.URLSearchParams(window.location.search).get('page');
    if (postsPage == null) {
        postsPage = '0';
    }
    $.ajax(
        {
            url: `/posts?page=${postsPage}`,
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