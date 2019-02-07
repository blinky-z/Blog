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

function replyToComment(commentReplyButton) {
    var commentID = commentReplyButton.getAttribute('data-comment-id');
    var currentLevelCommentsList = document.getElementById(commentID).parentNode;

    var postID = window.location.pathname.substr(window.location.pathname.lastIndexOf('/') + 1);

    var commentCreateRequest = {PostID: '', ParentID: commentID, Author: '', Content: ''};

    commentCreateRequest.PostID = postID;
    commentCreateRequest.ParentID = commentID;
    commentCreateRequest.Author = document.getElementById(commentID)
        .getElementsByClassName("comment-author-reply-input")[0].value;
    commentCreateRequest.Content = document.getElementById(commentID)
        .getElementsByClassName("comment-content-reply-textarea")[0].value;

    var encodedCommentData = JSON.stringify(commentCreateRequest);

    $.ajax(
        {
            url: '/api/comments',
            type: 'POST',
            contentType: 'application/json',
            data: encodedCommentData,
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                var createdComment = response.body;

                var commentsList = currentLevelCommentsList;
                var createdCommentChild = document.getElementById('comment-template').cloneNode(true);
                createdCommentChild.setAttribute('id', createdComment.id);
                createdCommentChild.removeAttribute('style');
                createdCommentChild.getElementsByClassName('username')[0].getElementsByTagName('a')[0]
                    .getElementsByTagName('b')[0].innerHTML = createdComment.author;
                createdCommentChild.getElementsByClassName('creation-time')[0].getElementsByTagName('a')[0]
                    .innerHTML = convertToGoTimeFormat(createdComment.date);
                createdCommentChild.getElementsByClassName('comment-content')[0].innerHTML = createdComment.content;
                createdCommentChild.getElementsByClassName('comment-reply')[0].setAttribute('data-comment-id',
                    createdComment.id);

                if (levelOfComment(document.getElementById(commentID)) < 5) {
                    var createdCommentChildList = document.createElement('ul');
                    createdCommentChildList.appendChild(createdCommentChild);
                    commentsList.appendChild(createdCommentChildList);
                } else {
                    commentsList.appendChild(createdCommentChild);
                }

                alert('Comment successfully created');

                var replyButton = commentReplyButton.parentNode.parentNode;
                replyButton.innerHTML = '<a class="comment-reply" data-comment-id="' + commentID + '"\n' +
                    'href="#" onclick="showCommentReplyBox(this); return false;">Reply</a>';
            },
            error: function (jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}