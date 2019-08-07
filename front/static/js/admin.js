var cutDelimiter = '&lt;cut&gt;';
var editor;

$(document).ready(function () {
    if ($('#editSection').length) {
        editor = new tui.Editor({
            el: document.querySelector('#editSection'),
            initialEditType: 'markdown',
            previewStyle: 'tab',
            height: '600px',
            usageStatistics: false,
            placeholder: 'write here...'
        });

        var contentTemp = $("#contentTemp");
        var content = contentTemp.html();
        if (content !== "") {
            editor.setHtml(content);
        }
        contentTemp.remove();
    }
});

function getEditorInput() {
    var title = $("#title").val();
    var content = editor.getHtml();

    var indexOfCut = content.indexOf(cutDelimiter);
    if (indexOfCut === -1) {
        alert("Please insert cut delimiter");
        return -1;
    }
    var snippet = content.substring(0, indexOfCut);
    content = content.substring(indexOfCut + cutDelimiter.length);

    var keywords = $("#metaKeywords").val();
    if (keywords !== "") {
        keywords = keywords.split(",");
    } else {
        keywords = [];
    }

    var metadata = {
        description: $("#metaDescription").val(),
        keywords: keywords
    };

    var tags = $("#tags").val();
    if (tags !== "") {
        tags = tags.split(",");
    } else {
        tags = [];
    }

    console.log({
        title: title,
        snippet: snippet,
        content: content,
        metadata: metadata,
        tags: tags
    });

    return {
        title: title,
        snippet: snippet,
        content: content,
        metadata: metadata,
        tags: tags
    };
}

function publishPost(action, domain) {
    var postID = $(action).attr("data-id");

    var post = getEditorInput();
    if (post === -1) {
        return
    }
    var encodedPost = JSON.stringify(post);

    // var token = sessionStorage.getItem("token");
    // if (token === null || token === "") {
    //     alert("Please Log In first");
    //     return
    // }

    var url, type;
    if (postID === "") {
        url = `/api/posts`;
        type = 'POST';
    } else {
        url = `/api/posts/${postID}`;
        type = 'PUT';
    }

    $.ajax(
        {
            url: url,
            type: type,
            contentType: 'application/json',
            data: encodedPost,
            beforeSend: function (xhr) {
                // xhr.setRequestHeader('Authorization', `bearer ${token}`);
            },
            success: function (data, textStatus, jqXHR) {
                alert("Post published");

                var response = JSON.parse(jqXHR.responseText);
                var createdPost = response.body;

                var postID = createdPost.ID;
                window.location.href = `${domain}/posts/${postID}`
            },
            statusCode: {
                401: function () {
                    alert("Please Log In first");
                }
            },
            error: function (jqXHR, textStatus, errorThrown) {
                alert(errorThrown);
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}

function deletePost(action) {
    var actions = $(action).parent();
    console.log(actions);
    var postID = actions.attr("data-id");
    console.log(postID);

    $.ajax(
        {
            url: `/api/posts/${postID}`,
            type: 'DELETE',
            beforeSend: function (xhr) {
                // xhr.setRequestHeader('Authorization', `bearer ${token}`);
            },
            success: function (data, textStatus, jqXHR) {
                alert("Post deleted");
                document.location.reload()
            },
            statusCode: {
                401: function () {
                    alert("Please Log In first");
                }
            },
            error: function (jqXHR, textStatus, errorThrown) {
                var response = JSON.parse(jqXHR.responseText);
                alert(response.error)
            }
        }
    );
}