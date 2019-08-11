package tagService

import (
	"database/sql"
	"fmt"
	"github.com/blinky-z/Blog/models"
	pg "github.com/lib/pq"
	"strings"
)

// TODO: написать тесты на все методы

const (
	postTagsInsertFields = "post_id, tag_id"
	tagsInsertFields     = "tag"
	tagsAllFields        = "tag_id, tag"
)

// GetAll - returns all tags sorted by ID in descending order
func GetAll(db *sql.DB) ([]models.Tag, error) {
	var tags []models.Tag

	rows, err := db.Query("select " + tagsAllFields + " from tags order by tag_id desc")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var tag models.Tag
		if err = rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// updatePostTags - updates post tags
// after executing of this function post will have the same tags as passed to this method
func updatePostTags(tx *sql.Tx, postID string, tags []string) error {
	_, err := tx.Exec("delete from post_tags where post_id = $1", postID)
	if err != nil {
		return err
	}

	if len(tags) != 0 {
		tagIDs, err := getTagIDsByNames(tx, tags)
		if err != nil {
			return err
		}

		query := "insert into post_tags (" + postTagsInsertFields + ") values "
		args := make([]interface{}, 0)

		iter := 1
		for _, tag := range tags {
			query += fmt.Sprintf("($%d, $%d),", iter, iter+1)
			args = append(args, postID, tagIDs[tag])
			iter = iter + 2
		}
		query = strings.TrimSuffix(query, ",")

		if stmt, err := tx.Prepare(query); err != nil {
			return err
		} else {
			_, err := stmt.Exec(args...)
			return err
		}
	}
	return nil
}

// DeletePostTags - removes all post tags. Usually you want to call this function when deleting a post
// Notice that tags itself will not be deleted
func DeletePostTags(tx *sql.Tx, postID string) error {
	_, err := tx.Exec("delete from post_tags where post_id = $1", postID)
	return err
}

func Save(db *sql.DB, tag string) (models.Tag, error) {
	savedTag := models.Tag{}
	row := db.QueryRow("insert into tags ("+tagsInsertFields+") values ($1) returning "+tagsAllFields, tag)
	err := row.Scan(&savedTag.ID, &savedTag.Name)
	return savedTag, err
}

func Update(db *sql.DB, tagID, tag string) (models.Tag, error) {
	updatedTag := models.Tag{}
	row := db.QueryRow("update tags set tag = $1 where tag_id = $2 returning "+tagsAllFields, tag, tagID)
	err := row.Scan(&updatedTag.ID, &updatedTag.Name)
	return updatedTag, err
}

// Delete - deletes a tag by its ID and removes this tag from all tagged posts
func DeleteByID(db *sql.DB, tagID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("delete from tags where tag_id = $1", tagID)
	if err != nil {
		return err
	}
	_, err = tx.Exec("delete from post_tags where tag_id = $1", tagID)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// saveNewTags - saves bunch of tags. If tag already exists, it is omitted
func saveNewTags(tx *sql.Tx, tags []string) error {
	if len(tags) != 0 {
		query := "insert into tags (" + tagsInsertFields + ") values "
		args := make([]interface{}, len(tags))

		for index, tag := range tags {
			query += fmt.Sprintf("($%d),", index+1)
			args[index] = tag
		}
		query = strings.TrimSuffix(query, ",")

		query += " on conflict do nothing"
		if stmt, err := tx.Prepare(query); err != nil {
			return err
		} else {
			_, err := stmt.Exec(args...)
			return err
		}
	}
	return nil
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

// getTagNamesByIDs - internal helper function that returns map where key is a tag ID and value is a tag name
// we need this function to get tag name by tag ID
func getTagNamesByIDs(db *sql.DB, tagIDs []string) (map[string]string, error) {
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

// getTagNamesByIDs - internal helper function that returns map where key is a tag name and value is a tag ID
// we need this function to get tag ID by tag name
func getTagIDsByNames(tx *sql.Tx, tags []string) (map[string]string, error) {
	tagIDs := make(map[string]string)

	rows, err := tx.Query("select tag_id, tag from tags where tag = any($1)", pg.Array(tags))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var tagID string
		var tag string
		if err = rows.Scan(&tagID, &tag); err != nil {
			return nil, err
		}
		tagIDs[tag] = tagID
	}

	return tagIDs, nil
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
	tagNames, err := getTagNamesByIDs(db, tagIDsAsSlice)
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

// SavePostTags - saves new tags and updates post tags
// after executing of this function post will have the same tags as passed to this method
// supports current transaction as we need to save tags together with post
func SavePostTags(tx *sql.Tx, postID string, postTags []string) error {
	err := saveNewTags(tx, postTags)
	if err != nil {
		return err
	}

	return updatePostTags(tx, postID, postTags)
}
