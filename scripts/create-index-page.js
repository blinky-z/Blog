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

        //     var postChild = document.createElement('div');
        //     postChild.className = 'blog-post';
        //
        //     var headerChild = document.createElement('h1');
        //     headerChild.className = 'blog-post-header';
        //     $(headerChild).text(post.title);
        //
        //     var descriptionChild = document.createElement('h3');
        //     descriptionChild.className = 'blog-post-description';
        //     $(descriptionChild).text(`Written by <b>Dmitry</b> on <b>${post.date}</b>`);
        //
        //     var postSnippetChild = document.createElement('p');
        //     postSnippetChild.className = 'blog-post-snippet';
        //     $(postSnippetChild).text(post.content.substring(0, 160));
        //
        //     var postShareChild = document.createElement('div');
        //     postShareChild.className = 'blog-post-share';
        //     $(postShareChild).text('Share: ');
        //
        //     var postShareSocials = document.createElement('div');
        //     postShareSocials.className = 'blog-post-share-socials';
        //
        //     var postShareSocialsList = document.createElement('ul');
        //
        //     var shareVk = document.createElement('li');
        //     var shareVkLink = document.createElement('a');
        //     shareVkLink.setAttribute('href', 'https://vk.com');
        //     shareVkLink.setAttribute('target', '_blank');
        //     var shareVkImage = document.createElement('img');
        //     shareVkImage.setAttribute('src', 'images/vk-icon.png');
        //     shareVkImage.setAttribute('width', '18');
        //     shareVkImage.setAttribute('height', '18');
        //     shareVkImage.setAttribute('title', 'VK');
        //
        //     shareVkLink.appendChild(shareVkImage);
        //     shareVk.appendChild(shareVkLink);
        //     postShareSocialsList.appendChild(shareVk);
        //
        //     var shareFacebook = document.createElement('li');
        //     var shareFacebookLink = document.createElement('a');
        //     shareVkLink.setAttribute('href', 'https://vk.com');
        //     shareVkLink.setAttribute('target', '_blank');
        //     var shareFacebookImage = document.createElement('img');
        //     shareVkImage.setAttribute('src', 'images/facebook-icon.png');
        //     shareVkImage.setAttribute('width', '18');
        //     shareVkImage.setAttribute('height', '18');
        //     shareVkImage.setAttribute('title', 'Facebook');
        //
        //     shareVkLink.appendChild(shareFacebookImage);
        //     shareVk.appendChild(shareFacebookLink);
        //     postShareSocialsList.appendChild(shareFacebook);
        //
        //     postShareChild.appendChild(postShareSocialsList);
        //
        //     var postReadMore = document.createElement('div');
        //     postReadMore.className = 'blog-post-read-more';
        //
        //     var postReadMoreLink = document.createElement('a');
        //     postReadMoreLink.setAttribute('href', `/posts/${post.id}`);
        //     $(postReadMoreLink).text('Read More');
        //
        //     postReadMore.appendChild(postReadMoreLink);
        //
        //     postChild.appendChild(headerChild);
        //     postChild.appendChild(descriptionChild);
        //     postChild.appendChild(postSnippetChild);
        //     postChild.appendChild(postShareChild);
        //     postChild.appendChild(postReadMore);
        //
        //     content.appendChild(postChild);
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