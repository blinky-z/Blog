function post() {
    var title = document.getElementById("titleInput").value;
    var content = document.getElementById("contentInput").value;

    var post = {title: title, content: content};

    var encodedPost = JSON.stringify(post);
    console.log(post);

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (this.readyState === 4) {
            var responseBody = JSON.parse(this.responseText);
            if (this.status === 202) {
                var createdPost = responseBody.body;

                console.log(createdPost);
            } else {
                var errorMessage = responseBody.error;

                console.log(errorMessage);
            }
        }
    };

    request.open("POST", "http://localhost:8080/posts", true);
    request.setRequestHeader("Content-type", "application/json");

    var token = sessionStorage.getItem("token");

    if (token === null || token === "") {
        console.log("Please Log In first");
        return
    }
    request.setRequestHeader("Authorization", "bearer " + token);
    request.send(encodedPost);
}
