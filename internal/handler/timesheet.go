package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// TimesheetHandler exposes HTTP endpoints for timesheet generation, lines, and approval workflows
type TimesheetHandler struct {
	svc service.TimesheetService
}

var GlobalTimesheetHandler *TimesheetHandler

// InitTimesheetHandler initializes the singleton timesheet handler
func InitTimesheetHandler(svc service.TimesheetService) {
	GlobalTimesheetHandler = &TimesheetHandler{svc: svc}
}

// NewTimesheetHandler creates a new TimesheetHandler instance
func NewTimesheetHandler(svc service.TimesheetService) *TimesheetHandler {
	return &TimesheetHandler{svc: svc}
}

func getTimesheetUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "user_1"
}

func writeTimesheetJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeTimesheetError(w http.ResponseWriter, status int, msg string) {
	writeTimesheetJSON(w, status, map[string]string{"error": msg})
}

func handleTimesheetServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrTimesheetNotFound):
		writeTimesheetError(w, http.StatusNotFound, "timesheet not found")
	case errors.Is(err, service.ErrTimesheetLineNotFound):
		writeTimesheetError(w, http.StatusNotFound, "timesheet line not found")
	case errors.Is(err, service.ErrWorkspaceNotFound):
		writeTimesheetError(w, http.StatusNotFound, "workspace not found")
	case errors.Is(err, service.ErrWorkspaceRequired):
		writeTimesheetError(w, http.StatusBadRequest, "workspace_id is required")
	case errors.Is(err, service.ErrUnauthorizedTimesheet):
		writeTimesheetError(w, http.StatusForbidden, "unauthorized to perform action on timesheet")
	case errors.Is(err, service.ErrInvalidPeriodDates):
		writeTimesheetError(w, http.StatusBadRequest, "invalid timesheet period dates: end date must be on or after start date")
	case errors.Is(err, service.ErrTimesheetAlreadyExists):
		writeTimesheetError(w, http.StatusConflict, "timesheet for this period already exists")
	case errors.Is(err, service.ErrTimesheetNotEditable):
		writeTimesheetError(w, http.StatusBadRequest, "timesheet cannot be edited in current status")
	case errors.Is(err, service.ErrTimesheetNotSubmittable):
		writeTimesheetError(w, http.StatusBadRequest, "timesheet must be in draft status to submit")
	case errors.Is(err, service.ErrTimesheetNotReviewable):
		writeTimesheetError(w, http.StatusBadRequest, "timesheet must be in submitted status to review")
	case errors.Is(err, service.ErrInvalidLineDuration):
		writeTimesheetError(w, http.StatusBadRequest, "duration must be greater than zero")
	case errors.Is(err, service.ErrInvalidWorkDate):
		writeTimesheetError(w, http.StatusBadRequest, "work date is required and must fall within timesheet period")
	case errors.Is(err, service.ErrRejectionReasonRequired):
		writeTimesheetError(w, http.StatusBadRequest, "rejection reason is required")
	default:
		writeTimesheetError(w, http.StatusInternalServerError, err.Error())
	}
}

// CreateTimesheet handles POST /v2/timesheets
func (h *TimesheetHandler) CreateTimesheet(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTimesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTimesheetError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getTimesheetUser(r)

	ts, err := h.svc.CreateTimesheet(userID, req)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusCreated, ts)
}

// GetTimesheet handles GET /v2/timesheets/{id}
func (h *TimesheetHandler) GetTimesheet(w http.ResponseWriter, r *http.Request, id int64) {
	ts, err := h.svc.GetTimesheet(id)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, ts)
}

// ListTimesheets handles GET /v2/timesheets
func (h *TimesheetHandler) ListTimesheets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	wsID, _ := strconv.ParseInt(q.Get("workspace_id"), 10, 64)

	filter := model.TimesheetFilter{
		WorkspaceID: wsID,
	}

	if uid := q.Get("user_id"); uid != "" {
		filter.UserID = &uid
	}
	if status := q.Get("status"); status != "" {
		filter.Status = &status
	}
	if start := q.Get("start_date"); start != "" {
		filter.StartDate = &start
	}
	if end := q.Get("end_date"); end != "" {
		filter.EndDate = &end
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if offsetStr := q.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	list, err := h.svc.ListTimesheets(filter)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, list)
}

// UpdateTimesheet handles PUT /v2/timesheets/{id}
func (h *TimesheetHandler) UpdateTimesheet(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateTimesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTimesheetError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getTimesheetUser(r)

	ts, err := h.svc.UpdateTimesheet(id, userID, req)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, ts)
}

// DeleteTimesheet handles DELETE /v2/timesheets/{id}
func (h *TimesheetHandler) DeleteTimesheet(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getTimesheetUser(r)
	if err := h.svc.DeleteTimesheet(id, userID); err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, map[string]string{"message": "timesheet deleted successfully"})
}

// PopulateFromTimeEntries handles POST /v2/timesheets/{id}/populate
func (h *TimesheetHandler) PopulateFromTimeEntries(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.PopulateTimesheetRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	userID := getTimesheetUser(r)

	ts, err := h.svc.PopulateFromTimeEntries(id, userID, req)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, ts)
}

// AddLine handles POST /v2/timesheets/{id}/lines
func (h *TimesheetHandler) AddLine(w http.ResponseWriter, r *http.Request, timesheetID int64) {
	var req model.TimesheetLineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTimesheetError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getTimesheetUser(r)

	line, err := h.svc.AddLine(timesheetID, userID, req)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusCreated, line)
}

// ListLines handles GET /v2/timesheets/{id}/lines
func (h *TimesheetHandler) ListLines(w http.ResponseWriter, r *http.Request, timesheetID int64) {
	lines, err := h.svc.ListLines(timesheetID)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, lines)
}

// UpdateLine handles PUT /v2/timesheets/{id}/lines/{line_id}
func (h *TimesheetHandler) UpdateLine(w http.ResponseWriter, r *http.Request, timesheetID, lineID int64) {
	var req model.TimesheetLineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTimesheetError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getTimesheetUser(r)

	line, err := h.svc.UpdateLine(timesheetID, lineID, userID, req)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, line)
}

// DeleteLine handles DELETE /v2/timesheets/{id}/lines/{line_id}
func (h *TimesheetHandler) DeleteLine(w http.ResponseWriter, r *http.Request, timesheetID, lineID int64) {
	userID := getTimesheetUser(r)
	if err := h.svc.DeleteLine(timesheetID, lineID, userID); err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, map[string]string{"message": "timesheet line deleted successfully"})
}

// SubmitTimesheet handles POST /v2/timesheets/{id}/submit
func (h *TimesheetHandler) SubmitTimesheet(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.SubmitTimesheetRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	userID := getTimesheetUser(r)

	ts, err := h.svc.SubmitTimesheet(id, userID, req)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, ts)
}

// ApproveTimesheet handles POST /v2/timesheets/{id}/approve
func (h *TimesheetHandler) ApproveTimesheet(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.ReviewTimesheetRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	reviewerID := getTimesheetUser(r)

	ts, err := h.svc.ApproveTimesheet(id, reviewerID, req)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, ts)
}

// RejectTimesheet handles POST /v2/timesheets/{id}/reject
func (h *TimesheetHandler) RejectTimesheet(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.ReviewTimesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTimesheetError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	reviewerID := getTimesheetUser(r)

	ts, err := h.svc.RejectTimesheet(id, reviewerID, req)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, ts)
}

// ReopenTimesheet handles POST /v2/timesheets/{id}/reopen
func (h *TimesheetHandler) ReopenTimesheet(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Comments string `json:"comments"`
	}
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	reviewerID := getTimesheetUser(r)

	ts, err := h.svc.ReopenTimesheet(id, reviewerID, body.Comments)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, ts)
}

// ListApprovals handles GET /v2/timesheets/{id}/approvals
func (h *TimesheetHandler) ListApprovals(w http.ResponseWriter, r *http.Request, timesheetID int64) {
	approvals, err := h.svc.ListApprovals(timesheetID)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, approvals)
}

// GetTimesheetMetrics handles GET /v2/timesheets/metrics
func (h *TimesheetHandler) GetTimesheetMetrics(w http.ResponseWriter, r *http.Request) {
	wsID, _ := strconv.ParseInt(r.URL.Query().Get("workspace_id"), 10, 64)
	metrics, err := h.svc.GetTimesheetMetrics(wsID)
	if err != nil {
		handleTimesheetServiceError(w, err)
		return
	}
	writeTimesheetJSON(w, http.StatusOK, metrics)
}
