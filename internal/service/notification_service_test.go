package service

import (
	"errors"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestNotificationSvc(t *testing.T) NotificationService {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	repo := repository.NewNotificationRepository(sqliteDB)
	return NewNotificationService(repo)
}

func TestNotificationService_SendAndPreferences(t *testing.T) {
	svc := setupTestNotificationSvc(t)
	user := "user-alice"

	// 1. Validation errors
	_, err := svc.Send(nil)
	if !errors.Is(err, ErrNotificationTitleEmpty) {
		t.Fatalf("expected ErrNotificationTitleEmpty, got %v", err)
	}

	_, err = svc.Send(&model.CreateNotificationRequest{
		UserID: "",
		Title:  "Hello",
		Type:   model.NotificationTypeSystemAlert,
	})
	if !errors.Is(err, ErrNotificationUserIDEmpty) {
		t.Fatalf("expected ErrNotificationUserIDEmpty, got %v", err)
	}

	_, err = svc.Send(&model.CreateNotificationRequest{
		UserID: user,
		Title:  "   ",
		Type:   model.NotificationTypeSystemAlert,
	})
	if !errors.Is(err, ErrNotificationTitleEmpty) {
		t.Fatalf("expected ErrNotificationTitleEmpty, got %v", err)
	}

	_, err = svc.Send(&model.CreateNotificationRequest{
		UserID: user,
		Title:  "Invalid",
		Type:   "unknown_type",
	})
	if !errors.Is(err, ErrInvalidNotificationType) {
		t.Fatalf("expected ErrInvalidNotificationType, got %v", err)
	}

	// 2. Successful Send
	n, err := svc.Send(&model.CreateNotificationRequest{
		UserID:     user,
		Type:       model.NotificationTypeTaskAssigned,
		Title:      "You have been assigned to task #42",
		Message:    "Please review requirements",
		EntityType: "task",
		EntityID:   "42",
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if n == nil || n.ID == 0 || n.IsRead {
		t.Fatalf("unexpected notification: %+v", n)
	}

	// 3. User disables task assignment notifications
	f := false
	_, err = svc.UpdatePreferences(user, &model.UpdatePreferencesRequest{
		NotifyTaskAssigned: &f,
	})
	if err != nil {
		t.Fatalf("update preferences failed: %v", err)
	}

	// Should now be suppressed
	suppressed, err := svc.Send(&model.CreateNotificationRequest{
		UserID: user,
		Type:   model.NotificationTypeTaskAssigned,
		Title:  "Another task assigned",
	})
	if err != nil {
		t.Fatalf("expected no error on suppression, got %v", err)
	}
	if suppressed != nil {
		t.Fatalf("expected notification to be suppressed by preferences, got %+v", suppressed)
	}
}

func TestNotificationService_ReadTrackingAndDelete(t *testing.T) {
	svc := setupTestNotificationSvc(t)
	user := "user-bob"
	otherUser := "user-charlie"

	n1, err := svc.Send(&model.CreateNotificationRequest{
		UserID: user,
		Type:   model.NotificationTypeSystemAlert,
		Title:  "Welcome to Task Manager",
	})
	if err != nil {
		t.Fatalf("send n1 failed: %v", err)
	}

	n2, err := svc.Send(&model.CreateNotificationRequest{
		UserID: user,
		Type:   model.NotificationTypeWorkspaceInvite,
		Title:  "Invited to Platform Team",
	})
	if err != nil {
		t.Fatalf("send n2 failed: %v", err)
	}

	// Unread count should be 2
	count, err := svc.GetUnreadCount(user)
	if err != nil || count != 2 {
		t.Fatalf("expected unread count 2, got %d, err: %v", count, err)
	}

	// Mark n1 as read
	if err := svc.MarkAsRead(n1.ID, user); err != nil {
		t.Fatalf("mark as read failed: %v", err)
	}

	// Unread count should now be 1
	count, err = svc.GetUnreadCount(user)
	if err != nil || count != 1 {
		t.Fatalf("expected unread count 1, got %d, err: %v", count, err)
	}

	// Unauthorized user cannot delete or mark n2
	err = svc.MarkAsRead(n2.ID, otherUser)
	if !errors.Is(err, ErrNotificationNotFound) {
		t.Fatalf("expected ErrNotificationNotFound for other user, got %v", err)
	}

	err = svc.DeleteNotification(n2.ID, otherUser)
	if !errors.Is(err, ErrNotificationNotFound) {
		t.Fatalf("expected ErrNotificationNotFound for other user on delete, got %v", err)
	}

	// Delete n2 by owner
	if err := svc.DeleteNotification(n2.ID, user); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Unread count should be 0
	count, err = svc.GetUnreadCount(user)
	if err != nil || count != 0 {
		t.Fatalf("expected unread count 0, got %d, err: %v", count, err)
	}
}

func TestNotificationService_MarkAllAsReadAndListFilter(t *testing.T) {
	svc := setupTestNotificationSvc(t)
	user := "user-dan"

	// Create 3 notifications
	for i := 1; i <= 3; i++ {
		_, err := svc.Send(&model.CreateNotificationRequest{
			UserID: user,
			Type:   model.NotificationTypeTaskDueSoon,
			Title:  "Task is due soon",
		})
		if err != nil {
			t.Fatalf("send failed: %v", err)
		}
	}

	// List all
	summary, err := svc.ListNotifications(model.NotificationFilter{
		UserID: user,
	})
	if err != nil || summary.Total != 3 || summary.UnreadCount != 3 {
		t.Fatalf("expected 3 total and 3 unread, got %d and %d", summary.Total, summary.UnreadCount)
	}

	// Mark all as read
	if err := svc.MarkAllAsRead(user); err != nil {
		t.Fatalf("mark all as read failed: %v", err)
	}

	// Verify unread count is 0
	count, err := svc.GetUnreadCount(user)
	if err != nil || count != 0 {
		t.Fatalf("expected unread count 0, got %d", count)
	}

	// Filter unread only should return 0 items
	summaryUnread, err := svc.ListNotifications(model.NotificationFilter{
		UserID:     user,
		UnreadOnly: true,
	})
	if err != nil || len(summaryUnread.Notifications) != 0 {
		t.Fatalf("expected 0 unread notifications in list, got %d", len(summaryUnread.Notifications))
	}

	// Preferences check and toggle all
	tTrue := true
	tFalse := false
	updatedPref, err := svc.UpdatePreferences(user, &model.UpdatePreferencesRequest{
		EmailEnabled:       &tFalse,
		InAppEnabled:       &tTrue,
		NotifyTaskAssigned: &tFalse,
		NotifyTaskDueSoon:  &tTrue,
		NotifyComment:      &tFalse,
		NotifyInvite:       &tTrue,
	})
	if err != nil {
		t.Fatalf("update preferences failed: %v", err)
	}
	if updatedPref.EmailEnabled || !updatedPref.InAppEnabled || updatedPref.NotifyTaskAssigned || !updatedPref.NotifyTaskDueSoon {
		t.Fatalf("preferences not saved correctly: %+v", updatedPref)
	}

	fetchedPref, err := svc.GetPreferences(user)
	if err != nil {
		t.Fatalf("get preferences failed: %v", err)
	}
	if fetchedPref.EmailEnabled != false || fetchedPref.InAppEnabled != true {
		t.Fatalf("mismatched fetched preferences: %+v", fetchedPref)
	}
}

