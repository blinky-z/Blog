// берем новый контент из созданного textarea и обновляем пост новым контентом
function closeSaveButton(comment) {
    var saveCommentButton = comment.getElementsByClassName('save-comment-button')[0];
    saveCommentButton.parentNode.removeChild(saveCommentButton);
}

function saveComment(saveCommentButtonLink) {
    var commentID = saveCommentButtonLink.getAttribute('data-comment-id');
    var comment = document.getElementById(commentID);

    var token = sessionStorage.getItem("token");
    if (token === null || token === "") {
        alert("Please Log In first");
        return
    }

    var editCommentTextarea = comment.getElementsByClassName('autoExpand comment-content editCommentTextarea')[0];

    var commentUpdateRequest = {Content: ''};
    commentUpdateRequest.Content = editCommentTextarea.value;

    var encodedData = JSON.stringify(commentUpdateRequest);

    $.ajax(
        {
            url: `/api/comments/${commentID}`,
            type: 'PUT',
            data: encodedData,
            beforeSend: function (xhr) {
                xhr.setRequestHeader('Authorization', `bearer ${token}`);
            },
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                var updatedComment = response.body;

                var newContent = document.createElement('p');
                newContent.className = 'comment-content';
                newContent.innerHTML = updatedComment.content;

                editCommentTextarea.parentNode.replaceChild(newContent, editCommentTextarea);

                closeSaveButton(comment);

                alert('Комментарий был успешно обновлен');

                comment.getElementsByClassName('comment-edit')[0].setAttribute('data-opened', 'false');
            },
            error: function (jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}

// меняем контент на textarea и добавляем кнопку сохранения
// если изменение уже открыто, то просто возвращаем прежний контент и убираем textarea
function editComment(editCommentButtonLink) {
    var isEditTextareaOpened = editCommentButtonLink.getAttribute('data-opened');

    var commentID = editCommentButtonLink.getAttribute('data-comment-id');
    var comment = document.getElementById(commentID);

    if (isEditTextareaOpened === "false") {
        var pContent = comment.getElementsByClassName("comment-content")[0];

        var editTextarea = document.createElement('textarea');
        editTextarea.setAttribute('data-old-content', pContent.innerHTML);
        editTextarea.className = 'autoExpand comment-content editCommentTextarea';
        editTextarea.value = pContent.innerHTML;

        var rows = Math.ceil((pContent.scrollHeight) / parseInt($(pContent).css("line-height")));
        editTextarea.setAttribute('rows', rows.toString());
        editTextarea.setAttribute('data-min-rows', rows.toString());

        pContent.parentNode.replaceChild(editTextarea, pContent);

        var commentActionsList = comment.getElementsByClassName('comment-actions')[0].getElementsByTagName('ul')[0];
        var saveNewContentButton = document.createElement('li');
        saveNewContentButton.className = 'save-comment-button';
        saveNewContentButton.innerHTML = '<a class="comment-save" data-comment-id="' + commentID + '"\n' +
            '                               href="#" onclick="saveComment(this); return false;">Save</a>'

        commentActionsList.appendChild(saveNewContentButton);

        editCommentButtonLink.setAttribute('data-opened', 'true');
    } else {
        var oldContent = document.createElement('p');
        oldContent.className = 'comment-content';
        var currentTextarea = comment.getElementsByClassName('autoExpand comment-content editCommentTextarea')[0];
        oldContent.innerHTML = currentTextarea.getAttribute('data-old-content');

        currentTextarea.parentNode.replaceChild(oldContent, currentTextarea);

        closeSaveButton(comment);

        editCommentButtonLink.setAttribute('data-opened', 'false');
    }
}