function updatePost() {
    var postID = new window.URLSearchParams(window.location.search).get('id');

    var post = getEditorInput();
    var encodedPost = JSON.stringify(post);

    var token = sessionStorage.getItem("token");
    if (token === null || token === "") {
        alert("Please Log In first");
        return
    }
    $.ajax(
        {
            url: `/api/posts/${postID}`,
            type: 'PUT',
            contentType: 'application/json',
            data: encodedPost,
            beforeSend: function (xhr) {
                xhr.setRequestHeader('Authorization', `bearer ${token}`);
            },
            success: function (data, textStatus, jqXHR) {
                window.location.replace(`/posts/${postID}`);
            },
            error: function (jqXHR, textStatus, error) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}