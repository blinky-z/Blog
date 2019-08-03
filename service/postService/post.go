package postService

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
)

const (
	// postsInsertFields - fields that should be filled while inserting a new entity
	postsInsertFields = "title, author, snippet, content, metadata"
	// postsAllFields - all entity fields
	postsAllFields = "id, title, author, date, snippet, content, metadata"
)

// Save - saves a new post in database
// returns a created post pointed to by 'createdPost' and error
func Save(db *sql.DB, request *SaveRequest) (createdPost *models.Post, err error) {
	createdPost = &models.Post{}
	err = nil

	encodedMetadata, err := json.Marshal(request.Metadata)
	if err != nil {
		return
	}

	var metadataAsJSONString string
	if err = db.QueryRow("insert into posts ("+postsInsertFields+") values($1, $2, $3, $4, $5) "+
		"RETURNING "+postsAllFields, request.Title, request.Author, request.Snippet, request.Content, encodedMetadata).
		Scan(&createdPost.ID, &createdPost.Title, &createdPost.Author, &createdPost.Date, &createdPost.Snippet,
			&createdPost.Content, &metadataAsJSONString); err != nil {
		return
	}

	err = json.Unmarshal([]byte(metadataAsJSONString), &createdPost.Metadata)
	if err != nil {
		return
	}

	return
}

// Update - updates post in database
// returns an updated post pointed to by 'updatedPost' and error
func Update(db *sql.DB, request *UpdateRequest) (updatedPost *models.Post, err error) {
	updatedPost = &models.Post{}
	err = nil

	encodedMetadata, err := json.Marshal(request.Metadata)
	if err != nil {
		return
	}

	var metadataAsJSONString string
	if err = db.QueryRow("UPDATE posts SET ("+postsInsertFields+") = ($1, $2, $3, $4, $5) "+
		"WHERE id = $6 RETURNING "+postsAllFields, request.Title, request.Author, request.Snippet, request.Content,
		encodedMetadata, request.ID).
		Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Author, &updatedPost.Date, &updatedPost.Snippet, &updatedPost.Content,
			&metadataAsJSONString); err != nil {
		return
	}
	err = json.Unmarshal([]byte(metadataAsJSONString), &updatedPost.Metadata)
	if err != nil {
		return
	}

	return
}

// ExistsById - checks if post with the given ID exists in database
// returns boolean indicating whether post exists or not and error
func ExistsById(db *sql.DB, postId string) (bool, error) {
	var postExists bool
	err := db.QueryRow("select exists(select 1 from posts where id = $1)", postId).Scan(&postExists)
	return postExists, err
}

// Delete - deletes post from database
func Delete(db *sql.DB, postId string) error {
	if _, err := db.Exec("DELETE FROM posts WHERE id = $1", postId); err != nil {
		return err
	}
	return nil
}

// GetById - retrieves post with the given ID from database
func GetById(db *sql.DB, id string) (models.Post, error) {
	var post models.Post

	var metadataAsJSONString string
	if err := db.QueryRow("select "+postsAllFields+" from posts where id = $1", id).
		Scan(&post.ID, &post.Title, &post.Author, &post.Date, &post.Snippet, &post.Content, &metadataAsJSONString); err != nil {
		if err == sql.ErrNoRows {
			return post, err
		}

		return post, err
	}

	err := json.Unmarshal([]byte(metadataAsJSONString), &post.Metadata)
	if err != nil {
		return post, err
	}

	return post, nil
}

// GetPostsInRange - retrieves all posts in the given range
// Range is described by page and entities per page args
// returns slice which len is equal to `postsPerPage` and error
func GetPostsInRange(db *sql.DB, page, postsPerPage int) ([]models.Post, error) {
	var posts []models.Post

	rows, err := db.Query("select "+postsAllFields+" from posts order by id DESC offset $1 limit $2",
		page*postsPerPage, postsPerPage)
	if err != nil {
		return posts, err
	}

	for rows.Next() {
		var currentPost models.Post
		var metadataAsJSONString string
		if err = rows.Scan(&currentPost.ID, &currentPost.Title, &currentPost.Author, &currentPost.Date,
			&currentPost.Snippet, &currentPost.Content, &metadataAsJSONString); err != nil {
			return posts, err
		}

		if err = json.Unmarshal([]byte(metadataAsJSONString), &currentPost.Metadata); err != nil {
			return posts, err
		}

		posts = append(posts, currentPost)
	}

	return posts, nil
}
