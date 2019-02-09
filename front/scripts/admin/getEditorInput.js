function getCookie(name) {
    var pair = document.cookie.split('; ').find(x => x.startsWith(name + '='));
    if (pair)
        return pair.split('=')[1]
}

function getEditorInput() {
    var title = document.getElementById("titleInput").value;
    var description = document.getElementById("descriptionInput").value;
    var keywords = document.getElementById("keywordsInput").value;
    var snippet = document.getElementById("snippetInput").value;
    var content = document.getElementById("contentInput").value;
    var author = getCookie('Login');

    var metadata = {description: description, keywords: keywords.split(",")};
    return {title: title, snippet: snippet, content: content, author: author, metadata: metadata};
}