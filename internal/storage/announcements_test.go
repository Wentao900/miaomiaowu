package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAnnouncementsCRUDAndActiveFilter(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	repo, err := NewTrafficRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	ctx := context.Background()

	created, err := repo.CreateAnnouncement(ctx, Announcement{
		Title:     "维护通知",
		Content:   "今晚 22:00 维护",
		Type:      "warning",
		IsActive:  true,
		CreatedBy: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID <= 0 {
		t.Fatal("expected positive id")
	}

	future := time.Now().Add(24 * time.Hour).UTC()
	past := time.Now().Add(-time.Hour).UTC()
	expired, err := repo.CreateAnnouncement(ctx, Announcement{
		Title:     "已过期",
		Content:   "不应展示",
		Type:      "info",
		IsActive:  true,
		ExpiresAt: &past,
		CreatedBy: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = expired

	scheduled, err := repo.CreateAnnouncement(ctx, Announcement{
		Title:     "未开始",
		Content:   "不应展示",
		Type:      "info",
		IsActive:  true,
		StartsAt:  &future,
		CreatedBy: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = scheduled

	active, err := repo.ListActiveAnnouncements(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].Title != "维护通知" {
		t.Fatalf("active announcements = %+v, want only 维护通知", active)
	}

	created.Content = "更新内容"
	updated, err := repo.UpdateAnnouncement(ctx, created)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Content != "更新内容" {
		t.Fatalf("content = %q", updated.Content)
	}

	if err := repo.DeleteAnnouncement(ctx, created.ID); err != nil {
		t.Fatal(err)
	}

	list, err := repo.ListAnnouncements(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range list {
		if item.ID == created.ID {
			t.Fatal("deleted announcement still listed")
		}
	}

	_, err = os.Stat(dbPath)
	if err != nil {
		t.Fatalf("db file missing: %v", err)
	}
}
