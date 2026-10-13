package service

import (
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestCommentSvc(t *testing.T) (CommentService, TaskDBService) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	taskRepo := repository.NewTaskRepository(sqliteDB)
	commentRepo := repository.NewCommentRepository(sqliteDB)

	commentSvc := NewCommentService(commentRepo, taskRepo)
	taskSvc := NewTaskDBService(taskRepo)

	return commentSvc, taskSvc
}

func TestCommentService_CRUD(t *testing.T) {
	commentSvc, taskSvc := setupTestCommentSvc(t)
	userID := "user-123"

	// 1. Create parent task
	task, err := taskSvc.CreateTask(userID, "Discuss Architecture", "Details", time.Now().Add(24*time.Hour), 45, "high")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// 2. Validation error on empty content
	_, err = commentSvc.AddComment(task.ID, userID, "   ")
	if err != ErrCommentContentRequired {
		t.Fatalf("expected ErrCommentContentRequired, got %v", err)
	}

	// 3. Unauthorized access to comment on another user's task
	_, err = commentSvc.AddComment(task.ID, "user-999", "Hello")
	if err != ErrUnauthorizedTask {
		t.Fatalf("expected ErrUnauthorizedTask, got %v", err)
	}

	// 4. Create comment
	c1, err := commentSvc.AddComment(task.ID, userID, "Initial discussion note")
	if err != nil {
		t.Fatalf("failed to add comment: %v", err)
	}
	if c1.ID == 0 || c1.Content != "Initial discussion note" {
		t.Errorf("unexpected comment: %#v", c1)
	}

	// 5. List comments
	list, err := commentSvc.GetTaskComments(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to list comments: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(list))
	}

	// 6. Update comment
	updated, err := commentSvc.UpdateComment(c1.ID, userID, "Revised discussion note")
	if err != nil {
		t.Fatalf("failed to update comment: %v", err)
	}
	if updated.Content != "Revised discussion note" {
		t.Errorf("expected updated content, got %s", updated.Content)
	}

	// 7. Unauthorized comment update
	_, err = commentSvc.UpdateComment(c1.ID, "user-999", "Hijacked content")
	if err != ErrUnauthorizedComment {
		t.Fatalf("expected ErrUnauthorizedComment, got %v", err)
	}

	// 8. Delete comment
	err = commentSvc.DeleteComment(c1.ID, userID)
	if err != nil {
		t.Fatalf("failed to delete comment: %v", err)
	}

	// 9. Verify deletion
	listAfter, err := commentSvc.GetTaskComments(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to list comments after deletion: %v", err)
	}
	if len(listAfter) != 0 {
		t.Fatalf("expected 0 comments after deletion, got %d", len(listAfter))
	}
}
