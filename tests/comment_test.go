package tests

import (
	"github.com/blinky-z/Blog/handler/restApi"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/commentService"
	"net/http"
	"strings"
	"testing"
)

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

		if workingComment.ParentID != sourceComment.ParentCommentID {
			t.Fatalf("Created comment Parent ID does not match source comment one\n"+
				"Created comment: %v\n Source comment: %v", workingComment.ParentID, sourceComment.ParentCommentID)
		}

		if workingComment.Content != sourceComment.Content {
			t.Fatalf("Created comment Content does not match source comment one\n"+
				"Created comment: %v\n Source comment: %v", workingComment.Content, sourceComment.Content)
		}
	}

	// Step 2: Get post with comments and compare received comment with created one
	{
		var response ResponsePostWithComments

		r := getPost(workingComment.PostID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusOK)

		decodePostWithCommentsResponse(r.Body, &response)

		comments := response.Body.Comments
		receivedComment := comments[0]
		if receivedComment.Comment != workingComment {
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
		checkNiceResponse(r, http.StatusCreated)

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
		var response ResponsePostWithComments

		r := getPost(workingComment.PostID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusOK)

		decodePostWithCommentsResponse(r.Body, &response)

		comments := response.Body.Comments
		receivedComment := comments[0]
		if receivedComment.Comment != workingComment {
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
		var response ResponsePostWithComments

		r := getPost(workingComment.PostID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		checkNiceResponse(r, http.StatusOK)

		decodePostWithCommentsResponse(r.Body, &response)

		comments := response.Body.Comments
		if len(comments) != 0 {
			t.Fatalf("Received post has comment, but comment should be deleted. Comments: %v",
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

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidCommentContent)
}

func TestCreateCommentWithTooLongContent(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = strings.Repeat("a", restApi.MaxCommentContentLen*2)

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidCommentContent)
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

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidUsername)
}

func TestCreateCommentWithTooShortAuthor(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Content = "test content"
	comment.Author = strings.Repeat("2", restApi.MinUsernameLen-1)

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidUsername)
}

func TestCreateCommentWithTooLongAuthor(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Content = "test content"
	comment.Author = strings.Repeat("1", restApi.MaxUsernameLen*2)

	r := createComment(&comment)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidUsername)
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

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidRequest)
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

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidRequest)
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

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidCommentContent)
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
	newComment.Content = strings.Repeat("s", restApi.MaxCommentContentLen*2)
	r = updateComment(createdComment.ID, &newComment)

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidCommentContent)
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

	comments, _ := commentService.GetAllByPostId(db, comment.PostID)
	for i := 1; i < len(comments); i++ {
		if !comments[i-1].Date.Before(comments[i].Date) {
			t.Fatalf("Received comments from commentService should be sorted in ascending order, but "+
				"comment[%d] date goes after comment[%d] date\nComment[%d]: %+v\nComment[%d]: %+v\nReceived comments: %+v",
				i-1, i, i-1, comments[i-1], i, comments[i], comments)
		}
	}
}

func TestEnsureCommentWithNoChildsDeletedFromDB(t *testing.T) {
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

	r = getPost(createdComment.PostID)
	checkNiceResponse(r, http.StatusOK)

	var response ResponsePostWithComments
	decodePostWithCommentsResponse(r.Body, &response)

	comments := response.Body.Comments
	if len(comments) != 0 {
		t.Fatalf("Post should have no comments. Comments: %v", comment)
	}
}

func TestEnsureCommentWithChildsWasNotDeletedButHasDeletedContent(t *testing.T) {
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

	comment = setCommentRequestParentID(comment, createdComment.ID)
	r = createComment(&comment)
	checkNiceResponse(r, http.StatusCreated)

	r = deleteComment(createdComment.ID)
	checkNiceResponse(r, http.StatusOK)

	r = getPost(createdComment.PostID)
	checkNiceResponse(r, http.StatusOK)

	var response ResponsePostWithComments
	decodePostWithCommentsResponse(r.Body, &response)

	comments := response.Body.Comments
	if len(comments) == 0 {
		t.Fatalf("Level 0 comment with childs should not be deleted but should has deleted comment")
	}
	if comments[0].Content != commentService.DeletedCommentContent {
		t.Fatalf("Level 0 comment with childs should have deleted body\nChild comment: %v", comments[0])
	}
}

func TestDeleteChildWithoutRepliesAndEnsureCommentHasNoChilds(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = "test content"

	r := createComment(&comment)
	checkNiceResponse(r, http.StatusCreated)
	createdComment := getCommentFromResponseBody(r)
	comment = setCommentRequestParentID(comment, createdComment.ID)
	r = createComment(&comment)
	checkNiceResponse(r, http.StatusCreated)
	replyComment := getCommentFromResponseBody(r)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	r = deleteComment(replyComment.ID)
	checkNiceResponse(r, http.StatusOK)

	var response ResponsePostWithComments

	r = getPost(createdComment.PostID)
	checkNiceResponse(r, http.StatusOK)

	decodePostWithCommentsResponse(r.Body, &response)

	comments := response.Body.Comments
	receivedComment := comments[0]
	if len(receivedComment.Childs) != 0 {
		t.Fatalf("Comment should have no reply comments\nReply comments: %v", receivedComment.Childs)
	}
}

func TestDeleteChildWithRepliesAndEnsureChildWasNotDeletedButHaveDeletedContent(t *testing.T) {
	comment := testCreateCommentFactory()
	comment.Author = "test author"
	comment.Content = "test content"

	r := createComment(&comment)
	checkNiceResponse(r, http.StatusCreated)
	createdComment := getCommentFromResponseBody(r)

	comment = setCommentRequestParentID(comment, createdComment.ID)
	r = createComment(&comment)
	checkNiceResponse(r, http.StatusCreated)
	replyComment1 := getCommentFromResponseBody(r)

	comment = setCommentRequestParentID(comment, replyComment1.ID)
	r = createComment(&comment)
	checkNiceResponse(r, http.StatusCreated)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	r = deleteComment(replyComment1.ID)
	checkNiceResponse(r, http.StatusOK)

	r = getPost(createdComment.PostID)
	checkNiceResponse(r, http.StatusOK)

	var response ResponsePostWithComments
	decodePostWithCommentsResponse(r.Body, &response)

	comments := response.Body.Comments
	if len(comments[0].Childs) == 0 {
		t.Fatalf("Child Comment with childs should not be deleted, but only have deleted content")
	}

	child1 := comments[0].Childs[0]
	if child1.Content != commentService.DeletedCommentContent {
		t.Fatalf("Child Comment with childs should have deleted body\nChild comment: %v", child1)
	}
}
