package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrAnnouncementNotFound = errors.New("announcement not found")

// Announcement is an admin broadcast shown to all authenticated users.
type Announcement struct {
	ID        int64
	Title     string
	Content   string
	Type      string // info, warning, critical
	IsActive  bool
	StartsAt  *time.Time
	ExpiresAt *time.Time
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *TrafficRepository) ListAnnouncements(ctx context.Context) ([]Announcement, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("traffic repository not initialized")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, content, type, is_active, starts_at, expires_at, created_by, created_at, updated_at
		FROM announcements
		ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	defer rows.Close()

	var list []Announcement
	for rows.Next() {
		item, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *TrafficRepository) ListActiveAnnouncements(ctx context.Context) ([]Announcement, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("traffic repository not initialized")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, content, type, is_active, starts_at, expires_at, created_by, created_at, updated_at
		FROM announcements
		WHERE is_active = 1
		  AND (starts_at IS NULL OR datetime(starts_at) <= datetime(?))
		  AND (expires_at IS NULL OR datetime(expires_at) > datetime(?))
		ORDER BY id DESC`, now, now)
	if err != nil {
		return nil, fmt.Errorf("list active announcements: %w", err)
	}
	defer rows.Close()

	var list []Announcement
	for rows.Next() {
		item, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *TrafficRepository) GetAnnouncement(ctx context.Context, id int64) (Announcement, error) {
	var item Announcement
	if r == nil || r.db == nil {
		return item, errors.New("traffic repository not initialized")
	}
	if id <= 0 {
		return item, errors.New("announcement id is required")
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, content, type, is_active, starts_at, expires_at, created_by, created_at, updated_at
		FROM announcements WHERE id = ?`, id)
	item, err := scanAnnouncement(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, ErrAnnouncementNotFound
		}
		return item, err
	}
	return item, nil
}

func (r *TrafficRepository) CreateAnnouncement(ctx context.Context, a Announcement) (Announcement, error) {
	if r == nil || r.db == nil {
		return Announcement{}, errors.New("traffic repository not initialized")
	}

	a.Title = strings.TrimSpace(a.Title)
	a.Content = strings.TrimSpace(a.Content)
	a.Type = normalizeAnnouncementType(a.Type)
	a.CreatedBy = strings.TrimSpace(a.CreatedBy)

	if a.Title == "" {
		return Announcement{}, errors.New("announcement title is required")
	}
	if a.Content == "" {
		return Announcement{}, errors.New("announcement content is required")
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO announcements (title, content, type, is_active, starts_at, expires_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.Title, a.Content, a.Type, boolToInt(a.IsActive), nullableTime(a.StartsAt), nullableTime(a.ExpiresAt), a.CreatedBy)
	if err != nil {
		return Announcement{}, fmt.Errorf("create announcement: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Announcement{}, fmt.Errorf("fetch announcement id: %w", err)
	}
	return r.GetAnnouncement(ctx, id)
}

func (r *TrafficRepository) UpdateAnnouncement(ctx context.Context, a Announcement) (Announcement, error) {
	if r == nil || r.db == nil {
		return Announcement{}, errors.New("traffic repository not initialized")
	}
	if a.ID <= 0 {
		return Announcement{}, errors.New("announcement id is required")
	}

	a.Title = strings.TrimSpace(a.Title)
	a.Content = strings.TrimSpace(a.Content)
	a.Type = normalizeAnnouncementType(a.Type)
	if a.Title == "" {
		return Announcement{}, errors.New("announcement title is required")
	}
	if a.Content == "" {
		return Announcement{}, errors.New("announcement content is required")
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE announcements
		SET title = ?, content = ?, type = ?, is_active = ?, starts_at = ?, expires_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		a.Title, a.Content, a.Type, boolToInt(a.IsActive), nullableTime(a.StartsAt), nullableTime(a.ExpiresAt), a.ID)
	if err != nil {
		return Announcement{}, fmt.Errorf("update announcement: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Announcement{}, ErrAnnouncementNotFound
	}
	return r.GetAnnouncement(ctx, a.ID)
}

func (r *TrafficRepository) DeleteAnnouncement(ctx context.Context, id int64) error {
	if r == nil || r.db == nil {
		return errors.New("traffic repository not initialized")
	}
	if id <= 0 {
		return errors.New("announcement id is required")
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM announcements WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrAnnouncementNotFound
	}
	return nil
}

func normalizeAnnouncementType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "warning", "critical":
		return strings.ToLower(strings.TrimSpace(t))
	default:
		return "info"
	}
}

func scanAnnouncement(scanner interface {
	Scan(dest ...any) error
}) (Announcement, error) {
	var item Announcement
	var isActive int
	var startsAt, expiresAt sql.NullString

	if err := scanner.Scan(
		&item.ID, &item.Title, &item.Content, &item.Type, &isActive,
		&startsAt, &expiresAt, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return item, fmt.Errorf("scan announcement: %w", err)
	}
	item.IsActive = isActive != 0
	item.StartsAt = parseNullableTime(startsAt.String)
	item.ExpiresAt = parseNullableTime(expiresAt.String)
	return item, nil
}

func parseNullableTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "0001-01-01") {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		if t2, err2 := time.Parse("2006-01-02 15:04:05", s); err2 == nil {
			return &t2
		}
		return nil
	}
	return &t
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
