package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func floatPtr(v float64) *float64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func setupTestWorkloadSvc(t *testing.T) (WorkloadService, repository.WorkloadRepository, repository.ProjectRepository, repository.TaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_3', 'charlie', 'charlie@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-wl', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Product WS', 'prod-ws-wl', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Core API', 'Core Backend Service', '#4f46e5', datetime('now'), datetime('now')),
		       (102, 'user_1', 'Web Client', 'React Web Frontend', '#10b981', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	workloadRepo := repository.NewWorkloadRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	svc := NewWorkloadService(workloadRepo, projectRepo)
	return svc, workloadRepo, projectRepo, taskRepo, sqliteDB
}

func TestWorkloadService_ProfileCRUD(t *testing.T) {
	svc, _, _, _, _ := setupTestWorkloadSvc(t)

	// 1. Validation: Workspace required
	_, err := svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID: 0,
		UserID:      "user_1",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	// 2. Validation: User ID required
	_, err = svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID: 1,
		UserID:      "   ",
	})
	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("expected ErrUserIDRequired, got: %v", err)
	}

	// 3. Validation: Invalid Capacity
	_, err = svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID:         1,
		UserID:              "user_1",
		WeeklyCapacityHours: floatPtr(200.0), // > 168
		DailyCapacityHours:  floatPtr(8.0),
	})
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("expected ErrInvalidCapacity for weekly > 168, got: %v", err)
	}

	_, err = svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID:        1,
		UserID:             "user_1",
		DailyCapacityHours: floatPtr(30.0), // > 24
	})
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("expected ErrInvalidCapacity for daily > 24, got: %v", err)
	}

	_, err = svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID: 1,
		UserID:      "user_1",
		HourlyRate:  floatPtr(-10.0), // negative rate
	})
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("expected ErrInvalidCapacity for negative hourly rate, got: %v", err)
	}

	// 4. Create Profile (with custom skills and rate)
	profile, err := svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID:         1,
		UserID:              "user_1",
		WeeklyCapacityHours: floatPtr(40.0),
		DailyCapacityHours:  floatPtr(8.0),
		Skills:              []string{"Go", "PostgreSQL", "Docker"},
		HourlyRate:          floatPtr(75.50),
	})
	if err != nil {
		t.Fatalf("failed to create resource profile: %v", err)
	}
	if profile.ID == 0 || profile.UserID != "user_1" || profile.WorkspaceID != 1 {
		t.Fatalf("unexpected profile created: %+v", profile)
	}
	if len(profile.Skills) != 3 || profile.Skills[0] != "Go" {
		t.Fatalf("expected skills Go, PostgreSQL, Docker, got: %v", profile.Skills)
	}
	if profile.HourlyRate != 75.50 || !profile.IsActive {
		t.Fatalf("unexpected rate or active flag: rate=%v, active=%v", profile.HourlyRate, profile.IsActive)
	}

	// 5. Duplicate creation returns ErrProfileExists
	_, err = svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID: 1,
		UserID:      "user_1",
	})
	if !errors.Is(err, ErrProfileExists) {
		t.Fatalf("expected ErrProfileExists on duplicate profile creation, got: %v", err)
	}

	// 6. Get Profile by ID and by User ID
	gotByID, err := svc.GetProfile(profile.ID)
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if gotByID.ID != profile.ID || gotByID.UserID != "user_1" {
		t.Fatalf("profile mismatch: got %+v, expected ID %d", gotByID, profile.ID)
	}

	gotByUser, err := svc.GetProfileByUser(1, "user_1")
	if err != nil {
		t.Fatalf("GetProfileByUser failed: %v", err)
	}
	if gotByUser.ID != profile.ID {
		t.Fatalf("profile mismatch: got ID %d, expected %d", gotByUser.ID, profile.ID)
	}

	// 7. List Profiles with workspace and active filter
	profiles, err := svc.ListProfiles(1, nil)
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}

	// 8. Update Profile
	newWeekly := 35.0
	newDaily := 7.0
	newRate := 95.0
	isActive := false
	newSkills := []string{"Go", "Kubernetes"}
	updated, err := svc.UpdateProfile(profile.ID, model.UpdateResourceProfileRequest{
		WeeklyCapacityHours: &newWeekly,
		DailyCapacityHours:  &newDaily,
		HourlyRate:          &newRate,
		Skills:              &newSkills,
		IsActive:            &isActive,
	})
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}
	if updated.WeeklyCapacityHours != 35.0 || updated.DailyCapacityHours != 7.0 || updated.HourlyRate != 95.0 || updated.IsActive {
		t.Fatalf("unexpected updated fields: %+v", updated)
	}
	if len(updated.Skills) != 2 || updated.Skills[1] != "Kubernetes" {
		t.Fatalf("unexpected updated skills: %v", updated.Skills)
	}

	// 9. Delete Profile
	err = svc.DeleteProfile(profile.ID)
	if err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}
	_, err = svc.GetProfile(profile.ID)
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound after delete, got: %v", err)
	}
}

func TestWorkloadService_TimeOff(t *testing.T) {
	svc, _, _, _, _ := setupTestWorkloadSvc(t)

	now := time.Now().Truncate(24 * time.Hour)
	startDate := now.Add(24 * time.Hour)
	endDate := now.Add(3 * 24 * time.Hour)

	// 1. Validations
	_, err := svc.CreateTimeOff(model.CreateTimeOffRequest{
		WorkspaceID: 0,
		UserID:      "user_1",
		StartDate:   startDate,
		EndDate:     endDate,
		TimeOffType: model.TimeOffTypeVacation,
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	_, err = svc.CreateTimeOff(model.CreateTimeOffRequest{
		WorkspaceID: 1,
		UserID:      "",
		StartDate:   startDate,
		EndDate:     endDate,
		TimeOffType: model.TimeOffTypeVacation,
	})
	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("expected ErrUserIDRequired, got: %v", err)
	}

	_, err = svc.CreateTimeOff(model.CreateTimeOffRequest{
		WorkspaceID: 1,
		UserID:      "user_1",
		StartDate:   endDate,
		EndDate:     startDate, // End before start
		TimeOffType: model.TimeOffTypeVacation,
	})
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("expected ErrInvalidDateRange, got: %v", err)
	}

	_, err = svc.CreateTimeOff(model.CreateTimeOffRequest{
		WorkspaceID: 1,
		UserID:      "user_1",
		StartDate:   startDate,
		EndDate:     endDate,
		TimeOffType: "invalid_type",
	})
	if !errors.Is(err, ErrInvalidTimeOffType) {
		t.Fatalf("expected ErrInvalidTimeOffType, got: %v", err)
	}

	_, err = svc.CreateTimeOff(model.CreateTimeOffRequest{
		WorkspaceID: 1,
		UserID:      "user_1",
		StartDate:   startDate,
		EndDate:     endDate,
		TimeOffType: model.TimeOffTypeVacation,
		HoursPerDay: floatPtr(25.0), // > 24
	})
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("expected ErrInvalidCapacity for hours > 24, got: %v", err)
	}

	// 2. Create TimeOff with default 8 hours/day
	timeOff, err := svc.CreateTimeOff(model.CreateTimeOffRequest{
		WorkspaceID: 1,
		UserID:      "user_1",
		StartDate:   startDate,
		EndDate:     endDate,
		TimeOffType: model.TimeOffTypeVacation,
		Description: "Annual leave",
	})
	if err != nil {
		t.Fatalf("failed to create time-off: %v", err)
	}
	if timeOff.ID == 0 || timeOff.HoursPerDay != 8.0 || timeOff.Description != "Annual leave" {
		t.Fatalf("unexpected time-off: %+v", timeOff)
	}

	// 3. Get TimeOff by ID
	got, err := svc.GetTimeOff(timeOff.ID)
	if err != nil {
		t.Fatalf("GetTimeOff failed: %v", err)
	}
	if got.ID != timeOff.ID || got.UserID != "user_1" {
		t.Fatalf("time-off mismatch: got %+v", got)
	}

	// 4. List TimeOffs
	list, err := svc.ListTimeOffs(1, "user_1", nil, nil)
	if err != nil {
		t.Fatalf("ListTimeOffs failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 time-off, got %d", len(list))
	}

	// 5. Update TimeOff
	newHours := 4.0 // half-day
	newDesc := "Conference afternoon"
	updated, err := svc.UpdateTimeOff(timeOff.ID, model.UpdateTimeOffRequest{
		HoursPerDay: &newHours,
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("UpdateTimeOff failed: %v", err)
	}
	if updated.HoursPerDay != 4.0 || updated.Description != newDesc {
		t.Fatalf("unexpected updated time-off: %+v", updated)
	}

	// 6. Delete TimeOff
	err = svc.DeleteTimeOff(timeOff.ID)
	if err != nil {
		t.Fatalf("DeleteTimeOff failed: %v", err)
	}
	_, err = svc.GetTimeOff(timeOff.ID)
	if !errors.Is(err, ErrTimeOffNotFound) {
		t.Fatalf("expected ErrTimeOffNotFound after delete, got: %v", err)
	}
}

func TestWorkloadService_Allocations(t *testing.T) {
	svc, _, _, _, _ := setupTestWorkloadSvc(t)

	now := time.Now().Truncate(24 * time.Hour)
	startDate := now.Add(24 * time.Hour)
	endDate := now.Add(5 * 24 * time.Hour)

	// 1. Validations
	_, err := svc.CreateAllocation(model.CreateResourceAllocationRequest{
		WorkspaceID:           0,
		ProjectID:             101,
		UserID:                "user_1",
		StartDate:             startDate,
		EndDate:               endDate,
		AllocatedHoursPerWeek: 20.0,
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	_, err = svc.CreateAllocation(model.CreateResourceAllocationRequest{
		WorkspaceID:           1,
		ProjectID:             999, // non-existent project
		UserID:                "user_1",
		StartDate:             startDate,
		EndDate:               endDate,
		AllocatedHoursPerWeek: 20.0,
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got: %v", err)
	}

	_, err = svc.CreateAllocation(model.CreateResourceAllocationRequest{
		WorkspaceID:           1,
		ProjectID:             101,
		UserID:                "user_1",
		StartDate:             endDate,
		EndDate:               startDate, // End before start
		AllocatedHoursPerWeek: 20.0,
	})
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("expected ErrInvalidDateRange, got: %v", err)
	}

	_, err = svc.CreateAllocation(model.CreateResourceAllocationRequest{
		WorkspaceID:           1,
		ProjectID:             101,
		UserID:                "user_1",
		StartDate:             startDate,
		EndDate:               endDate,
		AllocatedHoursPerWeek: -5.0,
	})
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("expected ErrInvalidCapacity for negative hours, got: %v", err)
	}

	// 2. Create Allocation (default planned status)
	alloc, err := svc.CreateAllocation(model.CreateResourceAllocationRequest{
		WorkspaceID:           1,
		ProjectID:             101,
		UserID:                "user_1",
		StartDate:             startDate,
		EndDate:               endDate,
		AllocatedHoursPerWeek: 20.0,
		Notes:                 "Tech Lead commitment",
	})
	if err != nil {
		t.Fatalf("failed to create allocation: %v", err)
	}
	if alloc.ID == 0 || alloc.Status != model.AllocationStatusPlanned || alloc.Notes != "Tech Lead commitment" {
		t.Fatalf("unexpected allocation: %+v", alloc)
	}

	// 3. List Allocations
	allocs, err := svc.ListAllocations(1, 101, "", "")
	if err != nil {
		t.Fatalf("ListAllocations failed: %v", err)
	}
	if len(allocs) != 1 {
		t.Fatalf("expected 1 allocation, got %d", len(allocs))
	}

	// 4. Update Allocation
	newHours := 25.0
	newStatus := model.AllocationStatusCompleted
	updated, err := svc.UpdateAllocation(alloc.ID, model.UpdateResourceAllocationRequest{
		AllocatedHoursPerWeek: &newHours,
		Status:                &newStatus,
	})
	if err != nil {
		t.Fatalf("UpdateAllocation failed: %v", err)
	}
	if updated.AllocatedHoursPerWeek != 25.0 || updated.Status != model.AllocationStatusCompleted {
		t.Fatalf("unexpected updated allocation: %+v", updated)
	}

	// 5. Delete Allocation
	err = svc.DeleteAllocation(alloc.ID)
	if err != nil {
		t.Fatalf("DeleteAllocation failed: %v", err)
	}
	allocs, _ = svc.ListAllocations(1, 101, "", "")
	if len(allocs) != 0 {
		t.Fatalf("expected 0 allocations after delete, got %d", len(allocs))
	}
}

func TestWorkloadService_UserWorkloadAndHeatmap(t *testing.T) {
	svc, _, _, taskRepo, _ := setupTestWorkloadSvc(t)

	// 1. Create Profile for user_1: 8 hrs/day, 40 hrs/week
	_, err := svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID:         1,
		UserID:              "user_1",
		DailyCapacityHours:  floatPtr(8.0),
		WeeklyCapacityHours: floatPtr(40.0),
		Skills:              []string{"Backend", "Go"},
	})
	if err != nil {
		t.Fatalf("failed to create user_1 profile: %v", err)
	}

	// Date range: Monday 2026-11-09 to Sunday 2026-11-15
	from := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	// 2. Schedule Time-Off on 2026-11-11 (Wednesday) for 8 hours (vacation)
	wednesday := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	_, err = svc.CreateTimeOff(model.CreateTimeOffRequest{
		WorkspaceID: 1,
		UserID:      "user_1",
		StartDate:   wednesday,
		EndDate:     wednesday,
		TimeOffType: model.TimeOffTypeVacation,
		HoursPerDay: floatPtr(8.0),
	})
	if err != nil {
		t.Fatalf("failed to create time-off: %v", err)
	}

	// 3. Add Project Allocation of 10 hours/week from 2026-11-09 to 2026-11-13 (2 hrs/day)
	fri := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)
	_, err = svc.CreateAllocation(model.CreateResourceAllocationRequest{
		WorkspaceID:           1,
		ProjectID:             101,
		UserID:                "user_1",
		StartDate:             from,
		EndDate:               fri,
		AllocatedHoursPerWeek: 10.0,
	})
	if err != nil {
		t.Fatalf("failed to create allocation: %v", err)
	}

	// 4. Add Task assigned to user_1 due on 2026-11-10 with 360 estimate minutes (6 hours)
	tueDue := time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC)
	err = taskRepo.Create(&model.Task{
		Title:           "Implement API endpoint",
		Status:          model.TaskStatusInProgress,
		Priority:        "high",
		UserID:          "user_1",
		EstimateMinutes: 360, // 6.0 hours
		DueDate:         tueDue,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// 5. Calculate User Workload
	workload, err := svc.GetUserWorkload(1, "user_1", from, to)
	if err != nil {
		t.Fatalf("GetUserWorkload failed: %v", err)
	}

	// Available capacity: 5 weekdays * 8 = 40. Minus 8 time off on Wednesday = 32 hours
	if workload.TotalAvailableHours != 32.0 {
		t.Fatalf("expected available capacity 32.0, got %v", workload.TotalAvailableHours)
	}
	if workload.TotalTimeOffHours != 8.0 {
		t.Fatalf("expected total time-off 8.0, got %v", workload.TotalTimeOffHours)
	}
	if workload.TotalAssignedHours != 6.0 {
		t.Fatalf("expected total task demand 6.0, got %v", workload.TotalAssignedHours)
	}
	if workload.TotalAllocatedHours != 10.0 {
		t.Fatalf("expected total allocation hours 10.0, got %v", workload.TotalAllocatedHours)
	}
	// Utilization rate: (6 + 10) / 32 * 100 = 50.0%
	if workload.OverallUtilizationRate != 50.0 {
		t.Fatalf("expected utilization rate 50.0%%, got %v%%", workload.OverallUtilizationRate)
	}

	// Inspect daily breakdown: 7 days
	if len(workload.DailyBreakdown) != 7 {
		t.Fatalf("expected 7 daily buckets, got %d", len(workload.DailyBreakdown))
	}
	// Tuesday (2026-11-10): assigned tasks 6.0, allocated 2.0 => total demand 8.0 => 100% (at_capacity)
	tue := workload.DailyBreakdown[1]
	if tue.Date != "2026-11-10" || tue.AssignedTaskHours != 6.0 || tue.AllocatedProjectHours != 2.0 || tue.UtilizationRate != 100.0 {
		t.Fatalf("unexpected Tuesday breakdown bucket: %+v", tue)
	}

	// 6. Test Team Heatmap
	heatmap, err := svc.GetTeamHeatmap(1, from, to)
	if err != nil {
		t.Fatalf("GetTeamHeatmap failed: %v", err)
	}
	if len(heatmap.Users) != 1 || heatmap.WorkspaceID != 1 {
		t.Fatalf("unexpected heatmap: %+v", heatmap)
	}

	// 7. Add Overallocated User (user_2) with 8 hrs capacity, but 60 hrs allocated across 5 days (12 hrs/day)
	_, err = svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID:         1,
		UserID:              "user_2",
		DailyCapacityHours:  floatPtr(8.0),
		WeeklyCapacityHours: floatPtr(40.0),
	})
	if err != nil {
		t.Fatalf("failed to create user_2 profile: %v", err)
	}

	_, err = svc.CreateAllocation(model.CreateResourceAllocationRequest{
		WorkspaceID:           1,
		ProjectID:             101,
		UserID:                "user_2",
		StartDate:             from,
		EndDate:               fri,
		AllocatedHoursPerWeek: 60.0,
	})
	if err != nil {
		t.Fatalf("failed to create user_2 allocation: %v", err)
	}

	// Test Overallocated Alerts
	alerts, err := svc.GetOverallocatedAlerts(1, from, to)
	if err != nil {
		t.Fatalf("GetOverallocatedAlerts failed: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatalf("expected overallocated alerts for user_2, got 0")
	}
	if alerts[0].UserID != "user_2" || alerts[0].OverHours <= 0 {
		t.Fatalf("unexpected alert: %+v", alerts[0])
	}
}

func TestWorkloadService_WorkspaceMetrics(t *testing.T) {
	svc, _, _, _, _ := setupTestWorkloadSvc(t)

	// Validation: Workspace required
	_, err := svc.GetWorkspaceMetrics(0)
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	// Empty workspace metrics
	metrics, err := svc.GetWorkspaceMetrics(1)
	if err != nil {
		t.Fatalf("GetWorkspaceMetrics failed on empty workspace: %v", err)
	}
	if metrics.TotalProfiles != 0 {
		t.Fatalf("expected 0 profiles, got %d", metrics.TotalProfiles)
	}

	// Seed profiles: user_1 (40 hrs), user_2 (30 hrs)
	_, _ = svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID:         1,
		UserID:              "user_1",
		DailyCapacityHours:  floatPtr(8.0),
		WeeklyCapacityHours: floatPtr(40.0),
	})
	_, _ = svc.CreateProfile(model.CreateResourceProfileRequest{
		WorkspaceID:         1,
		UserID:              "user_2",
		DailyCapacityHours:  floatPtr(6.0),
		WeeklyCapacityHours: floatPtr(30.0),
	})

	// Add allocation for user_1
	now := time.Now().Truncate(24 * time.Hour)
	_, _ = svc.CreateAllocation(model.CreateResourceAllocationRequest{
		WorkspaceID:           1,
		ProjectID:             101,
		UserID:                "user_1",
		StartDate:             now,
		EndDate:               now.Add(7 * 24 * time.Hour),
		AllocatedHoursPerWeek: 20.0,
	})

	metrics, err = svc.GetWorkspaceMetrics(1)
	if err != nil {
		t.Fatalf("GetWorkspaceMetrics failed: %v", err)
	}
	if metrics.TotalProfiles != 2 || metrics.ActiveProfiles != 2 {
		t.Fatalf("expected 2 active profiles, got total=%d, active=%d", metrics.TotalProfiles, metrics.ActiveProfiles)
	}
	if metrics.TotalWeeklyCapacityHours != 70.0 {
		t.Fatalf("expected 70.0 total weekly capacity hours, got %v", metrics.TotalWeeklyCapacityHours)
	}
	if metrics.TotalAllocatedProjectHours != 20.0 {
		t.Fatalf("expected 20.0 total allocated hours, got %v", metrics.TotalAllocatedProjectHours)
	}
}
