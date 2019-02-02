package postService

import (
	"database/sql"
	"github.com/blinky-z/Blog/models"
)

// GetCertainPost - return post as models.Post. Used for sending post back to user using http with following client-side
// rendering and also for server-side rendering of /posts/{id} page
func GetCertainPost(env *models.Env, id string) (models.Post, error) {
	env.LogInfo.Print("Got new Post GET job")

	var post models.Post

	env.LogInfo.Printf("Getting post with ID %s from database", id)

	if err := env.Db.QueryRow("select id, title, date, content from posts where id = $1", id).
		Scan(&post.ID, &post.Title, &post.Date, &post.Content); err != nil {
		if err == sql.ErrNoRows {
			env.LogInfo.Printf("Can not GET post with ID %s : post does not exist", id)
			return post, err
		}

		env.LogError.Print(err)
		return post, err
	}

	env.LogInfo.Printf("Post with ID %s succesfully arrived from database", id)

	return post, nil
}

// GetPosts - return posts list as models.Post array. Used for sending posts back to user using http with following client-side
// rendering and also for server-side rendering of index page
func GetPosts(env *models.Env, page, postsPerPage int) ([]models.Post, error) {
	env.LogInfo.Print("Got new Range of Posts GET job")

	var posts []models.Post

	env.LogInfo.Printf("Getting Range of Posts with following params: (page: %d, posts per page: %d) from database",
		page, postsPerPage)

	rows, err := env.Db.Query("select id, title, date, content from posts order by id DESC offset $1 limit $2",
		page*postsPerPage, postsPerPage)
	if err != nil {
		env.LogError.Print(err)
		return posts, err
	}

	for rows.Next() {
		var currentPost models.Post
		if err = rows.Scan(&currentPost.ID, &currentPost.Title, &currentPost.Date, &currentPost.Content); err != nil {
			env.LogError.Print(err)
			return posts, err
		}

		posts = append(posts, currentPost)
	}

	env.LogInfo.Print("Range of Posts successfully arrived from database")

	return posts, nil
}
