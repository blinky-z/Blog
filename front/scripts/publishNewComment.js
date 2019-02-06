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
                createdCommentChild.getElementsByClassName('username')[0].getElementsByTagName('a')[0]
                    .getElementsByTagName('b')[0].innerHTML = createdComment.author;
                var options = {
                    weekday: 'short', year: 'numeric', month: 'narrow', day: 'numeric',
                    hour: 'numeric', minute: '2-digit', second: '2-digit'
                };
                createdCommentChild.getElementsByClassName('creation-time')[0].getElementsByTagName('a')[0]
                    .innerHTML = convertToGoTimeFormat(createdComment.date);
                createdCommentChild.getElementsByClassName('comment-content')[0].innerHTML = createdComment.content;
                createdCommentChild.getElementsByClassName('comment-reply')[0].setAttribute('data-comment-id',
                    createdComment.id);

                commentsList.appendChild(createdCommentChild);

                alert('Comment successfully created');

                document.getElementById('comment-content-textarea').value = '';
                document.getElementById('comment-content-textarea')
                    .setAttribute('rows', document.getElementById('comment-content-textarea')
                        .getAttribute('data-min-rows'));
            },
            error: function (jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}