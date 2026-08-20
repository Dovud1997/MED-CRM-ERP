package app

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"clinicos/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func errorJSON(w http.ResponseWriter, status int, code, msg, requestID string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg, "requestId": requestID}})
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(a.recoverer, a.requestID, a.securityHeaders)
	r.Get("/api/v1/health/live", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	r.Get("/api/v1/health/ready", a.ready)
	r.Post("/api/v1/auth/login", a.login)
	r.Post("/api/v1/auth/refresh", a.refresh)
	r.Get("/api/v1/display/{token}", a.getQueueDisplayState)
	r.Get("/api/v1/display/{token}/events", a.streamQueueDisplay)
	r.Get("/api/v1/display/{token}/media/{id}", a.getPublicQueueMedia)
	r.Group(func(p chi.Router) {
		p.Use(a.authenticate)
		p.Post("/api/v1/auth/logout", a.logout)
		p.Get("/api/v1/auth/me", a.me)
		p.Get("/api/v1/branches", a.require("branches:read", a.listBranches))
		p.Post("/api/v1/branches", a.require("branches:write", a.createBranch))
		p.Patch("/api/v1/branches/{id}", a.require("branches:write", a.updateBranch))
		p.Delete("/api/v1/branches/{id}", a.require("branches:write", a.deleteBranch))
		p.Get("/api/v1/employees", a.require("employees:read", a.listEmployees))
		p.Post("/api/v1/employees", a.require("employees:write", a.createEmployee))
		p.Patch("/api/v1/employees/{id}", a.require("employees:write", a.updateEmployee))
		p.Patch("/api/v1/employees/{id}/status", a.require("employees:write", a.updateEmployeeStatus))
		p.Delete("/api/v1/employees/{id}", a.require("employees:write", a.deleteEmployee))
		p.Post("/api/v1/employees/{id}/photo", a.require("employees:write", a.saveProfilePhoto("employee")))
		p.Get("/api/v1/employees/{id}/photo", a.require("employees:read", a.getProfilePhoto("employee")))
		p.Get("/api/v1/roles", a.require("roles:read", a.listRoles))
		p.Post("/api/v1/roles", a.require("roles:write", a.createRole))
		p.Patch("/api/v1/roles/{id}", a.require("roles:write", a.updateRole))
		p.Delete("/api/v1/roles/{id}", a.require("roles:write", a.deleteRole))
		p.Get("/api/v1/permissions", a.require("roles:read", a.listPermissions))
		p.Get("/api/v1/patients", a.require("patients:read", a.listPatients))
		p.Post("/api/v1/patients", a.require("patients:write", a.createPatient))
		p.Get("/api/v1/patients/{id}", a.require("patients:read", a.getPatient))
		p.Patch("/api/v1/patients/{id}", a.require("patients:write", a.updatePatient))
		p.Delete("/api/v1/patients/{id}", a.require("patients:write", a.deletePatient))
		p.Post("/api/v1/patients/{id}/photo", a.require("patients:write", a.saveProfilePhoto("patient")))
		p.Get("/api/v1/patients/{id}/photo", a.require("patients:read", a.getProfilePhoto("patient")))
		p.Get("/api/v1/appointments", a.require("appointments:read", a.listAppointments))
		p.Get("/api/v1/appointments/dashboard", a.require("appointments:read", a.getAppointmentsDashboard))
		p.Post("/api/v1/appointments", a.require("appointments:write", a.createAppointment))
		p.Patch("/api/v1/appointments/{id}/status", a.require("appointments:write", a.updateAppointmentStatus))
		p.Get("/api/v1/doctor-schedules/{employeeId}", a.require("appointments:read", a.getDoctorSchedule))
		p.Put("/api/v1/doctor-schedules/{employeeId}", a.require("appointments:write", a.saveDoctorSchedule))
		p.Get("/api/v1/patients/{id}/clinical", a.require("clinical:read", a.getClinicalRecord))
		p.Put("/api/v1/patients/{id}/clinical-profile", a.require("clinical:write", a.updateClinicalProfile))
		p.Put("/api/v1/patients/{id}/blood-group", a.require("clinical:write", a.updateBloodGroup))
		p.Post("/api/v1/patients/{id}/allergies", a.require("clinical:write", a.addAllergy))
		p.Post("/api/v1/patients/{id}/vaccinations", a.require("clinical:write", a.addVaccination))
		p.Post("/api/v1/patients/{id}/lab-results", a.require("clinical:write", a.addLabResult))
		p.Post("/api/v1/patients/{id}/urine-panel", a.require("clinical:write", a.addUrinePanel))
		p.Post("/api/v1/patients/{id}/blood-panel", a.require("clinical:write", a.addBloodPanel))
		p.Get("/api/v1/lab-attachments/{id}", a.require("clinical:read", a.getLabAttachment))
		p.Get("/api/v1/laboratory/tests", a.require("clinical:read", a.listLaboratoryTests))
		p.Post("/api/v1/laboratory/tests", a.require("clinical:write", a.createLaboratoryTest))
		p.Patch("/api/v1/laboratory/tests/{id}", a.require("clinical:write", a.updateLaboratoryTest))
		p.Get("/api/v1/laboratory/orders", a.require("clinical:read", a.listLaboratoryOrders))
		p.Post("/api/v1/laboratory/orders", a.require("clinical:write", a.createLaboratoryOrder))
		p.Patch("/api/v1/laboratory/orders/{id}/status", a.require("clinical:write", a.updateLaboratoryOrderStatus))
		p.Post("/api/v1/laboratory/orders/{id}/results", a.require("clinical:write", a.completeLaboratoryOrder))
		p.Post("/api/v1/patients/{id}/imaging", a.require("clinical:write", a.addImagingStudy))
		p.Get("/api/v1/imaging/{id}/file", a.require("clinical:read", a.getImagingFile))
		p.Post("/api/v1/patients/{id}/diagnoses", a.require("clinical:write", a.addPatientDiagnosis))
		p.Patch("/api/v1/diagnoses/{id}/status", a.require("clinical:write", a.updateDiagnosisStatus))
		p.Post("/api/v1/patients/{id}/orders", a.require("clinical:write", a.addClinicalOrder))
		p.Patch("/api/v1/orders/{id}/status", a.require("clinical:write", a.updateClinicalOrderStatus))
		p.Get("/api/v1/diagnosis-catalog", a.require("specialists:read", a.listDiagnosisCatalog))
		p.Post("/api/v1/diagnosis-catalog", a.require("specialists:write", a.createDiagnosisCatalogItem))
		p.Patch("/api/v1/diagnosis-catalog/{id}", a.require("specialists:write", a.updateDiagnosisCatalogItem))
		p.Delete("/api/v1/diagnosis-catalog/{id}", a.require("specialists:write", a.deleteDiagnosisCatalogItem))
		p.Post("/api/v1/patients/{id}/history", a.require("clinical:write", a.addClinicalHistory))
		p.Post("/api/v1/backups/export", a.require("backup:export", a.exportBackup))
		p.Get("/api/v1/audit", a.require("audit:read", a.listAudit))
		p.Get("/api/v1/messages/contacts", a.require("messages:read", a.listMessageContacts))
		p.Get("/api/v1/messages/{userId}", a.require("messages:read", a.listMessages))
		p.Post("/api/v1/messages/{userId}", a.require("messages:write", a.sendMessage))
		p.Get("/api/v1/messages-archive", a.require("audit:read", a.listMessageArchive))
		p.Get("/api/v1/services", a.require("services:read", a.listServices))
		p.Post("/api/v1/services", a.require("services:write", a.createService))
		p.Patch("/api/v1/services/{id}", a.require("services:write", a.updateService))
		p.Delete("/api/v1/services/{id}", a.require("services:write", a.deleteService))
		p.Post("/api/v1/services/{id}/prices", a.require("services:write", a.upsertServicePrice))
		p.Put("/api/v1/services/{id}/providers", a.require("services:write", a.saveServiceProviders))
		p.Get("/api/v1/specialists", a.require("specialists:read", a.listSpecialties))
		p.Post("/api/v1/specialists", a.require("specialists:write", a.createSpecialty))
		p.Patch("/api/v1/specialists/{id}", a.require("specialists:write", a.updateSpecialty))
		p.Delete("/api/v1/specialists/{id}", a.require("specialists:write", a.deleteSpecialty))
		p.Get("/api/v1/inpatient/rooms", a.require("inpatient:read", a.listInpatientRooms))
		p.Post("/api/v1/inpatient/rooms", a.require("inpatient:write", a.createInpatientRoom))
		p.Patch("/api/v1/inpatient/rooms/{id}", a.require("inpatient:write", a.updateInpatientRoom))
		p.Delete("/api/v1/inpatient/rooms/{id}", a.require("inpatient:write", a.disableInpatientRoom))
		p.Post("/api/v1/inpatient/rooms/{id}/beds", a.require("inpatient:write", a.addInpatientBed))
		p.Patch("/api/v1/inpatient/beds/{id}/status", a.require("inpatient:write", a.updateInpatientBedStatus))
		p.Delete("/api/v1/inpatient/beds/{id}", a.require("inpatient:write", a.disableInpatientBed))
		p.Get("/api/v1/inpatient/bookings", a.require("inpatient:read", a.listInpatientBookings))
		p.Post("/api/v1/inpatient/bookings", a.require("inpatient:write", a.createInpatientBooking))
		p.Patch("/api/v1/inpatient/bookings/{id}/status", a.require("inpatient:write", a.updateInpatientBookingStatus))
		p.Get("/api/v1/reports/summary", a.require("reports:read", a.getReportSummary))
		p.Get("/api/v1/cash", a.require("finance:read", a.getCashDesk))
		p.Post("/api/v1/cash/shifts", a.require("finance:write", a.openCashShift))
		p.Post("/api/v1/cash/shifts/{id}/close", a.require("finance:write", a.closeCashShift))
		p.Post("/api/v1/cash/transactions", a.require("finance:write", a.createCashTransaction))
		p.Get("/api/v1/cash/patient-account", a.require("finance:read", a.getPatientAccount))
		p.Get("/api/v1/cash/debts/today", a.require("finance:read", a.getTodayPatientDebts))
		p.Get("/api/v1/accounting", a.require("finance:read", a.getAccounting))
		p.Post("/api/v1/accounting/entries", a.require("finance:write", a.createAccountingEntry))
		p.Post("/api/v1/accounting/obligations", a.require("finance:write", a.createObligation))
		p.Post("/api/v1/accounting/obligations/{id}/pay", a.require("finance:write", a.payObligation))
		p.Post("/api/v1/accounting/payroll", a.require("finance:write", a.createPayroll))
		p.Post("/api/v1/accounting/payroll/{id}/pay", a.require("finance:write", a.payPayroll))
		p.Get("/api/v1/document-templates", a.require("clinical:read", a.listDocumentTemplates))
		p.Post("/api/v1/document-templates", a.require("clinical:write", a.createDocumentTemplate))
		p.Patch("/api/v1/document-templates/{id}", a.require("clinical:write", a.updateDocumentTemplate))
		p.Delete("/api/v1/document-templates/{id}", a.require("clinical:write", a.deleteDocumentTemplate))
		p.Get("/api/v1/inventory/items", a.require("inventory:read", a.listInventoryItems))
		p.Post("/api/v1/inventory/items", a.require("inventory:write", a.createInventoryItem))
		p.Patch("/api/v1/inventory/items/{id}", a.require("inventory:write", a.updateInventoryItem))
		p.Get("/api/v1/inventory/movements", a.require("inventory:read", a.listInventoryMovements))
		p.Post("/api/v1/inventory/movements", a.require("inventory:write", a.createInventoryMovement))
		p.Get("/api/v1/inventory/assets", a.require("inventory:read", a.listInventoryAssets))
		p.Post("/api/v1/inventory/assets", a.require("inventory:write", a.saveInventoryAsset))
		p.Patch("/api/v1/inventory/assets/{id}", a.require("inventory:write", a.saveInventoryAsset))
		p.Get("/api/v1/inventory/counts", a.require("inventory:read", a.listInventoryCounts))
		p.Post("/api/v1/inventory/counts", a.require("inventory:write", a.createInventoryCount))
		p.Post("/api/v1/inventory/counts/{id}/complete", a.require("inventory:write", a.completeInventoryCount))
		p.Get("/api/v1/queue", a.require("queue:read", a.listQueueEntries))
		p.Post("/api/v1/queue", a.require("queue:manage", a.createQueueEntry))
		p.Post("/api/v1/queue/next/call", a.require("queue:call", a.callNextQueueEntry))
		p.Post("/api/v1/queue/{id}/{action}", a.require("queue:call", a.queueEntryAction))
		p.Get("/api/v1/queue/history", a.require("queue:read", a.listQueueHistory))
		p.Get("/api/v1/queue/rooms", a.require("queue:read", a.listQueueRooms))
		p.Post("/api/v1/queue/rooms", a.require("queue:settings", a.saveQueueRoom))
		p.Patch("/api/v1/queue/rooms/{id}", a.require("queue:settings", a.saveQueueRoom))
		p.Get("/api/v1/queue/displays", a.require("queue:read", a.listQueueDisplays))
		p.Post("/api/v1/queue/displays", a.require("queue:display", a.createQueueDisplay))
		p.Patch("/api/v1/queue/displays/{id}", a.require("queue:display", a.updateQueueDisplay))
		p.Post("/api/v1/queue/displays/{id}/token", a.require("queue:display", a.rotateQueueDisplayToken))
		p.Get("/api/v1/queue/media", a.require("queue:read", a.listQueueMedia))
		p.Post("/api/v1/queue/media", a.require("queue:media", a.uploadQueueMedia))
		p.Delete("/api/v1/queue/media/{id}", a.require("queue:media", a.deleteQueueMedia))
		p.Get("/api/v1/queue/tickers", a.require("queue:read", a.listQueueTickers))
		p.Post("/api/v1/queue/tickers", a.require("queue:settings", a.saveQueueTicker))
		p.Patch("/api/v1/queue/tickers/{id}", a.require("queue:settings", a.saveQueueTicker))
	})
	return r
}
func (a *App) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if _, err := uuid.Parse(id); err != nil {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}
func (a *App) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
func (a *App) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				a.log.Error("panic", "requestId", requestID(r), "panic", v, "stack", string(debug.Stack()))
				errorJSON(w, 500, "internal_error", "Internal server error", requestID(r))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func requestID(r *http.Request) string { v, _ := r.Context().Value(requestIDKey).(string); return v }
func (a *App) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if a.db.Ping(ctx) != nil || a.redis.Ping(ctx).Err() != nil {
		errorJSON(w, 503, "not_ready", "Dependencies unavailable", requestID(r))
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 40<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		errorJSON(w, 400, "invalid_request", "Invalid JSON body", requestID(r))
		return false
	}
	return true
}
func (a *App) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		principal, err := a.auth.VerifyAccess(r.Context(), token)
		if err != nil {
			errorJSON(w, 401, "unauthorized", "Authentication required", requestID(r))
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
	})
}
func (a *App) require(permission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.PrincipalFrom(r.Context()).Has(permission) {
			errorJSON(w, 403, "forbidden", "Insufficient permissions", requestID(r))
			return
		}
		next(w, r)
	}
}
