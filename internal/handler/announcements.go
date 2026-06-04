package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"miaomiaowu/internal/auth"
	"miaomiaowu/internal/storage"
)

type announcementPayload struct {
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Type      string  `json:"type"`
	IsActive  bool    `json:"is_active"`
	StartsAt  *string `json:"starts_at"`
	ExpiresAt *string `json:"expires_at"`
}

type announcementResponse struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Type      string  `json:"type"`
	IsActive  bool    `json:"is_active"`
	StartsAt  *string `json:"starts_at"`
	ExpiresAt *string `json:"expires_at"`
	CreatedBy string  `json:"created_by"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func NewAnnouncementsAdminHandler(repo *storage.TrafficRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/admin/announcements")
		path = strings.Trim(path, "/")

		switch {
		case path == "" && r.Method == http.MethodGet:
			handleListAnnouncements(w, r, repo, false)
		case path == "" && r.Method == http.MethodPost:
			handleCreateAnnouncement(w, r, repo)
		case path != "" && r.Method == http.MethodPut:
			handleUpdateAnnouncement(w, r, repo, path)
		case path != "" && r.Method == http.MethodDelete:
			handleDeleteAnnouncement(w, r, repo, path)
		default:
			methodNotAllowed(w, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
		}
	})
}

func NewAnnouncementsUserHandler(repo *storage.TrafficRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		handleListAnnouncements(w, r, repo, true)
	})
}

func handleListAnnouncements(w http.ResponseWriter, r *http.Request, repo *storage.TrafficRepository, activeOnly bool) {
	var (
		list []storage.Announcement
		err  error
	)
	if activeOnly {
		list, err = repo.ListActiveAnnouncements(r.Context())
	} else {
		list, err = repo.ListAnnouncements(r.Context())
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	resp := make([]announcementResponse, 0, len(list))
	for _, item := range list {
		resp = append(resp, toAnnouncementResponse(item))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"announcements": resp})
}

func handleCreateAnnouncement(w http.ResponseWriter, r *http.Request, repo *storage.TrafficRepository) {
	var payload announcementPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	username := auth.UsernameFromContext(r.Context())
	item, err := repo.CreateAnnouncement(r.Context(), payloadToAnnouncement(payload, 0, username))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"announcement": toAnnouncementResponse(item)})
}

func handleUpdateAnnouncement(w http.ResponseWriter, r *http.Request, repo *storage.TrafficRepository, idSegment string) {
	id, err := strconv.ParseInt(idSegment, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid announcement id"))
		return
	}

	var payload announcementPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	item, err := repo.UpdateAnnouncement(r.Context(), payloadToAnnouncement(payload, id, ""))
	if err != nil {
		if errors.Is(err, storage.ErrAnnouncementNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"announcement": toAnnouncementResponse(item)})
}

func handleDeleteAnnouncement(w http.ResponseWriter, r *http.Request, repo *storage.TrafficRepository, idSegment string) {
	id, err := strconv.ParseInt(idSegment, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid announcement id"))
		return
	}

	if err := repo.DeleteAnnouncement(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrAnnouncementNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func payloadToAnnouncement(p announcementPayload, id int64, createdBy string) storage.Announcement {
	return storage.Announcement{
		ID:        id,
		Title:     p.Title,
		Content:   p.Content,
		Type:      p.Type,
		IsActive:  p.IsActive,
		StartsAt:  parseTimePtr(p.StartsAt),
		ExpiresAt: parseTimePtr(p.ExpiresAt),
		CreatedBy: createdBy,
	}
}

func parseTimePtr(s *string) *time.Time {
	if s == nil {
		return nil
	}
	raw := strings.TrimSpace(*s)
	if raw == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return &t
		}
	}
	return nil
}

func toAnnouncementResponse(a storage.Announcement) announcementResponse {
	return announcementResponse{
		ID:        a.ID,
		Title:     a.Title,
		Content:   a.Content,
		Type:      a.Type,
		IsActive:  a.IsActive,
		StartsAt:  formatTimePtr(a.StartsAt),
		ExpiresAt: formatTimePtr(a.ExpiresAt),
		CreatedBy: a.CreatedBy,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
