function publishNewComment() {
    var postID = window.location.pathname.substr(window.location.pathname.lastIndexOf('/') + 1);

    var commentCreateRequest = {PostID: '', ParentID: null, Author: '', Content: ''};
    var userInput = getCommentCreationInput();
    commentCreateRequest.PostID = postID;
    commentCreateRequest.Author = userInput.Author;
    commentCreateRequest.Content = userInput.Content;

    var encodedCommentData = JSON.stringify(commentCreateRequest);
    console.log(encodedCommentData);

    $.ajax(
        {
            url: '/api/comments',
            type: 'POST',
            contentType: 'application/json',
            data: encodedCommentData,
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                var createdComment = response.body;

                var commentsList = document.getElementsByClassName("comments-list")[0]
                    .getElementsByTagName('ul')[0];
                var createdCommentChild = document.getElementById('comment-template').cloneNode(true);
                createdCommentChild.setAttribute('id', createdComment.id);
                createdCommentChild.removeAttribute('style');
                createdCommentChild.getElementsByClassName('username')[0].value = createdComment.author;
                createdCommentChild.getElementsByClassName('creation-time')[0].value = createdComment.date;
                createdCommentChild.getElementsByClassName('comment-content')[0].value = createdComment.content;
                createdCommentChild.getElementsByClassName('comment-reply')[0].setAttribute('data-comment-id',
                    createdComment.id);

                commentsList.appendChild(createdCommentChild);
            },
            error: function (jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}