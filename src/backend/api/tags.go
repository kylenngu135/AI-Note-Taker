package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *Handler) GetTagsHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tags, err := getAllTagsForUser(h.DB, userID)
	if err != nil {
		http.Error(w, "failed to get tags", http.StatusInternalServerError)
		return
	}

	if tags == nil {
		tags = []Tag{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func (h *Handler) CreateTagHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "tag name required", http.StatusBadRequest)
		return
	}

	color := req.Color
	if color == "" {
		color = "#6b7280"
	}

	tagID := uuid.New().String()
	tag, err := createOrGetTag(h.DB, tagID, userID, req.Name, "user", color)
	if err != nil {
		http.Error(w, "failed to create tag", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

func (h *Handler) DeleteTagHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tagID := r.PathValue("id")

	if err := deleteTagForUser(h.DB, tagID, userID); err != nil {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AddTagToUploadHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.PathValue("id")

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "tag name required", http.StatusBadRequest)
		return
	}

	color := req.Color
	if color == "" {
		color = "#6b7280"
	}

	note, err := getNoteByUploadID(h.DB, uploadID)
	if err != nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	tagID := uuid.New().String()
	tag, err := createOrGetTag(h.DB, tagID, userID, req.Name, "user", color)
	if err != nil {
		http.Error(w, "failed to create tag", http.StatusInternalServerError)
		return
	}

	if err := addTagToNote(h.DB, note.ID, tag.ID); err != nil {
		http.Error(w, "failed to add tag", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

func (h *Handler) RemoveTagFromUploadHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.PathValue("id")
	tagID := r.PathValue("tagId")

	note, err := getNoteByUploadID(h.DB, uploadID)
	if err != nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	// Verify tag belongs to user
	tag, err := getTagByID(h.DB, tagID)
	if err != nil || tag.UserID != userID {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	if err := removeTagFromNote(h.DB, note.ID, tagID); err != nil {
		http.Error(w, "failed to remove tag", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
