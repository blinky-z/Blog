function levelOfComment(commentDiv) {
    var level = 0;

    var currentLevelCommentsList = commentDiv.parentNode;

    for (; ;) {
        currentLevelCommentsList = currentLevelCommentsList.parentNode;
        if (currentLevelCommentsList.tagName === 'UL') {
            level++;
            continue;
        }

        break;
    }

    return (level - 1);
}

function replyToComment(commentReplyButtonLink) {
    var commentID = commentReplyButtonLink.getAttribute('data-comment-id');
    var currentComment = document.getElementById(commentID);

    var postID = window.location.pathname.substr(window.location.pathname.lastIndexOf('/') + 1);

    var commentCreateRequest = {PostID: '', parentCommentID: '', Author: '', Content: ''};

    commentCreateRequest.PostID = postID;
    commentCreateRequest.parentCommentID = commentID;
    commentCreateRequest.Author = currentComment.getElementsByClassName("comment-author-reply-input")[0].value;
    commentCreateRequest.Content = currentComment.getElementsByClassName("comment-content-reply-textarea")[0].value;

    var encodedCommentData = JSON.stringify(commentCreateRequest);

    $.ajax(
        {
            url: '/api/comments',
            type: 'POST',
            contentType: 'application/json',
            data: encodedCommentData,
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                var createCommentResponse = response.body;

                var currentCommentsList = currentComment.parentNode;

                var newComment = createCommentChild(createCommentResponse);

                if (levelOfComment(currentComment) < 5) {
                    var commentWithChilds = document.createElement('ul');
                    commentWithChilds.appendChild(newComment);
                    currentCommentsList.appendChild(commentWithChilds);
                } else {
                    currentCommentsList.appendChild(newComment);
                }

                alert('Comment successfully created');

                // close reply form
                currentComment.getElementsByClassName('replyBox')[0].parentNode
                    .removeChild(currentComment.getElementsByClassName('replyBox')[0]);
                currentComment.getElementsByClassName('comment-reply')[0]
                    .setAttribute('data-opened', 'false')
            },
            error: function (jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}