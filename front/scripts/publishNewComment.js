function publishNewComment() {
    var postID = window.location.pathname.substr(window.location.pathname.lastIndexOf('/') + 1);

    var commentCreateRequest = {PostID: '', ParentID: null, Author: '', Content: ''};

    commentCreateRequest.PostID = postID;
    commentCreateRequest.Author = document.getElementsByClassName("comment-author-input")[0].value;
    commentCreateRequest.Content = document.getElementsByClassName("comment-content-textarea")[0].value;

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

                var globalCommentsList = document.getElementsByClassName("comments-list")[0]
                    .getElementsByTagName('ul')[0];

                var newComment = createCommentChild(createCommentResponse);

                var commentWithChilds = document.createElement('ul');
                commentWithChilds.appendChild(newComment);

                globalCommentsList.appendChild(commentWithChilds);

                alert('Comment successfully created');

                // clear content and set default size of content form
                document.getElementsByClassName('autoExpand comment-content-textarea')[0].value = '';
                document.getElementsByClassName('autoExpand comment-content-textarea')[0]
                    .setAttribute('rows', document.getElementsByClassName('comment-content-textarea')[0]
                        .getAttribute('data-min-rows'));
            },
            error: function (jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}