function deleteComment(deleteCommentButtonLink) {
    var commentID = deleteCommentButtonLink.getAttribute('data-comment-id');

    var token = sessionStorage.getItem("token");
    if (token === null || token === "") {
        alert("Please Log In first");
        return
    }

    $.ajax(
        {
            url: `/api/comments/${commentID}`,
            type: 'DELETE',
            beforeSend: function (xhr) {
                xhr.setRequestHeader('Authorization', `bearer ${token}`);
            },
            success: function (data, textStatus, jqXHR) {
                document.getElementById(commentID).getElementsByClassName('comment-content')[0].innerHTML =
                    'Содержимое этого комментария было удалено';

                alert('Комментарий был успешно удален');
            },
            error: function (jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );

}