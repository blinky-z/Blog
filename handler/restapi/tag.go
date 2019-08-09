package restapi

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/tagService"
	"github.com/gorilla/mux"
	pg "github.com/lib/pq"
	"log"
	"net/http"
	"strings"
)

// PostAPIHandler - used for dependency injection
type TagAPIHandler struct {
	db       *sql.DB
	logInfo  *log.Logger
	logError *log.Logger
}

func NewTagAPIHandler(db *sql.DB, logInfo, logError *log.Logger) *TagAPIHandler {
	return &TagAPIHandler{
		db:       db,
		logInfo:  logInfo,
		logError: logError,
	}
}

// error codes for this API
var (
	// TagAlreadyExists - tag already exists
	TagAlreadyExists = models.NewRequestErrorCode("TAG_ALREADY_EXISTS")
)

func (api *TagAPIHandler) CreateTagHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := models.CreateTagRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		tag := strings.TrimSpace(request.Name)
		if tag == "" {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		logInfo.Printf("Got new tag creation request. Tag name: %s", tag)

		createdTag, err := tagService.Save(api.db, tag)
		if err != nil {
			logError.Printf("Error saving a tag. Tag: %s. Error: %s", tag, err)
			if err.(*pg.Error).Code == "23505" {
				RespondWithError(w, http.StatusBadRequest, TagAlreadyExists)
				return
			}

			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Tag saved. Saved tag: %v", createdTag)
		RespondWithBody(w, http.StatusOK, createdTag)
	})
}

func (api *TagAPIHandler) UpdateTagHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tagID := mux.Vars(r)["id"]
		request := models.CreateTagRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		tag := strings.TrimSpace(request.Name)
		if tag == "" {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		logInfo.Printf("Got new tag update request. Tag ID: %s, New tag name: %s", tagID, tag)

		createdTag, err := tagService.Update(api.db, tagID, tag)
		if err != nil {
			logError.Printf("Error updating a tag. Tag ID: %s. Error: %s", tag, err)
			if err.(*pg.Error).Code == "23505" {
				RespondWithError(w, http.StatusBadRequest, TagAlreadyExists)
				return
			}

			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Tag updated. Updated tag: %v", createdTag)
		RespondWithBody(w, http.StatusOK, createdTag)
	})
}

func (api *TagAPIHandler) DeleteTagHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tagID := mux.Vars(r)["id"]
		logInfo.Printf("Got new tag deletion request. Tag ID: %s", tagID)

		if err := tagService.DeleteByID(api.db, tagID); err != nil {
			logError.Printf("Error deleting a tag. Post ID: %s. Error: %s", tagID, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Tag deleted. Tag ID: %s", tagID)
		Respond(w, http.StatusOK)
	})
}
