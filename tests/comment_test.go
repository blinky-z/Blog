package tests

import (
	"github.com/blinky-z/Blog/handler/restapi"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/commentService"
	"gotest.tools/assert"
	"net/http"
	"testing"
)

func TestHandleCommentIntegrationTest(t *testing.T) {
	var workingComment models.Comment

	// create post for creating comments
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	assertNiceResponse(t, createPostResponse, http.StatusCreated)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body
	postId := post.ID

	// Step 1: Create Comment
	{
		request := createCommentRequestFactory(postId)

		r := createComment(request)
		assertNiceResponse(t, r, http.StatusCreated)

		resp := decodeResponseWithCommentBody(r.Body)

		workingComment = resp.Body
	}

	// Step 2: Get post with comments and compare received comment with created one
	{
		r := getCertainPost(postId)
		assertNiceResponse(t, r, http.StatusOK)

		resp := decodeResponseWithCertainPostBody(r.Body)
		post := resp.Body

		comments := post.Comments
		assert.Assert(t, len(comments) == 1, "Post should contain comment")

		receivedComment := comments[0].Comment
		if receivedComment != workingComment {
			t.Fatalf("Received comment does not match created one\nCreated comment: %v\nReceived comment: %v",
				workingComment, receivedComment)
		}
	}

	// Step 3: Update created comment
	{
		request := updateCommentRequestFactory()
		r := updateComment(workingComment.ID, request)
		assertNiceResponse(t, r, http.StatusCreated)

		resp := decodeResponseWithCommentBody(r.Body)

		workingComment = resp.Body
	}

	// Step 4: Get post with comments and compare received comment with updated one
	{
		r := getCertainPost(postId)
		assertNiceResponse(t, r, http.StatusOK)

		resp := decodeResponseWithCertainPostBody(r.Body)
		post := resp.Body

		comments := post.Comments

		receivedComment := comments[0].Comment
		if receivedComment != workingComment {
			t.Fatalf("Received comment does not match updated one\nUpdated comment: %v\nReceived comment: %v",
				workingComment, receivedComment)
		}
	}

	// Step 5: Delete updated comments
	{
		r := deleteComment(workingComment.ID)
		assertNiceResponse(t, r, http.StatusOK)
	}

	// Step 6: Get post with comments and assert there are no comments
	{
		r := getCertainPost(postId)
		assertNiceResponse(t, r, http.StatusOK)

		resp := decodeResponseWithCertainPostBody(r.Body)
		post := resp.Body

		comments := post.Comments

		if len(comments) != 0 {
			t.Fatalf("Received post should not contain comments, but was: %v", comments)
		}
	}
}

func TestCreateCommentToNonexistentPost(t *testing.T) {
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body

	deletePost(post.ID)

	request := createCommentRequestFactory(post.ID)
	r := createComment(request)

	assertErrorResponse(t, r, http.StatusBadRequest, restapi.InvalidRequest)
}

func TestReplyToComment(t *testing.T) {
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body

	firstCommentRequest := createCommentRequestFactory(post.ID)
	r := createComment(firstCommentRequest)
	assertNiceResponse(t, r, http.StatusCreated)
	parentComment := decodeResponseWithCommentBody(r.Body).Body

	replyCommentRequest := createCommentWithParentRequestFactory(post.ID, parentComment.ID)
	r = createComment(replyCommentRequest)
	assertNiceResponse(t, r, http.StatusCreated)
	replyComment := decodeResponseWithCommentBody(r.Body).Body

	assert.Assert(t, replyComment.ParentID.Valid == true, "Reply comment should contain parent")
	assert.Equal(t, replyComment.ParentID.String, parentComment.ID,
		"Reply comment should contain proper parent, but was: %s", replyComment.ParentID.String)
}

func TestEnsureReceivedCommentsInAscOrder(t *testing.T) {
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body

	commentsToCreate := 10
	for currentComment := 0; currentComment < commentsToCreate; currentComment++ {
		request := createCommentRequestFactory(post.ID)
		r := createComment(request)
		assertNiceResponse(t, r, http.StatusCreated)
	}

	comments, _ := commentService.GetAllByPostID(db, post.ID)
	for i := 1; i < len(comments); i++ {
		if !comments[i-1].Date.Before(comments[i].Date) {
			t.Fatalf("Comments returned from comment serivce should be sorted in ascending order")
		}
	}
}

func TestEnsureCommentWithNoChildsDeletedFromDB(t *testing.T) {
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body

	request := createCommentRequestFactory(post.ID)
	r := createComment(request)
	assertNiceResponse(t, r, http.StatusCreated)
	comment := decodeResponseWithCommentBody(r.Body).Body

	r = deleteComment(comment.ID)
	assertNiceResponse(t, r, http.StatusOK)

	r = getCertainPost(post.ID)
	assertNiceResponse(t, r, http.StatusOK)
	returnedPost := decodeResponseWithCertainPostBody(r.Body).Body

	comments := returnedPost.Comments
	if len(comments) != 0 {
		t.Fatalf("Post should have no comments, but was: %v", comments)
	}
}

func TestEnsureCommentWithChildsWasNotDeletedButContentReplaced(t *testing.T) {
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body

	parentCommentRequest := createCommentRequestFactory(post.ID)
	r := createComment(parentCommentRequest)
	assertNiceResponse(t, r, http.StatusCreated)
	parentComment := decodeResponseWithCommentBody(r.Body).Body

	replyCommentRequest := createCommentWithParentRequestFactory(post.ID, parentComment.ID)
	r = createComment(replyCommentRequest)
	assertNiceResponse(t, r, http.StatusCreated)

	r = deleteComment(parentComment.ID)
	assertNiceResponse(t, r, http.StatusOK)

	r = getCertainPost(post.ID)
	assertNiceResponse(t, r, http.StatusOK)
	actualPost := decodeResponseWithCertainPostBody(r.Body).Body

	comments := actualPost.Comments
	assert.Assert(t, len(comments) == 1, "Returned post should contain parent comment")

	actualParentComment := comments[0]
	assert.Assert(t, len(actualParentComment.Childs) == 1, "Parent comment should contain reply comment")

	if actualParentComment.Content != commentService.DeletedCommentContent {
		t.Fatalf("Parent comment's content should be replaced with special deletion message, but was: %v", actualParentComment)
	}
}
