function getCommentCreationInput() {
    var author = document.getElementById("comment-author-input").value;
    var content = document.getElementById("comment-content-textarea").value;

    return {Author: author, Content: content};
}