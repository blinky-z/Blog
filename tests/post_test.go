package tests

import (
	"github.com/blinky-z/Blog/handler/restapi"
	"github.com/blinky-z/Blog/models"
	"github.com/stretchr/testify/require"
	"net/http"
	"strconv"
	"testing"
)

func TestPostIntegrationTest(t *testing.T) {
	var workingPost models.Post

	// Step 1: Create post
	{
		request := createPostRequestFactory()

		r := createPost(request)
		resp := decodeResponseWithPostBody(r.Body)
		assertNiceResponse(r, http.StatusCreated)

		workingPost = resp.Body
	}

	// Step 2: Get created post and compare it with working
	{
		r := getCertainPost(workingPost.ID)
		resp := decodeResponseWithPostBody(r.Body)
		assertNiceResponse(r, http.StatusOK)

		receivedPost := resp.Body

		if !comparePosts(receivedPost, workingPost) {
			t.Fatalf("Received post does not match created post\nReceived post: %v\nCreated post: %v",
				receivedPost, workingPost)
		}
	}

	// Step 3: Update created post
	{
		request := updatePostRequestFactory()

		r := updatePost(workingPost.ID, request)
		resp := decodeResponseWithPostBody(r.Body)
		assertNiceResponse(r, http.StatusCreated)

		workingPost = resp.Body
	}

	// Step 4: Get updated post and compare it with working
	{
		r := getCertainPost(workingPost.ID)
		resp := decodeResponseWithCertainPostBody(r.Body)
		assertNiceResponse(r, http.StatusOK)

		receivedPost := resp.Body.Post

		if !comparePosts(receivedPost, workingPost) {
			t.Fatalf("Received post does not match created post\nReceived post: %v\nCreated post: %v",
				receivedPost, workingPost)
		}
	}

	// Step 5: Delete post
	{
		r := deletePost(workingPost.ID)
		assertNiceResponse(r, http.StatusOK)
	}

	// Step 6: Get deleted post
	{
		r := getCertainPost(workingPost.ID)
		assertErrorResponse(r, http.StatusNotFound, restapi.NoSuchPost)
	}
}

func TestCreatePostWithInvalidRequestBody(t *testing.T) {
	message := `{"bad request body"}`
	r := createPost(message)

	assertErrorResponse(r, http.StatusBadRequest, restapi.BadRequestBody)
}

func TestGetPostWithInvalidID(t *testing.T) {
	r := getCertainPost("post1")

	assertErrorResponse(r, http.StatusBadRequest, restapi.InvalidRequest)
}

func TestGetNotExistingPost(t *testing.T) {
	// create post
	createPostRequest := createPostRequestFactory()
	r := createPost(createPostRequest)
	post := decodeResponseWithPostBody(r.Body).Body

	// delete created post
	deletePost(post.ID)

	// get deleted post
	r = getCertainPost(post.ID)

	assertErrorResponse(r, http.StatusNotFound, restapi.NoSuchPost)
}

func TestUpdatePostSetInvalidRequestBody(t *testing.T) {
	// create post
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body

	message := `{"bad request body":"asd"}`
	r := updatePost(post.ID, message)

	assertErrorResponse(r, http.StatusBadRequest, restapi.BadRequestBody)
}

func TestUpdateNotExistingPost(t *testing.T) {
	// create post
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body

	// delete created post
	deletePost(post.ID)

	updatePostRequest := updatePostRequestFactory()
	r := updatePost(post.ID, updatePostRequest)

	assertErrorResponse(r, http.StatusBadRequest, restapi.NoSuchPost)
}

func TestUpdatePostWithInvalidID(t *testing.T) {
	updatePostRequest := updatePostRequestFactory()
	r := updatePost("post1", updatePostRequest)

	assertErrorResponse(r, http.StatusBadRequest, restapi.InvalidRequest)
}

func TestDeleteNotExistingPostShouldBeIdempotent(t *testing.T) {
	// create post
	createPostRequest := createPostRequestFactory()
	createPostResponse := createPost(createPostRequest)
	post := decodeResponseWithPostBody(createPostResponse.Body).Body

	// delete created post
	deletePost(post.ID)

	r := deletePost(post.ID)

	assertNiceResponse(r, http.StatusOK)
}

func TestDeletePostWithInvalidID(t *testing.T) {
	r := deletePost("post1")

	assertErrorResponse(r, http.StatusBadRequest, restapi.InvalidRequest)
}

func TestGetRangeOfPostsWithCustomPostsPerPage(t *testing.T) {
	var workingPosts []models.Post

	postsToCreate := 20
	postsToCreateAsString := strconv.Itoa(postsToCreate)

	for i := 0; i < postsToCreate; i++ {
		createPostRequest := createPostRequestFactory()
		r := createPost(createPostRequest)

		post := decodeResponseWithPostBody(r.Body).Body

		workingPosts = append(workingPosts, post)
	}

	r := getPostsInRange("0", postsToCreateAsString)
	receivedPosts := decodeResponseWithRangeOfPostsBody(r.Body).Body

	require.ElementsMatch(t, receivedPosts, workingPosts)
}

func TestGetRangeOfPostsWithDefaultPostsPerPage(t *testing.T) {
	var workingPosts []models.Post

	postsToCreateAsString := restapi.DefaultPostsPerPage
	postsToCreate, _ := strconv.Atoi(postsToCreateAsString)

	for i := 0; i < postsToCreate; i++ {
		createPostRequest := createPostRequestFactory()
		r := createPost(createPostRequest)

		post := decodeResponseWithPostBody(r.Body).Body

		workingPosts = append(workingPosts, post)
	}

	r := getPostsInRange("0", postsToCreateAsString)
	receivedPosts := decodeResponseWithRangeOfPostsBody(r.Body).Body

	require.ElementsMatch(t, receivedPosts, workingPosts)
}

func TestGetRangeOfPostsGetEmptyPage(t *testing.T) {
	r := getPostsInRange("10000000", "")
	receivedPosts := decodeResponseWithRangeOfPostsBody(r.Body).Body

	require.Equal(t, 0, len(receivedPosts))
}
