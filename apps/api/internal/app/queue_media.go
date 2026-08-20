package app

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const queueMediaLimit = 220 << 20

var queueMediaTypes = map[string]map[string]bool{
	"AUDIO": {"audio/mpeg": true, "audio/wav": true, "audio/x-wav": true, "audio/ogg": true},
	"VIDEO": {"video/mp4": true, "video/webm": true},
	"BANNER": {"image/jpeg": true, "image/png": true, "image/webp": true},
	"IMAGE": {"image/jpeg": true, "image/png": true, "image/webp": true},
	"LOGO": {"image/jpeg": true, "image/png": true, "image/webp": true},
	"QR": {"image/jpeg": true, "image/png": true, "image/webp": true},
	"NOTIFICATION": {"audio/mpeg": true, "audio/wav": true, "audio/x-wav": true, "audio/ogg": true},
}

func (a *App) listQueueMedia(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	rows, err := a.db.Query(r.Context(), `SELECT id,kind,title,asset_key,language,content_type,file_name,byte_size,sort_order,is_active,starts_at,ends_at,duration_seconds,created_at FROM queue_media_assets WHERE organization_id=$1 ORDER BY kind,sort_order,title`, p.OrganizationID)
	if err != nil { errorJSON(w, 500, "database_error", "Could not load media", requestID(r)); return }
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, kind, title, contentType, fileName string
		var key, language *string
		var size int64
		var order int
		var active bool
		var starts, ends *time.Time
		var duration *int
		var created time.Time
		if rows.Scan(&id,&kind,&title,&key,&language,&contentType,&fileName,&size,&order,&active,&starts,&ends,&duration,&created)==nil {
			items=append(items,map[string]any{"id":id,"kind":kind,"title":title,"assetKey":key,"language":language,"contentType":contentType,"fileName":fileName,"byteSize":size,"sortOrder":order,"isActive":active,"startsAt":starts,"endsAt":ends,"durationSeconds":duration,"createdAt":created})
		}
	}
	writeJSON(w,200,map[string]any{"items":items})
}

func queueMediaFile(header *multipart.FileHeader, kind string) (multipart.File, string, error) {
	allowed, ok := queueMediaTypes[kind]
	if !ok { return nil,"",errors.New("invalid kind") }
	f, err := header.Open(); if err != nil { return nil,"",err }
	buf := make([]byte,512); n, readErr := io.ReadFull(f,buf); if readErr!=nil && readErr!=io.ErrUnexpectedEOF { f.Close(); return nil,"",readErr }
	contentType:=http.DetectContentType(buf[:n])
	if !allowed[contentType] { f.Close(); return nil,"",errors.New("invalid content type") }
	if _,err=f.Seek(0,io.SeekStart);err!=nil { f.Close(); return nil,"",err }
	return f,contentType,nil
}

func mediaExtension(contentType string) string {
	switch contentType { case "audio/mpeg": return ".mp3"; case "audio/wav","audio/x-wav": return ".wav"; case "audio/ogg": return ".ogg"; case "video/mp4": return ".mp4"; case "video/webm": return ".webm"; case "image/jpeg": return ".jpg"; case "image/png": return ".png"; case "image/webp": return ".webp" }
	return ".bin"
}

func (a *App) uploadQueueMedia(w http.ResponseWriter,r *http.Request) {
	r.Body=http.MaxBytesReader(w,r.Body,queueMediaLimit)
	if err:=r.ParseMultipartForm(queueMediaLimit);err!=nil { errorJSON(w,413,"file_too_large","File is too large",requestID(r));return }
	kind:=strings.ToUpper(strings.TrimSpace(r.FormValue("kind"))); title:=strings.TrimSpace(r.FormValue("title")); language:=strings.TrimSpace(r.FormValue("language")); key:=strings.TrimSpace(r.FormValue("assetKey"))
	file,header,err:=r.FormFile("file"); if err!=nil { errorJSON(w,422,"file_required","Media file is required",requestID(r));return }; file.Close()
	if title=="" || header.Size<=0 || header.Size>queueMediaLimit { errorJSON(w,422,"validation_error","Media fields are invalid",requestID(r));return }
	source,contentType,err:=queueMediaFile(header,kind); if err!=nil { errorJSON(w,422,"unsupported_media","File format does not match media type",requestID(r));return }; defer source.Close()
	p:=auth.PrincipalFrom(r.Context()); id:=uuid.NewString(); dir:=filepath.Join(a.cfg.QueueMediaPath,p.OrganizationID); if err=os.MkdirAll(dir,0700);err!=nil { errorJSON(w,500,"storage_error","Could not prepare media storage",requestID(r));return }
	storedName:=id+mediaExtension(contentType); destination:=filepath.Join(dir,storedName); out,err:=os.OpenFile(destination,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600); if err!=nil { errorJSON(w,500,"storage_error","Could not save media",requestID(r));return }
	written,copyErr:=io.Copy(out,io.LimitReader(source,queueMediaLimit+1)); closeErr:=out.Close(); if copyErr!=nil||closeErr!=nil||written>queueMediaLimit { _=os.Remove(destination); errorJSON(w,500,"storage_error","Could not save media",requestID(r));return }
	order,_:=strconv.Atoi(r.FormValue("sortOrder")); var lang any=nil; if language!="" { if language!="ru"&&language!="uz"&&language!="en" { _=os.Remove(destination); errorJSON(w,422,"validation_error","Language is invalid",requestID(r));return }; lang=language }
	_,err=a.db.Exec(r.Context(),`INSERT INTO queue_media_assets(id,organization_id,kind,title,asset_key,language,content_type,file_name,storage_path,byte_size,sort_order,created_by) VALUES($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8,$9,$10,$11,$12)`,id,p.OrganizationID,kind,title,key,lang,contentType,filepath.Base(header.Filename),destination,written,order,p.UserID)
	if err!=nil { _=os.Remove(destination); errorJSON(w,409,"media_conflict","Could not register media asset",requestID(r));return }
	_,_=a.db.Exec(r.Context(),`INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result,changes) VALUES($1,$2,'QUEUE_MEDIA_CREATED','queue_media',$3,'success',jsonb_build_object('kind',$4::text,'bytes',$5::bigint))`,p.OrganizationID,p.UserID,id,kind,written)
	writeJSON(w,201,map[string]any{"id":id,"contentType":contentType,"byteSize":written})
}

func (a *App) deleteQueueMedia(w http.ResponseWriter,r *http.Request) {
	p:=auth.PrincipalFrom(r.Context()); id:=chi.URLParam(r,"id"); var path string
	tx,err:=a.db.Begin(r.Context()); if err!=nil { errorJSON(w,500,"database_error","Could not delete media",requestID(r));return }; defer tx.Rollback(r.Context())
	if err=tx.QueryRow(r.Context(),`DELETE FROM queue_media_assets WHERE id=$1 AND organization_id=$2 RETURNING storage_path`,id,p.OrganizationID).Scan(&path);err!=nil { errorJSON(w,404,"not_found","Media not found",requestID(r));return }
	_,_=tx.Exec(r.Context(),`INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'QUEUE_MEDIA_DELETED','queue_media',$3,'success')`,p.OrganizationID,p.UserID,id)
	if err=tx.Commit(r.Context());err!=nil { errorJSON(w,500,"database_error","Could not delete media",requestID(r));return }
	cleanRoot,_:=filepath.Abs(a.cfg.QueueMediaPath); cleanPath,_:=filepath.Abs(path); if strings.HasPrefix(cleanPath,cleanRoot+string(os.PathSeparator)) { _=os.Remove(cleanPath) }
	writeJSON(w,200,map[string]bool{"ok":true})
}

func (a *App) getPublicQueueMedia(w http.ResponseWriter,r *http.Request) {
	_,org,err:=a.resolveDisplay(r); if err!=nil { errorJSON(w,404,"display_not_found","Display not found",requestID(r));return }
	var path,contentType,fileName string
	err=a.db.QueryRow(r.Context(),`SELECT storage_path,content_type,file_name FROM queue_media_assets WHERE id=$1 AND organization_id=$2 AND is_active AND (starts_at IS NULL OR starts_at<=now()) AND (ends_at IS NULL OR ends_at>=now())`,chi.URLParam(r,"id"),org).Scan(&path,&contentType,&fileName)
	if err!=nil { errorJSON(w,404,"not_found","Media not found",requestID(r));return }
	w.Header().Set("Content-Type",contentType); w.Header().Set("Content-Disposition",`inline; filename="`+strings.ReplaceAll(fileName,`"`,"")+`"`); w.Header().Set("Cache-Control","private, max-age=3600"); http.ServeFile(w,r,path)
}

type queueTickerInput struct { Text string `json:"text"`; Language string `json:"language"`; Speed int `json:"speed"`; SortOrder int `json:"sortOrder"`; IsActive *bool `json:"isActive"`; StartsAt *time.Time `json:"startsAt"`; EndsAt *time.Time `json:"endsAt"` }

func (a *App) listQueueTickers(w http.ResponseWriter,r *http.Request) {
	p:=auth.PrincipalFrom(r.Context()); rows,err:=a.db.Query(r.Context(),`SELECT id,text,language,speed,sort_order,is_active,starts_at,ends_at,created_at FROM queue_ticker_messages WHERE organization_id=$1 ORDER BY sort_order,created_at DESC`,p.OrganizationID); if err!=nil { errorJSON(w,500,"database_error","Could not load ticker",requestID(r));return }; defer rows.Close(); items:=[]map[string]any{}
	for rows.Next(){var id,textValue,language string;var speed,order int;var active bool;var starts,ends *time.Time;var created time.Time;if rows.Scan(&id,&textValue,&language,&speed,&order,&active,&starts,&ends,&created)==nil{items=append(items,map[string]any{"id":id,"text":textValue,"language":language,"speed":speed,"sortOrder":order,"isActive":active,"startsAt":starts,"endsAt":ends,"createdAt":created})}}
	writeJSON(w,200,map[string]any{"items":items})
}

func (a *App) saveQueueTicker(w http.ResponseWriter,r *http.Request) {
	var in queueTickerInput;if !decode(w,r,&in){return};in.Text=strings.TrimSpace(in.Text);if in.Language==""{in.Language="ru"};if in.Speed==0{in.Speed=40};if in.Text==""||len([]rune(in.Text))>500||(in.Language!="ru"&&in.Language!="uz"&&in.Language!="en")||in.Speed<10||in.Speed>200{errorJSON(w,422,"validation_error","Ticker fields are invalid",requestID(r));return}
	p:=auth.PrincipalFrom(r.Context());id:=chi.URLParam(r,"id");if id==""{id=uuid.NewString()};active:=true;if in.IsActive!=nil{active=*in.IsActive};tag,err:=a.db.Exec(r.Context(),`INSERT INTO queue_ticker_messages(id,organization_id,text,language,speed,sort_order,is_active,starts_at,ends_at,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT(id) DO UPDATE SET text=EXCLUDED.text,language=EXCLUDED.language,speed=EXCLUDED.speed,sort_order=EXCLUDED.sort_order,is_active=EXCLUDED.is_active,starts_at=EXCLUDED.starts_at,ends_at=EXCLUDED.ends_at,updated_at=now() WHERE queue_ticker_messages.organization_id=$2`,id,p.OrganizationID,in.Text,in.Language,in.Speed,in.SortOrder,active,in.StartsAt,in.EndsAt,p.UserID);if err!=nil||tag.RowsAffected()==0{errorJSON(w,409,"ticker_conflict","Could not save ticker",requestID(r));return};writeJSON(w,200,map[string]string{"id":id})
}

func queueMediaJSON(v any) json.RawMessage { data,_:=json.Marshal(v);return data }
