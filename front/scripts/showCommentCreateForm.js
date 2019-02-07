function showCommentCreateBox() {
    document.getElementById('add-comment-button').innerHTML =
        '<div class="comment-create-box">\n' +
        '<div class="comment-content-container">\n' +
        '<textarea class="autoExpand comment-content-textarea"\n' +
        'placeholder="Enter your comment..."\n' +
        'data-min-rows="4" rows="4" cols="50"></textarea>\n' +
        '</div>\n' +
        '<div class="comment-author-container">\n' +
        '<span><b><i>Comment as:</i></b></span>\n' +
        '<input class="comment-author-input" placeholder="Enter your username...">\n' +
        '</div>\n' +
        '<div class="publish-comment-button" onclick="publishNewComment()">Publish</div>\n' +
        '</div>'
}

function showCommentReplyBox(commentReplyLink) {
    var commentID = commentReplyLink.getAttribute('data-comment-id');
    var replyButton = commentReplyLink.parentNode;
    var commentActionsList = replyButton.parentNode;
    var commentActionsDiv = commentActionsList.parentNode;

    var isReplyBoxOpened = commentReplyLink.getAttribute('data-opened');
    if (isReplyBoxOpened === "false") {
        var replyBox = document.createElement('div');
        replyBox.className = 'replyBox';
        replyBox.innerHTML =
            '<div class="comment-create-box">\n' +
            '<div class="comment-content-container">\n' +
            '<textarea class="autoExpand comment-content-reply-textarea"\n' +
            'placeholder="Enter your comment..."\n' +
            'data-min-rows="4" rows="4" cols="50"></textarea>\n' +
            '</div>\n' +
            '<div class="comment-author-container">\n' +
            '<span><b><i>Comment as:</i></b></span>\n' +
            '<input class="comment-author-reply-input" placeholder="Enter your username...">\n' +
            '</div>\n' +
            '<div class="publish-comment-button" data-comment-id="' + commentID + '" ' +
            'onclick="replyToComment(this)">Publish</div>\n' +
            '</div>';

        commentActionsDiv.appendChild(replyBox);
        commentReplyLink.setAttribute('data-opened', 'true');
    } else {
        commentActionsDiv.removeChild(commentActionsDiv.getElementsByClassName('replyBox')[0]);
        commentReplyLink.setAttribute('data-opened', 'false');
    }
}