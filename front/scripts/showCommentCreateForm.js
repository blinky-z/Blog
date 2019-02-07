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
    replyButton.innerHTML =
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
        '</div>'
}