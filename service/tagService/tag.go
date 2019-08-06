package tagService

import (
	"database/sql"
	pg "github.com/lib/pq"
)

// TODO: написать тесты на все методы

// GetAll - returns all tags
func GetAll(db *sql.DB) ([]string, error) {
	var tags []string

	rows, err := db.Query("select tag from tags")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var tag string
		if err = rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetAllByPostID - returns all tags of the given post
func GetAllByPostID(db *sql.DB, postID string) ([]string, error) {
	var tags []string

	rows, err := db.Query("select tag from tags inner join post_tags ON tags.tag_id=post_tags.tag_id where post_id = $1",
		postID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var tag string
		if err = rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// getAllByTagIDs - internal helper function that returns map where key is a tag ID and value is a tag name
// we need this function to get tag names by tag IDs
func getAllByTagIDs(db *sql.DB, tagIDs []string) (map[string]string, error) {
	tags := make(map[string]string)

	rows, err := db.Query("select tag_id, tag from tags where tag_id = any($1)", pg.Array(tagIDs))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var tagID string
		var tag string
		if err = rows.Scan(&tagID, &tag); err != nil {
			return nil, err
		}
		tags[tagID] = tag

	}

	return tags, nil
}

// GetAllInRangeOfPosts - retrieves all tags of the range of posts
// returns a map where key is a post ID and value is a slice of tag names related to this post
// you probably want to call GetAllPostIDsByTag function first, and then call this function
func GetAllInRangeOfPosts(db *sql.DB, postIDs []string) (map[string][]string, error) {
	rows, err := db.Query("select post_id, tag_id from post_tags where post_id = any($1) order by post_id, tag_id", pg.Array(postIDs))
	if err != nil {
		return nil, err
	}

	postTags := make(map[string][]string)
	tagIDs := make(map[string]bool)
	for rows.Next() {
		var postID string
		var tagID string
		if err = rows.Scan(&postID, &tagID); err != nil {
			return nil, err
		}
		postTags[postID] = append(postTags[postID], tagID)
		tagIDs[tagID] = true
	}

	var tagIDsAsSlice []string
	for tag := range tagIDs {
		tagIDsAsSlice = append(tagIDsAsSlice, tag)
	}
	tagNames, err := getAllByTagIDs(db, tagIDsAsSlice)
	if err != nil {
		return nil, err
	}

	// replace tag IDs with tag names
	for postID := range postTags {
		for tagIndex, tagID := range postTags[postID] {
			postTags[postID][tagIndex] = tagNames[tagID]
		}
	}

	return postTags, nil
}

// GetAllPostIDsByTag - returns all post IDs that have the given tag name
// returned slice is sorted in the ascending order
func GetAllPostIDsByTag(db *sql.DB, tag string) ([]string, error) {
	var ids []string

	rows, err := db.Query("select post_id from post_tags where tag_id = (select tag_id from tags where tag = $1)", tag)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}
