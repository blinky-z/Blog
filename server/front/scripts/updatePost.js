function updatePost() {
    var postID = new window.URLSearchParams(window.location.search).get('id');

    var title = document.getElementById("titleInput").value;
    var content = document.getElementById("contentInput").value;

    var post = {title: title, content: content};

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
            error: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}