function createCommentChild(createCommentResponse) {
    var commentChild = document.createElement('li');
    commentChild.setAttribute('class', 'comment');
    commentChild.setAttribute('id', createCommentResponse.id);
    commentChild.innerHTML = '                                <div class="comment-block">\n' +
        '                                    <div class="comment-header">\n' +
        '                                        <div class="username"><a href="#" onclick="return false;" rel="nofollow"><b></b></a></div>\n' +
        '                                        <div class="creation-time"><a href="#" onclick="return false;" rel="nofollow"></a></div>\n' +
        '                                    </div>\n' +
        '                                    <p class="comment-content"></p>\n' +
        '                                    <div class="comment-actions">\n' +
        '                                        <div class="reply-comment-button">\n' +
        '                                            <a class="comment-reply" data-comment-id=""\n' +
        '                                               href="#" onclick="showCommentReplyBox(this); return false;">Reply</a>\n' +
        '                                        </div>\n' +
        '                                    </div>\n' +
        '                                </div>\n';


    commentChild.getElementsByClassName('username')[0].getElementsByTagName('a')[0]
        .getElementsByTagName('b')[0].innerHTML = createCommentResponse.author;
    commentChild.getElementsByClassName('creation-time')[0].getElementsByTagName('a')[0]
        .innerHTML = convertToGoTimeFormat(createCommentResponse.date);
    commentChild.getElementsByClassName('comment-content')[0].innerHTML = createCommentResponse.content;
    commentChild.getElementsByClassName('comment-reply')[0].setAttribute('data-comment-id',
        createCommentResponse.id);

    return commentChild;
}