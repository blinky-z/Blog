package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/commentService"
	"github.com/blinky-z/Blog/handler/api"
	"github.com/blinky-z/Blog/models"
	"io"
	"net/http"
	"strings"
	"testing"
)

type ResponseComment struct {
	Error api.PostErrorCode
	Body  models.Comment
}

func testCreateCommentFactory() models.CommentCreateRequest {
	post := testPostFactory()
	post.Title = "post for testing comments"
	post.Content = "post for testing comments"
	r := createPost(post)

	var responseCreatePost ResponseSinglePost
	decodeSinglePostResponse(r.Body, &responseCreatePost)
	createdPost := responseCreatePost.Body

	var testComment models.CommentCreateRequest

	testComment.PostID = createdPost.ID

	return testComment
}

func testUpdateCommentFactory() models.CommentUpdateRequest {
	var testComment models.CommentUpdateRequest

	return testComment
}

func getCommentFromResponseBody(r *http.Response) models.Comment {
	var response ResponseComment
	decodeCommentResponse(r.Body, &response)
	comment := response.Body

	return comment
}

func setCommentRequestParentID(comment models.CommentCreateRequest, parentID string) models.CommentCreateRequest {
	comment.ParentID.Valid = true
	comment.ParentID.String = parentID

	return comment
}

// API for encoding and decoding messages

func decodeCommentResponse(responseBody io.ReadCloser, response *ResponseComment) {
	err := json.NewDecoder(responseBody).Decode(&response)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

// -----------
// API for sending comments handling http requests

func createComment(message interface{}) *http.Response {
	return sendCommentHandleMessage("POST", "http://"+Address+"/api/comments", message)
}

func updateComment(id string, message interface{}) *http.Response {
	return sendCommentHandleMessage("PUT", "http://"+Address+"/api/comments/"+id, message)
}

func deleteComment(id string) *http.Response {
	return sendCommentHandleMessage("DELETE", "http://"+Address+"/api/comments/"+id, "")
}

func sendCommentHandleMessage(method, address string, message interface{}) *http.Response {
	var response *http.Response

	switch method {
	case "POST":
		encodedMessage := encodeMessage(message)

		request, err := http.NewRequest("POST", address, bytes.NewReader(encodedMessage))
		if err != nil {
			panic(fmt.Sprintf("Can not create POST comment request. Error: %s", err))
		}
		request.Header.Set("Content-Type", "application/json")

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send POST comment request. Error: %s", err))
		}
	case "PUT":
		encodedMessage := encodeMessage(message)

		request, err := http.NewRequest("PUT", address, bytes.NewReader(encodedMessage))
		if err != nil {
			panic(fmt.Sprintf("Can not create UPDATE comment request. Error: %s", err))
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		request.AddCookie(ctxCookie)

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send UPDATE comment request. Error: %s", err))
		}
	case "DELETE":
		request, err := http.NewRequest("DELETE", address, strings.NewReader(""))
		if err != nil {
			panic(fmt.Sprintf("Can not create DELETE comment request. Error: %s", err))
		}
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		request.AddCookie(ctxCookie)

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send DELETE comment request. Error: %s", err))
		}
	}

	return response
}

// -----------
// tests

func TestHandleCommentIntegrationTest(t *testing.T) {
	var workingComment models.Comment

	// Step 1: Create Comment
	{
		var response ResponseComment

		sourceComment := testCreateCommentFactory()
		sourceComment.Author = "test author"
		sourceComment.Content = "test comment1"

		r := createComment(&sourceComment)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusCreated)

		decodeCommentResponse(r.Body, &response)

		workingComment = response.Body

		if workingComment.PostID != sourceComment.PostID {
			t.Fatalf("Created comment Post ID does not match source comment one\n"+
				"Created comment: %v\n Source comment: %v", workingComment.PostID, sourceComment.PostID)
		}

		if workingComment.ParentID != sourceComment.ParentID {
			t.Fatalf("Created comment Parent ID does not match source comment one\n"+
				"Created comment: %v\n Source comment: %v", workingComment.ParentID, sourceComment.ParentID)
		}

		if workingComment.Content != sourceComment.Content {
			t.Fatalf("Created comment Content does not match source comment one\n"+
				"Created comment: %v\n Source comment: %v", workingComment.Content, sourceComment.Content)
		}
	}

	// Step 2: Get post with comments and compare received comment with created one
	{
		var response ResponseSinglePostWithComments

		r := getPost(workingComment.PostID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusOK)

		decodeSinglePostWithCommentsResponse(r.Body, &response)

		comments := response.Body.Comments
		receivedComment := comments[0]
		if receivedComment != workingComment {
			t.Fatalf("Received comment does not match created one\nCreated comment: %v\nReceived comment: %v",
				workingComment, receivedComment)
		}
	}

	// Step 3: Update created comment
	{
		var response ResponseComment

		newPost := testUpdateCommentFactory()
		newPost.Content = "new test comment1 content"

		r := updateComment(workingComment.ID, &newPost)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusOK)

		decodeCommentResponse(r.Body, &response)

		updatedComment := response.Body
		if updatedComment.Content != newPost.Content {
			t.Fatalf("Updated comment Content does not match source one\nUpdated comment: %v\n Source comment: %v",
				updatedComment.Content, newPost.Content)
		}

		workingComment.Content = updatedComment.Content
	}

	// Step 4: Get post with comments and compare received comment with updated one
	{
		var response ResponseSinglePostWithComments

		r := getPost(workingComment.PostID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusOK)

		decodeSinglePostWithCommentsResponse(r.Body, &response)

		comments := response.Body.Comments
		receivedComment := comments[0]
		if receivedComment != workingComment {
			t.Fatalf("Received comment does not match updated one\nUpdated comment: %v\n Received comment: %v",
				workingComment, receivedComment)
		}
	}

	// Step 5: Delete updated comments
	{
		r := deleteComment(workingComment.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusOK)
	}

	// Step 6: Get post with comments and ensure that there's no comments
	{
		var response ResponseSinglePostWithComments

		r := getPost(workingComment.PostID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusOK)

		decodeSinglePostWithCommentsResponse(r.Body, &response)

		comments := response.Body.Comments
		receivedComment := comments[0]
		if receivedComment.Deleted != true {
			t.Fatalf("Received post has undeleted comment, but comment should be marked as deleted. Comments: %v",
				comments)
		}
	}
}

func TestCreateCommentWithEmptyContent(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = ""

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidContent)
}

func TestCreateCommentWithTooLongContent(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = strings.Repeat("a", api.MaxCommentContentLen*2)

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidContent)
}

func TestCreateCommentWithEmptyAuthor(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Content = "test content"
	comment.Author = ""

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidLogin)
}

func TestCreateCommentWithTooShortAuthor(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Content = "test content"
	comment.Author = strings.Repeat("2", api.MinAuthorLen-1)

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidLogin)
}

func TestCreateCommentWithTooLongAuthor(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Content = "test content"
	comment.Author = strings.Repeat("1", api.MaxAuthorLen*2)

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidLogin)
}

func TestCreateCommentToNonexistentPost(t *testing.T) {
	comment := testCreateCommentFactory()
	deletePost(comment.PostID)

	comment.Author = "test author"
	comment.Content = "test content"

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusNotFound, api.NoSuchPost)
}

func TestReplyToComment(t *testing.T) {
	firstComment := testCreateCommentFactory()

	firstComment.Author = "test author1"
	firstComment.Content = "test content1"

	r := createComment(&firstComment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	firstCreatedComment := getCommentFromResponseBody(r)

	replyComment := testCreateCommentFactory()
	replyComment.PostID = firstComment.PostID
	replyComment = setCommentRequestParentID(replyComment, firstCreatedComment.ID)
	replyComment.Author = "test author2"
	replyComment.Content = "test content2"

	r = createComment(&replyComment)

	checkNiceResponse(r, http.StatusCreated)

	replyCreatedComment := getCommentFromResponseBody(r)

	if !replyCreatedComment.ParentID.Valid || replyCreatedComment.ParentID.String != firstCreatedComment.ID {
		t.Fatalf("Added comment should be a child of firstly created comment. Added comment's parent id: %s\n"+
			"First comment's id: %s", replyCreatedComment.ParentID.String, firstCreatedComment.ID)
	}
}

func TestReplyToCommentBelongsToOtherPost(t *testing.T) {
	firstComment := testCreateCommentFactory()

	firstComment.Author = "test author1"
	firstComment.Content = "test content1"

	r := createComment(&firstComment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	firstCreatedComment := getCommentFromResponseBody(r)

	replyComment := testCreateCommentFactory()
	replyComment = setCommentRequestParentID(replyComment, firstCreatedComment.ID)
	replyComment.Author = "test author2"
	replyComment.Content = "test content2"

	r = createComment(&replyComment)

	checkErrorResponse(r, http.StatusNotFound, api.NoSuchComment)
}

func TestUpdateCommentWithEmptyContent(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = "test content"

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	createdComment := getCommentFromResponseBody(r)

	newComment := testUpdateCommentFactory()
	newComment.Content = ""
	r = updateComment(createdComment.ID, &newComment)

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidContent)
}

func TestUpdateCommentWithTooLongContent(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = "test content"

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	createdComment := getCommentFromResponseBody(r)

	newComment := testUpdateCommentFactory()
	newComment.Content = strings.Repeat("s", api.MaxCommentContentLen*2)
	r = updateComment(createdComment.ID, &newComment)

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidContent)
}

func TestEnsureReceivedCommentsInAscOrder(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = "test content"

	r := createComment(&comment)
	createComment(&comment)
	createComment(&comment)
	createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	comments, _ := commentService.GetComments(env, comment.PostID)
	for i := 1; i < len(comments); i++ {
		if !comments[i-1].Date.Before(comments[i].Date) {
			t.Fatalf("Received comments from commentService should be sorted in ascending order, but "+
				"comment[%d] date goes after comment[%d] date\nComment[%d]: %+v\nComment[%d]: %+v\nReceived comments: %+v",
				i-1, i, i-1, comments[i-1], i, comments[i], comments)
		}
	}
}

func TestEnsureDeletedCommentContentIsRemoved(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = "test content"

	r := createComment(&comment)
	checkNiceResponse(r, http.StatusCreated)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	createdComment := getCommentFromResponseBody(r)

	r = deleteComment(createdComment.ID)
	checkNiceResponse(r, http.StatusOK)

	var response ResponseSinglePostWithComments

	r = getPost(createdComment.PostID)
	checkNiceResponse(r, http.StatusOK)

	decodeSinglePostWithCommentsResponse(r.Body, &response)

	comments := response.Body.Comments
	receivedComment := comments[0]
	if receivedComment.Content != api.DeletedCommentContent {
		t.Fatalf("Deleted comment has content\nComment: %v", receivedComment)
	}
}
