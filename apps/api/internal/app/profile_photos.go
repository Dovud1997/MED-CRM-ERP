package app

import (
	"encoding/base64"
	"net/http"
	"strings"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type profilePhotoRequest struct {
	ContentType string `json:"contentType"`
	Data        string `json:"data"`
}

func (a *App) saveProfilePhoto(entityType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			errorJSON(w, 400, "invalid_id", "Invalid profile id", requestID(r))
			return
		}
		var in profilePhotoRequest
		if !decode(w, r, &in) {
			return
		}
		in.ContentType = strings.ToLower(strings.TrimSpace(in.ContentType))
		allowed := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}
		data, err := base64.StdEncoding.DecodeString(in.Data)
		if err != nil || !allowed[in.ContentType] || len(data) == 0 || len(data) > 5*1024*1024 {
			errorJSON(w, 422, "invalid_photo", "Use JPG, PNG or WEBP up to 5 MB", requestID(r))
			return
		}
		detected := http.DetectContentType(data)
		if detected != in.ContentType && !(in.ContentType == "image/webp" && strings.Contains(detected, "octet-stream")) {
			errorJSON(w, 422, "invalid_photo", "Photo content does not match its type", requestID(r))
			return
		}
		p := auth.PrincipalFrom(r.Context())
		table := "patients"
		if entityType == "employee" {
			table = "employees"
		}
		var exists bool
		_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM `+table+` WHERE id=$1 AND organization_id=$2`+map[bool]string{true: " AND deleted_at IS NULL", false: ""}[entityType == "patient"]+`)`, id, p.OrganizationID).Scan(&exists)
		if !exists {
			errorJSON(w, 404, "not_found", "Profile not found", requestID(r))
			return
		}
		_, err = a.db.Exec(r.Context(), `INSERT INTO profile_photos(organization_id,entity_type,entity_id,content_type,byte_size,content,uploaded_by) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(organization_id,entity_type,entity_id) DO UPDATE SET content_type=excluded.content_type,byte_size=excluded.byte_size,content=excluded.content,uploaded_by=excluded.uploaded_by,updated_at=now()`, p.OrganizationID, entityType, id, in.ContentType, len(data), data, p.UserID)
		if err != nil {
			errorJSON(w, 500, "database_error", "Could not save photo", requestID(r))
			return
		}
		_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'PROFILE_PHOTO_UPDATED',$3,$4,'success',jsonb_build_object('bytes',$5::int))`, p.OrganizationID, p.UserID, entityType, id, len(data))
		writeJSON(w, 200, map[string]bool{"success": true})
	}
}

func (a *App) getProfilePhoto(entityType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			errorJSON(w, 400, "invalid_id", "Invalid profile id", requestID(r))
			return
		}
		p := auth.PrincipalFrom(r.Context())
		var contentType string
		var content []byte
		err := a.db.QueryRow(r.Context(), `SELECT content_type,content FROM profile_photos WHERE organization_id=$1 AND entity_type=$2 AND entity_id=$3`, p.OrganizationID, entityType, id).Scan(&contentType, &content)
		if err != nil {
			errorJSON(w, 404, "not_found", "Photo not found", requestID(r))
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(200)
		_, _ = w.Write(content)
	}
}
