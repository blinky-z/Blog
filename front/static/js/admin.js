var cutDelimiter = '&lt;cut&gt;';
var editor;
var tagsInputTagify;
const editorTextBackupKey = "editor-text";

// initialize editor section: create tui-editor and add available tags to tagify suggestions
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
        var postID = contentTemp.attr("data-postID");
        editor.setHtml(contentTemp.html());
        contentTemp.remove();

        var tagsInput = document.querySelector("#tags");
        var allTags = tagsInput.getAttribute("data-allTags");
        if (allTags !== "") {
            allTags = allTags.split(",")
        } else {
            allTags = []
        }

        tagsInputTagify = new Tagify(tagsInput, {
            whitelist: allTags,
        });

        window.setInterval(function () {
            localStorage.setItem(editorTextBackupKey + postID, editor.getHtml())
        }, 5000);
    }
});

function restoreCache(action) {
    var postID = $(action).attr("data-id");

    var backupText = localStorage.getItem(editorTextBackupKey + postID);
    if (backupText != null) {
        editor.setHtml(backupText)
    } else {
        alert("Cache is empty!")
    }
}

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

    var tagsTagify = tagsInputTagify.value;
    var tags = [];
    tagsTagify.forEach(function (elem) {
        tags.push(elem.value);
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

    localStorage.setItem(editorTextBackupKey + postID, editor.getHtml());

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

                var createdPostID = createdPost.ID;

                // avoid situation when post might be published again
                if (postID === "") {
                    window.location.replace(`${domain}/posts/${createdPostID}`)
                } else {
                    window.location.href = `${domain}/posts/${createdPostID}`
                }
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
    var result = confirm("You sure you want to delete this post?");
    if (result) {
        var actions = $(action).parent();
        var postID = actions.attr("data-id");

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
}

function createTag() {
    var tagName = prompt("Enter a tag name");
    if (tagName == null) {
        return
    }

    var data = {Name: tagName};

    $.ajax(
        {
            url: '/api/tags',
            type: 'POST',
            data: JSON.stringify(data),
            beforeSend: function (xhr) {
                // xhr.setRequestHeader('Authorization', `bearer ${token}`);
            },
            success: function (data, textStatus, jqXHR) {
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

function editTag(action) {
    var actions = $(action).parent();
    var tagID = actions.attr("data-id");
    var tagName = actions.attr("data-name");

    var newTagName = prompt("Enter a new tag name", tagName);
    if (newTagName == null) {
        return false
    }

    var data = {Name: newTagName};

    $.ajax(
        {
            url: `/api/tags/${tagID}`,
            type: 'PUT',
            data: JSON.stringify(data),
            beforeSend: function (xhr) {
                // xhr.setRequestHeader('Authorization', `bearer ${token}`);
            },
            success: function (data, textStatus, jqXHR) {
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

function deleteTag(action) {
    var result = confirm("You sure you want to delete this tag?");
    if (result) {
        var actions = $(action).parent();
        var tagID = actions.attr("data-id");

        $.ajax(
            {
                url: `/api/tags/${tagID}`,
                type: 'DELETE',
                beforeSend: function (xhr) {
                    // xhr.setRequestHeader('Authorization', `bearer ${token}`);
                },
                success: function (data, textStatus, jqXHR) {
                    alert("Tag deleted");
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
}