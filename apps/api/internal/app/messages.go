package app

import (
	"net/http"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (a *App) listMessageArchive(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	var owner bool
	_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id=ur.role_id WHERE ur.user_id=$1 AND ro.organization_id=$2 AND ro.code='OWNER')`, p.UserID, p.OrganizationID).Scan(&owner)
	if !owner {
		errorJSON(w, 403, "owner_only", "Only the clinic owner can view message history", requestID(r))
		return
	}
	from := time.Now().AddDate(0, -3, 0)
	to := time.Now().Add(24 * time.Hour)
	if value := r.URL.Query().Get("from"); value != "" {
		if parsed, err := time.Parse("2006-01-02", value); err == nil {
			from = parsed
		}
	}
	if value := r.URL.Query().Get("to"); value != "" {
		if parsed, err := time.Parse("2006-01-02", value); err == nil {
			to = parsed.Add(24 * time.Hour)
		}
	}
	participant := r.URL.Query().Get("participant")
	if participant != "" {
		if _, err := uuid.Parse(participant); err != nil {
			errorJSON(w, 422, "invalid_participant", "Invalid participant", requestID(r))
			return
		}
	}
	rows, err := a.db.Query(r.Context(), `SELECT m.id,m.sender_id,concat(es.last_name,' ',es.first_name),m.recipient_id,concat(er.last_name,' ',er.first_name),m.body,m.read_at,m.created_at FROM internal_messages m JOIN employees es ON es.user_id=m.sender_id JOIN employees er ON er.user_id=m.recipient_id WHERE m.organization_id=$1 AND m.created_at >= $2 AND m.created_at < $3 AND ($4='' OR m.sender_id::text=$4 OR m.recipient_id::text=$4) ORDER BY m.created_at DESC LIMIT 500`, p.OrganizationID, from, to, participant)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load message archive", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, senderID, senderName, recipientID, recipientName, body string
		var readAt any
		var createdAt time.Time
		if rows.Scan(&id, &senderID, &senderName, &recipientID, &recipientName, &body, &readAt, &createdAt) == nil {
			items = append(items, map[string]any{"id": id, "senderId": senderID, "senderName": senderName, "recipientId": recipientID, "recipientName": recipientName, "body": body, "readAt": readAt, "createdAt": createdAt})
		}
	}
	_, _ = a.db.Exec(r.Context(), `INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,result,changes) VALUES($1,$2,'MESSAGE_ARCHIVE_VIEWED','internal_message','success',jsonb_build_object('from',$3::text,'to',$4::text,'participant',$5::text,'count',$6::int))`, p.OrganizationID, p.UserID, from.Format(time.RFC3339), to.Format(time.RFC3339), participant, len(items))
	writeJSON(w, 200, map[string]any{"items": items, "retentionMonths": 3})
}

func (a *App) listMessageContacts(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT u.id,e.first_name,e.last_name,e.position,e.specialty,b.name,u.is_active,(SELECT count(*) FROM internal_messages m WHERE m.organization_id=$1 AND m.sender_id=u.id AND m.recipient_id=$2 AND m.read_at IS NULL) FROM users u JOIN employees e ON e.user_id=u.id LEFT JOIN branches b ON b.id=e.branch_id WHERE u.organization_id=$1 AND u.id<>$2 AND u.deleted_at IS NULL ORDER BY u.is_active DESC,e.last_name,e.first_name`, p.OrganizationID, p.UserID)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load message contacts", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, first, last, position string
		var specialty, branch *string
		var active bool
		var unread int
		if rows.Scan(&id, &first, &last, &position, &specialty, &branch, &active, &unread) == nil {
			items = append(items, map[string]any{"id": id, "name": last + " " + first, "position": position, "specialty": specialty, "branch": branch, "isActive": active, "unread": unread})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (a *App) listMessages(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	peer := chi.URLParam(r, "userId")
	if _, err := uuid.Parse(peer); err != nil {
		errorJSON(w, 400, "invalid_id", "Invalid user id", requestID(r))
		return
	}
	var allowed bool
	_ = a.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND organization_id=$2 AND deleted_at IS NULL)`, peer, p.OrganizationID).Scan(&allowed)
	if !allowed || peer == p.UserID {
		errorJSON(w, 404, "not_found", "Employee not found", requestID(r))
		return
	}
	_, _ = a.db.Exec(r.Context(), `UPDATE internal_messages SET read_at=now() WHERE organization_id=$1 AND sender_id=$2 AND recipient_id=$3 AND read_at IS NULL`, p.OrganizationID, peer, p.UserID)
	rows, err := a.db.Query(r.Context(), `SELECT id,sender_id,recipient_id,body,read_at,created_at FROM (SELECT id,sender_id,recipient_id,body,read_at,created_at FROM internal_messages WHERE organization_id=$1 AND ((sender_id=$2 AND recipient_id=$3) OR (sender_id=$3 AND recipient_id=$2)) ORDER BY created_at DESC LIMIT 100) x ORDER BY created_at`, p.OrganizationID, p.UserID, peer)
	if err != nil {
		errorJSON(w, 500, "database_error", "Could not load messages", requestID(r))
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, sender, recipient, body string
		var readAt any
		var created any
		if rows.Scan(&id, &sender, &recipient, &body, &readAt, &created) == nil {
			items = append(items, map[string]any{"id": id, "senderId": sender, "recipientId": recipient, "body": body, "readAt": readAt, "createdAt": created, "mine": sender == p.UserID})
		}
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (a *App) sendMessage(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	peer := chi.URLParam(r, "userId")
	var in struct {
		Body string `json:"body"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Body = strings.TrimSpace(in.Body)
	if _, err := uuid.Parse(peer); err != nil || peer == p.UserID || len([]rune(in.Body)) < 1 || len([]rune(in.Body)) > 4000 {
		errorJSON(w, 422, "validation_error", "Message is invalid", requestID(r))
		return
	}
	var id string
	var created any
	err := a.db.QueryRow(r.Context(), `INSERT INTO internal_messages(organization_id,sender_id,recipient_id,body) SELECT $1,$2,u.id,$4 FROM users u WHERE u.id=$3 AND u.organization_id=$1 AND u.is_active AND u.deleted_at IS NULL RETURNING id,created_at`, p.OrganizationID, p.UserID, peer, in.Body).Scan(&id, &created)
	if err != nil {
		errorJSON(w, 404, "not_found", "Recipient not found", requestID(r))
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "senderId": p.UserID, "recipientId": peer, "body": in.Body, "createdAt": created, "mine": true})
}
