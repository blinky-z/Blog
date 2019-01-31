function post() {
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
            url: '/api/posts',
            type: 'POST',
            contentType: 'application/json',
            data: encodedPost,
            beforeSend: function (xhr) {
                xhr.setRequestHeader('Authorization', `bearer ${token}`);
            },
            success: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                var createdPost = response.body;

                alert("Post successfully created");
                var postID = createdPost.id;
                window.location.replace(`/posts/${postID}`)
            },
            error: function (data, textStatus, jqXHR) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}
