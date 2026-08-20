package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func principal(r *http.Request) auth.Principal { return auth.PrincipalFrom(r.Context()) }

type documentTemplateInput struct {
	Name string `json:"name"`
	DocumentType string `json:"documentType"`
	PageSize string `json:"pageSize"`
	WidthMM float64 `json:"widthMm"`
	HeightMM float64 `json:"heightMm"`
	Layout json.RawMessage `json:"layout"`
	IsDefault bool `json:"isDefault"`
}

func validTemplate(in documentTemplateInput) bool {
	types := map[string]bool{"ASSIGNMENT":true,"PRESCRIPTION":true,"RECEIPT":true,"OTHER":true}
	sizes := map[string]bool{"A4":true,"A5":true,"RECEIPT_80":true,"CUSTOM":true}
	return len(strings.TrimSpace(in.Name)) >= 2 && types[in.DocumentType] && sizes[in.PageSize] && in.WidthMM >= 40 && in.WidthMM <= 500 && in.HeightMM >= 40 && in.HeightMM <= 1000 && len(in.Layout) > 0 && len(in.Layout) <= 2*1024*1024 && json.Valid(in.Layout)
}

func (a *App) listDocumentTemplates(w http.ResponseWriter, r *http.Request) {
	p:=principal(r); rows,err:=a.db.Query(r.Context(),`SELECT id,name,document_type,page_size,width_mm,height_mm,layout,is_default,created_at,updated_at FROM document_templates WHERE organization_id=$1 ORDER BY document_type,is_default DESC,name`,p.OrganizationID)
	if err!=nil { errorJSON(w,500,"database_error","Could not load templates",requestID(r)); return }; defer rows.Close()
	items:=[]map[string]any{}; for rows.Next(){var id,name,kind,size string;var width,height float64;var layout json.RawMessage;var def bool;var created,updated any;if rows.Scan(&id,&name,&kind,&size,&width,&height,&layout,&def,&created,&updated)!=nil{continue};items=append(items,map[string]any{"id":id,"name":name,"documentType":kind,"pageSize":size,"widthMm":width,"heightMm":height,"layout":layout,"isDefault":def,"createdAt":created,"updatedAt":updated})}; writeJSON(w,200,map[string]any{"items":items})
}

func (a *App) createDocumentTemplate(w http.ResponseWriter,r *http.Request){
	p:=principal(r);var in documentTemplateInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,2200000)).Decode(&in)!=nil||!validTemplate(in){errorJSON(w,422,"validation_error","Invalid document template",requestID(r));return};id:=uuid.NewString();tx,e:=a.db.Begin(r.Context());if e!=nil{errorJSON(w,500,"database_error","Could not save template",requestID(r));return};defer tx.Rollback(r.Context());if in.IsDefault{_,e=tx.Exec(r.Context(),`UPDATE document_templates SET is_default=false,updated_at=now() WHERE organization_id=$1 AND document_type=$2`,p.OrganizationID,in.DocumentType)};if e==nil{_,e=tx.Exec(r.Context(),`INSERT INTO document_templates(id,organization_id,name,document_type,page_size,width_mm,height_mm,layout,is_default,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,id,p.OrganizationID,strings.TrimSpace(in.Name),in.DocumentType,in.PageSize,in.WidthMM,in.HeightMM,in.Layout,in.IsDefault,p.UserID)};if e!=nil{errorJSON(w,500,"database_error","Could not save template",requestID(r));return};_,_=tx.Exec(r.Context(),`INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'DOCUMENT_TEMPLATE_CREATED','document_template',$3,'success')`,p.OrganizationID,p.UserID,id);if tx.Commit(r.Context())!=nil{errorJSON(w,500,"database_error","Could not save template",requestID(r));return};writeJSON(w,201,map[string]string{"id":id})
}

func (a *App) updateDocumentTemplate(w http.ResponseWriter,r *http.Request){
	p:=principal(r);id:=chi.URLParam(r,"id");if _,e:=uuid.Parse(id);e!=nil{errorJSON(w,404,"not_found","Template not found",requestID(r));return};var in documentTemplateInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,2200000)).Decode(&in)!=nil||!validTemplate(in){errorJSON(w,422,"validation_error","Invalid document template",requestID(r));return};tx,e:=a.db.Begin(r.Context());if e!=nil{errorJSON(w,500,"database_error","Could not update template",requestID(r));return};defer tx.Rollback(r.Context());if in.IsDefault{_,e=tx.Exec(r.Context(),`UPDATE document_templates SET is_default=false,updated_at=now() WHERE organization_id=$1 AND document_type=$2 AND id<>$3`,p.OrganizationID,in.DocumentType,id)};if e==nil{_,e=tx.Exec(r.Context(),`UPDATE document_templates SET name=$1,document_type=$2,page_size=$3,width_mm=$4,height_mm=$5,layout=$6,is_default=$7,updated_at=now() WHERE id=$8 AND organization_id=$9`,strings.TrimSpace(in.Name),in.DocumentType,in.PageSize,in.WidthMM,in.HeightMM,in.Layout,in.IsDefault,id,p.OrganizationID)};if e!=nil{errorJSON(w,500,"database_error","Could not update template",requestID(r));return};_,_=tx.Exec(r.Context(),`INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'DOCUMENT_TEMPLATE_UPDATED','document_template',$3,'success')`,p.OrganizationID,p.UserID,id);if tx.Commit(r.Context())!=nil{errorJSON(w,500,"database_error","Could not update template",requestID(r));return};writeJSON(w,200,map[string]bool{"ok":true})
}

func (a *App) deleteDocumentTemplate(w http.ResponseWriter,r *http.Request){p:=principal(r);id:=chi.URLParam(r,"id");tag,e:=a.db.Exec(r.Context(),`DELETE FROM document_templates WHERE id=$1 AND organization_id=$2`,id,p.OrganizationID);if e!=nil{errorJSON(w,409,"template_in_use","Template cannot be deleted",requestID(r));return};if tag.RowsAffected()==0{errorJSON(w,404,"not_found","Template not found",requestID(r));return};_,_=a.db.Exec(r.Context(),`INSERT INTO audit_logs(organization_id,actor_id,action,entity_type,entity_id,result) VALUES($1,$2,'DOCUMENT_TEMPLATE_DELETED','document_template',$3,'success')`,p.OrganizationID,p.UserID,id);writeJSON(w,200,map[string]bool{"ok":true})}
