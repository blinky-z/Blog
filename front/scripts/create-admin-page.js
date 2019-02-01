$(document).ready(function () {
    function generatePostsList(posts) {
        var postsList = document.getElementById("admin-posts-list");
        var adminPostTemplate = Handlebars.compile(document.getElementById("admin-post-template").innerHTML);
        postsList.innerHTML = '';

        if (posts == null) {
            return;
        }

        for (var currentPostNum = 0; currentPostNum < posts.length && currentPostNum < 10; currentPostNum++) {
            var post = posts[currentPostNum];
            var data = {postHeader: '', postCreationTime: '', postID: '', postLink: ''};
            data.postHeader = post.title;
            data.postCreationTime = new Date(post.date).toLocaleString();
            data.postID = post.id;
            data.postLink = `/posts/${post.id}`;

            var blogPostHTML = adminPostTemplate(data);

            postsList.innerHTML += blogPostHTML;
        }
    }

    function generatePageSelector(currentPage, posts) {
        var pageSelector = document.getElementById("page-navigation-bar");
        var pageSelectorTemplate = Handlebars.compile(document.getElementById("page-navigation-bar-template").innerHTML);

        var data = {newerPostsLink: '', olderPostsLink: '', currentPageLink: '', currentPageNumber: ''};

        if (posts == null) {
            pageSelector.className = 'has-no-posts';
            pageSelector.innerHTML = pageSelectorTemplate(data);
            return;
        }

        if (currentPage !== 0) {
            data.newerPostsLink = `/admin?page=${currentPage - 1}`;
        }
        if (posts.length > 10) {
            data.olderPostsLink = `/admin?page=${currentPage + 1}`;
        }
        data.currentPageLink = document.documentURI;
        data.currentPageNumber = currentPage;

        pageSelector.innerHTML = pageSelectorTemplate(data);

        if (currentPage === 0) {
            document.getElementById("page-navigation-bar-newer-posts").className = 'has-no-posts';
        }
        if (posts.length <= 10) {
            document.getElementById("page-navigation-bar-older-posts").className = 'has-no-posts';
        }
    }

    function generateAdminPage(posts) {
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

                generateAdminPage(posts);
            },
            error: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                console.log(response.error)
            }
        }
    );
});