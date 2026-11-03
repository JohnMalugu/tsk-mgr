package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestDocSvc(t *testing.T) (DocumentService, repository.DocumentRepository, repository.ProjectRepository, repository.TaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-doc', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Product WS', 'prod-ws-doc', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Core API', 'Core Backend Service', '#4f46e5', datetime('now'), datetime('now')),
		       (102, 'user_1', 'Web Client', 'React Web Frontend', '#10b981', datetime('now'), datetime('now'));

		INSERT INTO tasks (id, user_id, project_id, title, description, status, completed, due_date, estimate_minutes, priority, created_at, updated_at)
		VALUES (501, 'user_1', 101, 'Implement Auth V2', 'Task auth spec', 'todo', 0, datetime('now'), 60, 'high', datetime('now'), datetime('now')),
		       (502, 'user_1', 101, 'Database Migration', 'Task db migration', 'in_progress', 0, datetime('now'), 30, 'medium', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	docRepo := repository.NewDocumentRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	svc := NewDocumentService(docRepo, projectRepo, taskRepo)
	return svc, docRepo, projectRepo, taskRepo, sqliteDB
}

func TestDocumentService_DocumentCRUD(t *testing.T) {
	svc, _, _, _, _ := setupTestDocSvc(t)

	// 1. Validation: Workspace required
	_, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 0,
		Title:       "Architecture Spec",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	// 2. Validation: Title required
	_, err = svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "   ",
	})
	if !errors.Is(err, ErrDocumentTitleRequired) {
		t.Fatalf("expected ErrDocumentTitleRequired, got %v", err)
	}

	// 3. Validation: Invalid doc type
	_, err = svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Spec",
		DocType:     "unknown_type",
	})
	if !errors.Is(err, ErrInvalidDocType) {
		t.Fatalf("expected ErrInvalidDocType, got %v", err)
	}

	// 4. Validation: Invalid status
	_, err = svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Spec",
		Status:      "invalid_status",
	})
	if !errors.Is(err, ErrInvalidPublishStatus) {
		t.Fatalf("expected ErrInvalidPublishStatus, got %v", err)
	}

	// 5. Validation: Project not found
	invalidProj := 9999
	_, err = svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		ProjectID:   &invalidProj,
		Title:       "Spec",
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}

	// 6. Validation: Parent document not found
	invalidParent := int64(9999)
	_, err = svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID:      1,
		ParentDocumentID: &invalidParent,
		Title:            "Sub Spec",
	})
	if !errors.Is(err, ErrParentDocumentNotFound) {
		t.Fatalf("expected ErrParentDocumentNotFound, got %v", err)
	}

	// 7. Successful creation & slug generation
	projID := 101
	doc1, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		ProjectID:   &projID,
		Title:       "Architecture Overview RFC 1",
		Content:     "# Architecture\n\nInitial overview.",
		DocType:     model.DocTypeRFC,
		Status:      model.DocStatusDraft,
	})
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	if doc1.ID <= 0 {
		t.Fatalf("expected positive doc ID, got %d", doc1.ID)
	}
	if doc1.Slug != "architecture-overview-rfc-1" {
		t.Fatalf("expected slug 'architecture-overview-rfc-1', got '%s'", doc1.Slug)
	}
	if doc1.Version != 1 {
		t.Fatalf("expected initial version 1, got %d", doc1.Version)
	}
	if doc1.DocType != model.DocTypeRFC {
		t.Fatalf("expected doc type rfc, got %s", doc1.DocType)
	}

	// 8. Handle slug collision by appending suffix
	doc2, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Architecture Overview RFC 1",
		Content:     "Second copy.",
	})
	if err != nil {
		t.Fatalf("failed to create second document: %v", err)
	}
	if doc2.Slug == doc1.Slug {
		t.Fatalf("expected unique slug for second document, got identical '%s'", doc2.Slug)
	}

	// 9. GetDocument with view recording
	fetched, err := svc.GetDocument(doc1.ID, true)
	if err != nil {
		t.Fatalf("failed to get document: %v", err)
	}
	if fetched.ViewCount != 1 {
		t.Fatalf("expected ViewCount 1, got %d", fetched.ViewCount)
	}

	// GetDocument without view recording
	fetched2, err := svc.GetDocument(doc1.ID, false)
	if err != nil {
		t.Fatalf("failed to get document: %v", err)
	}
	if fetched2.ViewCount != 1 {
		t.Fatalf("expected ViewCount unchanged at 1, got %d", fetched2.ViewCount)
	}

	// 10. GetDocumentBySlug
	bySlug, err := svc.GetDocumentBySlug(1, doc1.Slug, false)
	if err != nil {
		t.Fatalf("failed to get document by slug: %v", err)
	}
	if bySlug.ID != doc1.ID {
		t.Fatalf("expected document ID %d, got %d", doc1.ID, bySlug.ID)
	}

	// 11. UpdateDocument: title and content change creates point-in-time revision
	newTitle := "Architecture Overview RFC 1 (Updated)"
	newContent := "# Architecture\n\nUpdated overview with system diagrams."
	updated, err := svc.UpdateDocument(doc1.ID, "user_2", model.UpdateDocumentRequest{
		Title:         &newTitle,
		Content:       &newContent,
		ChangeSummary: "Added diagram section",
	})
	if err != nil {
		t.Fatalf("failed to update document: %v", err)
	}
	if updated.Title != newTitle {
		t.Fatalf("expected updated title '%s', got '%s'", newTitle, updated.Title)
	}
	if updated.Version != 2 {
		t.Fatalf("expected version incremented to 2, got %d", updated.Version)
	}

	// Verify revision snapshot was created for v1 and v2
	revs, err := svc.ListRevisions(doc1.ID)
	if err != nil {
		t.Fatalf("failed to list revisions: %v", err)
	}
	if len(revs) != 2 {
		t.Fatalf("expected 2 revisions (initial v1 + updated v2), got %d", len(revs))
	}

	// 12. Hierarchy & Circular Prevention
	// Make doc2 a child of doc1
	doc1ID := doc1.ID
	_, err = svc.UpdateDocument(doc2.ID, "user_1", model.UpdateDocumentRequest{
		ParentDocumentID: &doc1ID,
	})
	if err != nil {
		t.Fatalf("failed to set parent of doc2 to doc1: %v", err)
	}

	// Attempting to make doc1 a child of doc2 should trigger circular hierarchy detection
	doc2ID := doc2.ID
	_, err = svc.UpdateDocument(doc1.ID, "user_1", model.UpdateDocumentRequest{
		ParentDocumentID: &doc2ID,
	})
	if !errors.Is(err, ErrCircularDocumentHierarchy) {
		t.Fatalf("expected ErrCircularDocumentHierarchy, got %v", err)
	}

	// Attempting to make doc1 a child of itself should trigger circular hierarchy detection
	_, err = svc.UpdateDocument(doc1.ID, "user_1", model.UpdateDocumentRequest{
		ParentDocumentID: &doc1ID,
	})
	if !errors.Is(err, ErrCircularDocumentHierarchy) {
		t.Fatalf("expected ErrCircularDocumentHierarchy, got %v", err)
	}

	// 13. ListDocuments with filters
	list, err := svc.ListDocuments(model.DocumentFilter{
		WorkspaceID: 1,
		DocType:     model.DocTypeRFC,
	})
	if err != nil {
		t.Fatalf("failed to list documents: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 RFC document, got %d", len(list))
	}

	// 14. DeleteDocument
	err = svc.DeleteDocument(doc2.ID)
	if err != nil {
		t.Fatalf("failed to delete document: %v", err)
	}
	_, err = svc.GetDocument(doc2.ID, false)
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Fatalf("expected ErrDocumentNotFound after deletion, got %v", err)
	}
}

func TestDocumentService_DocumentTree(t *testing.T) {
	svc, _, _, _, _ := setupTestDocSvc(t)

	// Create root doc 1
	root1, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Engineering Handbook",
		Content:     "Root handbook",
	})
	if err != nil {
		t.Fatalf("failed to create root1: %v", err)
	}

	// Create child of root 1
	root1ID := root1.ID
	child1, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID:      1,
		ParentDocumentID: &root1ID,
		Title:            "Backend Guidelines",
		Content:          "Child guidelines",
	})
	if err != nil {
		t.Fatalf("failed to create child1: %v", err)
	}

	// Create grandchild
	child1ID := child1.ID
	grandchild1, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID:      1,
		ParentDocumentID: &child1ID,
		Title:            "Go Conventions",
		Content:          "Grandchild conventions",
	})
	if err != nil {
		t.Fatalf("failed to create grandchild1: %v", err)
	}

	// Create root doc 2
	root2, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Runbooks",
		Content:     "Second root runbooks",
	})
	if err != nil {
		t.Fatalf("failed to create root2: %v", err)
	}

	// Fetch document tree
	tree, err := svc.GetDocumentTree(1, nil)
	if err != nil {
		t.Fatalf("failed to get document tree: %v", err)
	}

	// Expect 2 root nodes
	if len(tree) != 2 {
		t.Fatalf("expected 2 root nodes in tree, got %d", len(tree))
	}

	// Find Engineering Handbook node
	var handbookNode *model.DocumentTreeNode
	for i := range tree {
		if tree[i].ID == root1.ID {
			handbookNode = &tree[i]
			break
		}
	}
	if handbookNode == nil {
		t.Fatalf("could not find handbook node in tree")
	}

	// Verify child nesting
	if len(handbookNode.Children) != 1 {
		t.Fatalf("expected 1 child under handbook, got %d", len(handbookNode.Children))
	}
	backendNode := handbookNode.Children[0]
	if backendNode.ID != child1.ID {
		t.Fatalf("expected child1 id %d, got %d", child1.ID, backendNode.ID)
	}

	// Verify grandchild nesting
	if len(backendNode.Children) != 1 {
		t.Fatalf("expected 1 grandchild under backend node, got %d", len(backendNode.Children))
	}
	if backendNode.Children[0].ID != grandchild1.ID {
		t.Fatalf("expected grandchild1 id %d, got %d", grandchild1.ID, backendNode.Children[0].ID)
	}

	// Verify second root node
	var runbookNode *model.DocumentTreeNode
	for i := range tree {
		if tree[i].ID == root2.ID {
			runbookNode = &tree[i]
			break
		}
	}
	if runbookNode == nil {
		t.Fatalf("could not find runbook node in tree")
	}
	if len(runbookNode.Children) != 0 {
		t.Fatalf("expected 0 children under runbook, got %d", len(runbookNode.Children))
	}
}

func TestDocumentService_PublishAndPin(t *testing.T) {
	svc, _, _, _, _ := setupTestDocSvc(t)

	doc, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Deployment Checklist",
		Content:     "Checklist steps...",
		Status:      model.DocStatusDraft,
	})
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Transition draft -> in_review
	pub1, err := svc.PublishDocument(doc.ID, model.DocStatusInReview)
	if err != nil {
		t.Fatalf("failed to publish to in_review: %v", err)
	}
	if pub1.Status != model.DocStatusInReview {
		t.Fatalf("expected status in_review, got %s", pub1.Status)
	}

	// Transition in_review -> published
	pub2, err := svc.PublishDocument(doc.ID, model.DocStatusPublished)
	if err != nil {
		t.Fatalf("failed to publish to published: %v", err)
	}
	if pub2.Status != model.DocStatusPublished {
		t.Fatalf("expected status published, got %s", pub2.Status)
	}

	// Invalid status rejected
	_, err = svc.PublishDocument(doc.ID, "not_a_status")
	if !errors.Is(err, ErrInvalidPublishStatus) {
		t.Fatalf("expected ErrInvalidPublishStatus, got %v", err)
	}

	// TogglePin
	if doc.IsPinned {
		t.Fatalf("expected initial is_pinned to be false")
	}
	pinned, err := svc.TogglePin(doc.ID)
	if err != nil {
		t.Fatalf("failed to pin document: %v", err)
	}
	if !pinned.IsPinned {
		t.Fatalf("expected is_pinned to be true")
	}

	unpinned, err := svc.TogglePin(doc.ID)
	if err != nil {
		t.Fatalf("failed to unpin document: %v", err)
	}
	if unpinned.IsPinned {
		t.Fatalf("expected is_pinned to be false")
	}
}

func TestDocumentService_RevisionsAndRollback(t *testing.T) {
	svc, _, _, _, _ := setupTestDocSvc(t)

	doc, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Original Spec V1",
		Content:     "Content of version 1.",
	})
	if err != nil {
		t.Fatalf("failed to create doc: %v", err)
	}

	// Update to v2
	t2 := "Spec V2"
	c2 := "Content of version 2."
	_, err = svc.UpdateDocument(doc.ID, "user_1", model.UpdateDocumentRequest{
		Title:         &t2,
		Content:       &c2,
		ChangeSummary: "v2 changes",
	})
	if err != nil {
		t.Fatalf("failed to update to v2: %v", err)
	}

	// Update to v3
	t3 := "Spec V3"
	c3 := "Content of version 3."
	_, err = svc.UpdateDocument(doc.ID, "user_2", model.UpdateDocumentRequest{
		Title:         &t3,
		Content:       &c3,
		ChangeSummary: "v3 changes",
	})
	if err != nil {
		t.Fatalf("failed to update to v3: %v", err)
	}

	// List revisions
	revs, err := svc.ListRevisions(doc.ID)
	if err != nil {
		t.Fatalf("failed to list revisions: %v", err)
	}
	if len(revs) != 3 {
		t.Fatalf("expected 3 revisions, got %d", len(revs))
	}

	// Get revision 1
	rev1, err := svc.GetRevision(doc.ID, 1)
	if err != nil {
		t.Fatalf("failed to get revision 1: %v", err)
	}
	if rev1.Title != "Original Spec V1" {
		t.Fatalf("expected rev1 title 'Original Spec V1', got '%s'", rev1.Title)
	}
	if rev1.Content != "Content of version 1." {
		t.Fatalf("expected rev1 content 'Content of version 1.', got '%s'", rev1.Content)
	}

	// Non-existent revision
	_, err = svc.GetRevision(doc.ID, 999)
	if !errors.Is(err, ErrRevisionNotFound) {
		t.Fatalf("expected ErrRevisionNotFound, got %v", err)
	}

	// Rollback to version 1
	rolledBack, err := svc.RollbackDocument(doc.ID, "user_1", model.RollbackRevisionRequest{
		TargetVersion: 1,
		Reason:        "Reverting breaking changes from v2 and v3",
	})
	if err != nil {
		t.Fatalf("failed to rollback document: %v", err)
	}
	if rolledBack.Title != "Original Spec V1" {
		t.Fatalf("expected restored title 'Original Spec V1', got '%s'", rolledBack.Title)
	}
	if rolledBack.Content != "Content of version 1." {
		t.Fatalf("expected restored content 'Content of version 1.', got '%s'", rolledBack.Content)
	}
	if rolledBack.Version != 4 {
		t.Fatalf("expected rollback to produce version 4, got %d", rolledBack.Version)
	}
}

func TestDocumentService_TaskLinks(t *testing.T) {
	svc, _, _, _, _ := setupTestDocSvc(t)

	doc, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Auth Architecture",
		Content:     "Auth spec details...",
	})
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Invalid relationship type
	_, err = svc.LinkTask(doc.ID, "user_1", model.LinkDocumentTaskRequest{
		TaskID:       501,
		Relationship: "invalid_rel",
	})
	if !errors.Is(err, ErrInvalidTaskRelationship) {
		t.Fatalf("expected ErrInvalidTaskRelationship, got %v", err)
	}

	// Non-existent task
	_, err = svc.LinkTask(doc.ID, "user_1", model.LinkDocumentTaskRequest{
		TaskID:       9999,
		Relationship: model.DocTaskRelSpecification,
	})
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}

	// Link task 501
	link, err := svc.LinkTask(doc.ID, "user_1", model.LinkDocumentTaskRequest{
		TaskID:       501,
		Relationship: model.DocTaskRelSpecification,
	})
	if err != nil {
		t.Fatalf("failed to link task: %v", err)
	}
	if link.TaskID != 501 || link.Relationship != model.DocTaskRelSpecification {
		t.Fatalf("unexpected link result: %+v", link)
	}

	// Duplicate link updates relationship type
	linkUpdated, err := svc.LinkTask(doc.ID, "user_1", model.LinkDocumentTaskRequest{
		TaskID:       501,
		Relationship: model.DocTaskRelRelated,
	})
	if err != nil {
		t.Fatalf("failed to update task link: %v", err)
	}
	if linkUpdated.Relationship != model.DocTaskRelRelated {
		t.Fatalf("expected updated relationship related, got %s", linkUpdated.Relationship)
	}

	// List linked tasks with enriched title and status
	tasks, err := svc.ListDocumentTasks(doc.ID)
	if err != nil {
		t.Fatalf("failed to list document tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task link, got %d", len(tasks))
	}
	if tasks[0].TaskTitle != "Implement Auth V2" {
		t.Fatalf("expected enriched task title 'Implement Auth V2', got '%s'", tasks[0].TaskTitle)
	}
	if tasks[0].TaskStatus != "todo" {
		t.Fatalf("expected enriched task status 'todo', got '%s'", tasks[0].TaskStatus)
	}
	if tasks[0].TaskPriority != "high" {
		t.Fatalf("expected enriched task priority 'high', got '%s'", tasks[0].TaskPriority)
	}

	// Unlink task
	err = svc.UnlinkTask(doc.ID, 501)
	if err != nil {
		t.Fatalf("failed to unlink task: %v", err)
	}
	tasksAfter, err := svc.ListDocumentTasks(doc.ID)
	if err != nil {
		t.Fatalf("failed to list document tasks: %v", err)
	}
	if len(tasksAfter) != 0 {
		t.Fatalf("expected 0 task links after unlink, got %d", len(tasksAfter))
	}
}

func TestDocumentService_Comments(t *testing.T) {
	svc, _, _, _, _ := setupTestDocSvc(t)

	doc, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "API Design Spec",
		Content:     "Endpoints specification...",
	})
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Empty comment content
	_, err = svc.CreateComment(doc.ID, "user_2", model.CreateDocumentCommentRequest{
		Content: "   ",
	})
	if !errors.Is(err, ErrCommentContentRequired) {
		t.Fatalf("expected ErrCommentContentRequired, got %v", err)
	}

	// Valid comment
	c1, err := svc.CreateComment(doc.ID, "user_2", model.CreateDocumentCommentRequest{
		Content: "Should we add rate limiting headers here?",
	})
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}
	if c1.ID <= 0 {
		t.Fatalf("expected positive comment ID, got %d", c1.ID)
	}
	if c1.IsResolved {
		t.Fatalf("expected newly created comment to be unresolved")
	}

	// List comments
	comments, err := svc.ListComments(doc.ID)
	if err != nil {
		t.Fatalf("failed to list comments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	// Resolve comment
	resolved, err := svc.ResolveComment(c1.ID, "user_1")
	if err != nil {
		t.Fatalf("failed to resolve comment: %v", err)
	}
	if !resolved.IsResolved {
		t.Fatalf("expected comment to be resolved")
	}
	if resolved.ResolvedBy == nil || *resolved.ResolvedBy != "user_1" {
		t.Fatalf("expected resolved_by user_1, got %v", resolved.ResolvedBy)
	}
	if resolved.ResolvedAt == nil {
		t.Fatalf("expected resolved_at timestamp to be set")
	}

	// Non-existent comment resolve
	_, err = svc.ResolveComment(99999, "user_1")
	if !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("expected ErrCommentNotFound, got %v", err)
	}
}

func TestDocumentService_WorkspaceMetrics(t *testing.T) {
	svc, _, _, _, _ := setupTestDocSvc(t)

	// Create a draft RFC doc
	doc1, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Doc One",
		DocType:     model.DocTypeRFC,
		Status:      model.DocStatusDraft,
	})
	if err != nil {
		t.Fatalf("failed to create doc1: %v", err)
	}

	// Pin doc1
	_, _ = svc.TogglePin(doc1.ID)

	// Create a published Runbook doc
	doc2, err := svc.CreateDocument("user_1", model.CreateDocumentRequest{
		WorkspaceID: 1,
		Title:       "Doc Two",
		DocType:     model.DocTypeRunbook,
		Status:      model.DocStatusPublished,
	})
	if err != nil {
		t.Fatalf("failed to create doc2: %v", err)
	}

	// Add comments to doc2
	_, _ = svc.CreateComment(doc2.ID, "user_2", model.CreateDocumentCommentRequest{
		Content: "Looks good!",
	})

	// Get workspace metrics
	metrics, err := svc.GetDocumentMetrics(1)
	if err != nil {
		t.Fatalf("failed to get document metrics: %v", err)
	}

	if metrics.TotalDocuments != 2 {
		t.Fatalf("expected 2 total documents, got %d", metrics.TotalDocuments)
	}
	if metrics.PinnedDocuments != 1 {
		t.Fatalf("expected 1 pinned document, got %d", metrics.PinnedDocuments)
	}
	if metrics.DraftDocuments != 1 {
		t.Fatalf("expected 1 draft document, got %d", metrics.DraftDocuments)
	}
	if metrics.PublishedDocuments != 1 {
		t.Fatalf("expected 1 published document, got %d", metrics.PublishedDocuments)
	}
	if metrics.TotalComments != 1 {
		t.Fatalf("expected 1 total comment, got %d", metrics.TotalComments)
	}
	if metrics.DocTypeBreakdown[model.DocTypeRFC] != 1 {
		t.Fatalf("expected 1 RFC doc in type breakdown, got %d", metrics.DocTypeBreakdown[model.DocTypeRFC])
	}
	if metrics.DocTypeBreakdown[model.DocTypeRunbook] != 1 {
		t.Fatalf("expected 1 Runbook doc in type breakdown, got %d", metrics.DocTypeBreakdown[model.DocTypeRunbook])
	}
	if len(metrics.RecentlyUpdated) != 2 {
		t.Fatalf("expected 2 recently updated documents, got %d", len(metrics.RecentlyUpdated))
	}
}
