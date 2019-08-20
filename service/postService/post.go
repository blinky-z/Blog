package postService

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/tagService"
	pg "github.com/lib/pq"
)

const (
	// postsInsertFields - fields that should be filled while inserting a new entity
	postsInsertFields = "title, snippet, content, content_md, metadata"
	// postsAllFieldsWithHtmlContent - all entity fields with content as html
	postsAllFieldsWithHtmlContent = "id, title, date, snippet, content, metadata"
	// postsAllFieldsWithMarkdownContent - all entity fields with content as markdown
	postsAllFieldsWithMarkdownContent = "id, title, date, snippet, content_md, metadata"
)

// Save - saves a new post
// returns a created post pointed to by 'createdPost' and error
func Save(db *sql.DB, request *SaveRequest) (*models.Post, error) {
	createdPost := &models.Post{}

	encodedMetadata, err := json.Marshal(request.Metadata)
	if err != nil {
		return createdPost, err
	}

	tx, err := db.Begin()
	if err != nil {
		return createdPost, err
	}

	var metadataAsJSONString string
	if err = tx.QueryRow("insert into posts ("+postsInsertFields+") values($1, $2, $3, $4, $5) "+
		"RETURNING "+postsAllFieldsWithHtmlContent,
		request.Title, request.Snippet, request.Content, request.ContentMD, encodedMetadata).
		Scan(&createdPost.ID, &createdPost.Title, &createdPost.Date, &createdPost.Snippet, &createdPost.Content,
			&metadataAsJSONString); err != nil {
		return createdPost, err
	}

	if err = json.Unmarshal([]byte(metadataAsJSONString), &createdPost.Metadata); err != nil {
		tx.Rollback()
		return createdPost, err
	}

	err = tagService.SavePostTags(tx, createdPost.ID, request.Tags)
	if err != nil {
		tx.Rollback()
		return createdPost, err
	}

	createdPost.Tags = request.Tags
	return createdPost, tx.Commit()
}

// Update - updates post
// returns an updated post pointed to by 'updatedPost' and error
func Update(db *sql.DB, request *UpdateRequest) (*models.Post, error) {
	updatedPost := &models.Post{}

	encodedMetadata, err := json.Marshal(request.Metadata)
	if err != nil {
		return updatedPost, err
	}

	tx, err := db.Begin()
	if err != nil {
		return updatedPost, err
	}

	var metadataAsJSONString string
	if err = tx.QueryRow("UPDATE posts SET ("+postsInsertFields+") = ($1, $2, $3, $4, $5) "+
		"WHERE id = $6 RETURNING "+postsAllFieldsWithHtmlContent,
		request.Title, request.Snippet, request.Content, request.ContentMD, encodedMetadata, request.ID).
		Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Date, &updatedPost.Snippet, &updatedPost.Content,
			&metadataAsJSONString); err != nil {
		return updatedPost, err
	}

	if err = json.Unmarshal([]byte(metadataAsJSONString), &updatedPost.Metadata); err != nil {
		tx.Rollback()
		return updatedPost, err
	}

	err = tagService.SavePostTags(tx, updatedPost.ID, request.Tags)
	if err != nil {
		tx.Rollback()
		return updatedPost, err
	}

	updatedPost.Tags = request.Tags
	return updatedPost, tx.Commit()
}

// DeleteByID - deletes post from database
func DeleteByID(db *sql.DB, postID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM posts WHERE id = $1", postID); err != nil {
		return err
	}

	if err = tagService.DeletePostTags(tx, postID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetByID - retrieves post with the given ID
// if post does not exist, err.SqlNoRows error will be returned
func GetByID(db *sql.DB, postID string) (models.Post, error) {
	var post models.Post

	var metadataAsJSONString string
	if err := db.QueryRow("select "+postsAllFieldsWithHtmlContent+" from posts where id = $1", postID).
		Scan(&post.ID, &post.Title, &post.Date, &post.Snippet, &post.Content, &metadataAsJSONString); err != nil {
		return post, err
	}

	err := json.Unmarshal([]byte(metadataAsJSONString), &post.Metadata)
	if err != nil {
		return post, err
	}

	if tags, err := tagService.GetAllByPostID(db, postID); err != nil {
		return post, err
	} else {
		post.Tags = tags
	}

	return post, nil
}

// GetByIDWithMarkdownContent - retrieves post with the given ID, the content as markdown
// if post does not exist, err.SqlNoRows error will be returned
func GetByIDWithMarkdownContent(db *sql.DB, postID string) (models.Post, error) {
	var post models.Post

	var metadataAsJSONString string
	if err := db.QueryRow("select "+postsAllFieldsWithMarkdownContent+" from posts where id = $1", postID).
		Scan(&post.ID, &post.Title, &post.Date, &post.Snippet, &post.Content, &metadataAsJSONString); err != nil {
		return post, err
	}

	err := json.Unmarshal([]byte(metadataAsJSONString), &post.Metadata)
	if err != nil {
		return post, err
	}

	if tags, err := tagService.GetAllByPostID(db, postID); err != nil {
		return post, err
	} else {
		post.Tags = tags
	}

	return post, nil
}

func fillTags(db *sql.DB, posts []models.Post) ([]models.Post, error) {
	postIDs := make([]string, 0)
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}
	postTags, err := tagService.GetAllInRangeOfPosts(db, postIDs)
	if err != nil {
		return posts, err
	}

	for postIndex, post := range posts {
		posts[postIndex].Tags = postTags[post.ID]
	}

	return posts, nil
}

// TODO: тесты
// GetPostsInRangeByTag - retrieves all posts in the given range with the given tag
// the returned slice is sorted by post creation time in descending order
func GetPostsInRangeByTag(db *sql.DB, offset, postsPerPage int, tag string) ([]models.Post, error) {
	var posts []models.Post

	postIds, err := tagService.GetAllPostIDsByTag(db, tag)
	if err != nil {
		return posts, err
	}

	rows, err := db.Query("select "+postsAllFieldsWithHtmlContent+" from posts where id = any($1) order by date DESC offset $2 limit $3",
		pg.Array(postIds), offset, postsPerPage)
	if err != nil {
		return posts, err
	}

	for rows.Next() {
		var currentPost models.Post
		var metadataAsJSONString string
		if err = rows.Scan(&currentPost.ID, &currentPost.Title, &currentPost.Date,
			&currentPost.Snippet, &currentPost.Content, &metadataAsJSONString); err != nil {
			return posts, err
		}

		if err = json.Unmarshal([]byte(metadataAsJSONString), &currentPost.Metadata); err != nil {
			return posts, err
		}

		posts = append(posts, currentPost)
	}

	return fillTags(db, posts)
}

// TODO: тесты
// GetPostsInRange - retrieves all posts in the given range
// Range is described by page and entities per page args
// returns slice which len is equal to `postsPerPage` and error
// the returned slice is sorted by post creation time in descending order
func GetPostsInRange(db *sql.DB, offset, postsPerPage int) ([]models.Post, error) {
	var posts []models.Post

	rows, err := db.Query("select "+postsAllFieldsWithHtmlContent+" from posts order by date DESC offset $1 limit $2",
		offset, postsPerPage)
	if err != nil {
		return posts, err
	}

	for rows.Next() {
		var currentPost models.Post
		var metadataAsJSONString string
		if err = rows.Scan(&currentPost.ID, &currentPost.Title, &currentPost.Date,
			&currentPost.Snippet, &currentPost.Content, &metadataAsJSONString); err != nil {
			return posts, err
		}

		if err = json.Unmarshal([]byte(metadataAsJSONString), &currentPost.Metadata); err != nil {
			return posts, err
		}

		posts = append(posts, currentPost)
	}

	return fillTags(db, posts)
}
