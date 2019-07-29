package tests

import (
	"github.com/blinky-z/Blog/handler/restApi"
	"github.com/blinky-z/Blog/models"
	"gotest.tools/assert"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestHandlePostIntegrationTest(t *testing.T) {
	workingPost := testPostFactory()

	// Step 1: Create Post
	{
		var response ResponsePost

		sourcePost := testPostFactory()

		r := createPost(sourcePost)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodePostResponse(r.Body, &response)
		if r.StatusCode != http.StatusCreated {
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		workingPost = response.Body

		if workingPost.Title != sourcePost.Title {
			t.Fatalf("Created post title does not match source post one\nCreated post: %v\n Source post: %v",
				workingPost.Title, sourcePost.Title)
		}

		if workingPost.Content != sourcePost.Content {
			t.Fatalf("Created post content does not match source post one\nCreated post: %v\n Source post: %v",
				workingPost.Content, sourcePost.Content)
		}
	}

	// Step 2: Get created post and compare it with returned in prev step one
	{
		var response ResponsePost

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodePostResponse(r.Body, &response)
		if r.StatusCode != http.StatusOK {
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		receivedPost := response.Body

		if !comparePosts(receivedPost, workingPost) {
			t.Fatalf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
				receivedPost, workingPost)
		}
	}

	// Step 3: Update created post
	{
		var response ResponsePost

		newPost := workingPost
		newPost.Title = "newTitle"
		newPost.Content = "NewContent"

		r := updatePost(workingPost.ID, newPost)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodePostResponse(r.Body, &response)
		if r.StatusCode != http.StatusOK {
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		updatedPost := response.Body
		if !comparePosts(updatedPost, newPost) {
			t.Fatalf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
				updatedPost, newPost)
		}

		workingPost = updatedPost
	}

	// Step 4: Get Updated post
	{
		var response ResponsePost

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodePostResponse(r.Body, &response)
		if r.StatusCode != http.StatusOK {
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		receivedPost := response.Body

		if !comparePosts(receivedPost, workingPost) {
			t.Fatalf("Received post does not match updated post\nReceived post: %v\n Updated post: %v",
				receivedPost, workingPost)
		}
	}

	// Step 5: Delete updated post
	{
		r := deletePost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		if r.StatusCode != http.StatusOK {
			var response ResponsePost

			decodePostResponse(r.Body, &response)

			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}
	}

	// Step 6: Get deleted post
	{
		var response ResponsePostWithComments

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodePostWithCommentsResponse(r.Body, &response)
		if r.StatusCode != http.StatusNotFound {
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}
	}
}

func TestCreatePostWithInvalidRequestBody(t *testing.T) {
	message := `{"bad request body"}`

	r := createPost(message)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.BadRequestBody)
}

func TestCreatePostWithNullTitle(t *testing.T) {
	message := map[string]interface{}{
		"title":   "",
		"content": "Content1 Content2 Content3",
	}

	r := createPost(message)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostTitle)
}

func TestCreatePostWithTooLongTitle(t *testing.T) {
	message := map[string]interface{}{
		"title":   strings.Repeat("a", restApi.MaxPostTitleLen*2),
		"content": "Content1 Content2 Content3",
	}

	r := createPost(message)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostTitle)
}

func TestCreatePostWithNullContent(t *testing.T) {
	post := testPostFactory()
	post.Content = ""

	r := createPost(post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostContent)
}

func TestGetPostWithInvalidID(t *testing.T) {
	r := getPost("post1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidID)
}

func TestGetNonexistentPost(t *testing.T) {
	deletePost("0")

	r := getPost("0")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusNotFound, restApi.NoSuchPost)
}

func TestUpdatePostWithInvalidRequestBody(t *testing.T) {
	message := `{"bad request body":"asd"}`

	r := updatePost("1", message)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.BadRequestBody)
}

func TestUpdatePostWithNullTitle(t *testing.T) {
	post := testPostFactory()
	post.Title = ""

	r := updatePost("1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostTitle)
}

func TestUpdatePostWithTooLongTitle(t *testing.T) {
	post := testPostFactory()
	post.Title = strings.Repeat("a", restApi.MaxPostTitleLen*2)

	r := updatePost("1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostTitle)
}

func TestUpdatePostWithNullContent(t *testing.T) {
	post := testPostFactory()
	post.Content = ""

	r := updatePost("1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostContent)
}

func TestUpdatePostNonexistentPost(t *testing.T) {
	deletePost("0")

	post := testPostFactory()

	r := updatePost("0", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusNotFound, restApi.NoSuchPost)
}

func TestUpdatePostWithInvalidID(t *testing.T) {
	post := testPostFactory()

	r := updatePost("post1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidID)
}

func TestDeletePostNonexistentPost(t *testing.T) {
	postToDeleteId := "1"

	deletePost(postToDeleteId)

	r := deletePost(postToDeleteId)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkNiceResponse(r, http.StatusOK)
}

func TestDeletePostWithInvalidID(t *testing.T) {
	r := deletePost("post1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidID)
}

func TestGetRangeOfPostsWithCustomPostsPerPage(t *testing.T) {
	var workingPosts []models.Post

	testPostsNumber := 20

	for i := 0; i < testPostsNumber; i++ {
		currentPost := testPostFactory()

		var response ResponsePost
		r := sendPostHandleMessage("POST", "http://"+address+"/api/posts", currentPost)
		decodePostResponse(r.Body, &response)

		workingPosts = append([]models.Post{response.Body}, workingPosts...)
	}

	r := getPosts("0", "20")

	var response ResponseRangePosts
	decodeRangePostsResponse(r.Body, &response)

	receivedPosts := response.Body

	if !comparePostLists(receivedPosts, workingPosts) {
		log.Fatalf("Received post list does not match proper post list\nReceived post list: %v\n Proper post list: %v",
			receivedPosts, workingPosts)
	}
}

func TestGetRangeOfPostsWithDefaultPostsPerPage(t *testing.T) {
	var workingPosts []models.Post

	testPostsNumber := 10

	for i := 0; i < testPostsNumber; i++ {
		currentPost := testPostFactory()

		var response ResponsePost
		r := sendPostHandleMessage("POST", "http://"+address+"/api/posts", currentPost)
		decodePostResponse(r.Body, &response)

		workingPosts = append([]models.Post{response.Body}, workingPosts...)
	}

	r := getPosts("0", "")

	var response ResponseRangePosts
	decodeRangePostsResponse(r.Body, &response)

	receivedPosts := response.Body

	if !comparePostLists(receivedPosts, workingPosts) {
		log.Fatalf("Received post list does not match proper post list\nReceived post list: %v\n Proper post list: %v",
			receivedPosts, workingPosts)
	}
}

func TestGetRangeOfPostsWithNegativePage(t *testing.T) {
	r := getPosts("-1", "")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostsRange)
}

func TestGetRangeOfPostsWithNonNumberPage(t *testing.T) {
	r := getPosts("adasdf", "")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostsRange)
}

func TestGetRangeOfPostsWithTooLongPostsPerPage(t *testing.T) {
	r := getPosts("0", strconv.Itoa(restApi.MaxPostsPerPage*2))
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostsRange)
}

func TestGetRangeOfPostsWithNonNumberPostsPerPage(t *testing.T) {
	r := getPosts("0", "asddfa")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, restApi.InvalidPostsRange)
}

func TestGetRangeOfPostsGetEmptyPage(t *testing.T) {
	r := getPosts("10000000", "")

	var response ResponseRangePosts
	decodeRangePostsResponse(r.Body, &response)

	receivedPosts := response.Body

	assert.Assert(t, len(receivedPosts) == 0)
}
