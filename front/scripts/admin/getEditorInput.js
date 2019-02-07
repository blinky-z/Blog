function getEditorInput() {
    var title = document.getElementById("titleInput").value;
    var description = document.getElementById("descriptionInput").value;
    var keywords = document.getElementById("keywordsInput").value;
    var content = document.getElementById("contentInput").value;

    var metadata = {description: description, keywords: keywords.split(",")};
    return {title: title, content: content, metadata: metadata};
}