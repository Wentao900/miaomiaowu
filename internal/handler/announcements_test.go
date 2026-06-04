package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"miaomiaowu/internal/auth"
	"miaomiaowu/internal/storage"
)

func TestAnnouncementsAdminAndUserHandlers(t *testing.T) {
	dir := t.TempDir()
	repo, err := storage.NewTrafficRepository(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	adminHandler := NewAnnouncementsAdminHandler(repo)
	userHandler := NewAnnouncementsUserHandler(repo)

	// Create
	body, _ := json.Marshal(map[string]any{
		"title":     "系统公告",
		"content":   "欢迎使用妙妙屋",
		"type":      "info",
		"is_active": true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/announcements", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithUsername(req.Context(), "admin"))
	rec := httptest.NewRecorder()
	adminHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}

	var createResp struct {
		Announcement struct {
			ID int64 `json:"id"`
		} `json:"announcement"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatal(err)
	}
	if createResp.Announcement.ID <= 0 {
		t.Fatal("invalid announcement id")
	}

	// User list
	req = httptest.NewRequest(http.MethodGet, "/api/user/announcements", nil)
	req = req.WithContext(auth.ContextWithUsername(req.Context(), "user1"))
	rec = httptest.NewRecorder()
	userHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("user list status = %d body=%s", rec.Code, rec.Body.String())
	}
	var userResp struct {
		Announcements []struct {
			Title string `json:"title"`
		} `json:"announcements"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &userResp); err != nil {
		t.Fatal(err)
	}
	if len(userResp.Announcements) != 1 || userResp.Announcements[0].Title != "系统公告" {
		t.Fatalf("user announcements = %+v", userResp.Announcements)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/announcements/"+strconv.FormatInt(createResp.Announcement.ID, 10), nil)
	req = req.WithContext(auth.ContextWithUsername(req.Context(), "admin"))
	rec = httptest.NewRecorder()
	adminHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", rec.Code, rec.Body.String())
	}
}
