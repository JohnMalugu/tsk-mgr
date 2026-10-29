package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	if err := createSchema(db); err != nil {
		return nil, err
	}

	log.Println("Database connection established and schema initialized")
	return db, nil
}

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	projectQuery := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		color TEXT NOT NULL DEFAULT '#6366f1',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(projectQuery)
	if err != nil {
		return err
	}

	taskQuery := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		project_id INTEGER,
		title TEXT NOT NULL,
		description TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		due_date DATETIME NOT NULL,
		estimate_minutes INTEGER,
		completed BOOLEAN NOT NULL DEFAULT 0,
		status TEXT NOT NULL,
		priority TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
	);`

	_, err = db.Exec(taskQuery)
	if err != nil {
		return err
	}

	// Safely attempt migration for existing databases
	_, _ = db.Exec(`ALTER TABLE tasks ADD COLUMN project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL;`)

	tagQuery := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		color TEXT NOT NULL DEFAULT '#10b981',
		created_at DATETIME NOT NULL,
		UNIQUE(user_id, name),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(tagQuery)
	if err != nil {
		return err
	}

	taskTagQuery := `
	CREATE TABLE IF NOT EXISTS task_tags (
		task_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		PRIMARY KEY (task_id, tag_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(taskTagQuery)
	if err != nil {
		return err
	}

	subtaskQuery := `
	CREATE TABLE IF NOT EXISTS subtasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT 0,
		position INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(subtaskQuery)
	if err != nil {
		return err
	}

	commentQuery := `
	CREATE TABLE IF NOT EXISTS task_comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(commentQuery)
	if err != nil {
		return err
	}

	timeEntriesQuery := `
	CREATE TABLE IF NOT EXISTS time_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		ended_at DATETIME,
		duration_seconds INTEGER NOT NULL DEFAULT 0,
		note TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(timeEntriesQuery)
	if err != nil {
		return err
	}

	taskDependenciesQuery := `
	CREATE TABLE IF NOT EXISTS task_dependencies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		depends_on_task_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		UNIQUE(task_id, depends_on_task_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (depends_on_task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(taskDependenciesQuery)
	if err != nil {
		return err
	}

	activitiesQuery := `
	CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		task_id INTEGER,
		action TEXT NOT NULL,
		summary TEXT NOT NULL,
		details_json TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(activitiesQuery)
	if err != nil {
		return err
	}

	workspacesQuery := `
	CREATE TABLE IF NOT EXISTS workspaces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		owner_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(workspacesQuery)
	if err != nil {
		return err
	}

	workspaceMembersQuery := `
	CREATE TABLE IF NOT EXISTS workspace_members (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'member',
		joined_at DATETIME NOT NULL,
		UNIQUE(workspace_id, user_id),
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(workspaceMembersQuery)
	if err != nil {
		return err
	}

	workspaceInvitesQuery := `
	CREATE TABLE IF NOT EXISTS workspace_invites (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		email TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'member',
		invited_by TEXT NOT NULL,
		token TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL DEFAULT 'pending',
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(workspaceInvitesQuery)
	if err != nil {
		return err
	}

	notificationsQuery := `
	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		actor_id TEXT,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		message TEXT NOT NULL,
		entity_type TEXT,
		entity_id TEXT,
		is_read BOOLEAN NOT NULL DEFAULT 0,
		read_at DATETIME,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL
	);`

	_, err = db.Exec(notificationsQuery)
	if err != nil {
		return err
	}

	notificationPreferencesQuery := `
	CREATE TABLE IF NOT EXISTS notification_preferences (
		user_id TEXT PRIMARY KEY,
		email_enabled BOOLEAN NOT NULL DEFAULT 1,
		in_app_enabled BOOLEAN NOT NULL DEFAULT 1,
		notify_task_assigned BOOLEAN NOT NULL DEFAULT 1,
		notify_task_due_soon BOOLEAN NOT NULL DEFAULT 1,
		notify_comment BOOLEAN NOT NULL DEFAULT 1,
		notify_invite BOOLEAN NOT NULL DEFAULT 1,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(notificationPreferencesQuery)
	if err != nil {
		return err
	}

	webhookSubscriptionsQuery := `
	CREATE TABLE IF NOT EXISTS webhook_subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		target_url TEXT NOT NULL,
		secret_token TEXT NOT NULL,
		event_types TEXT NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		description TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(webhookSubscriptionsQuery)
	if err != nil {
		return err
	}

	webhookDeliveriesQuery := `
	CREATE TABLE IF NOT EXISTS webhook_deliveries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		subscription_id INTEGER NOT NULL,
		event_type TEXT NOT NULL,
		payload_json TEXT NOT NULL,
		status_code INTEGER NOT NULL DEFAULT 0,
		response_headers TEXT NOT NULL DEFAULT '',
		response_body TEXT NOT NULL DEFAULT '',
		duration_ms INTEGER NOT NULL DEFAULT 0,
		attempt_number INTEGER NOT NULL DEFAULT 1,
		status TEXT NOT NULL,
		error_message TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		FOREIGN KEY (subscription_id) REFERENCES webhook_subscriptions(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(webhookDeliveriesQuery)
	if err != nil {
		return err
	}

	sprintsQuery := `
	CREATE TABLE IF NOT EXISTS sprints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		project_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		goal TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'planned',
		start_date DATETIME,
		end_date DATETIME,
		total_points INTEGER NOT NULL DEFAULT 0,
		completed_points INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(sprintsQuery)
	if err != nil {
		return err
	}

	sprintTasksQuery := `
	CREATE TABLE IF NOT EXISTS sprint_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sprint_id INTEGER NOT NULL,
		task_id INTEGER NOT NULL,
		story_points INTEGER NOT NULL DEFAULT 0,
		added_at DATETIME NOT NULL,
		UNIQUE(sprint_id, task_id),
		FOREIGN KEY (sprint_id) REFERENCES sprints(id) ON DELETE CASCADE,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(sprintTasksQuery)
	if err != nil {
		return err
	}

	customFieldsQuery := `
	CREATE TABLE IF NOT EXISTS custom_field_definitions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		project_id INTEGER,
		name TEXT NOT NULL,
		field_key TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL,
		required INTEGER NOT NULL DEFAULT 0,
		options TEXT NOT NULL DEFAULT '[]',
		default_value TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		UNIQUE(workspace_id, field_key),
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(customFieldsQuery)
	if err != nil {
		return err
	}

	customValuesQuery := `
	CREATE TABLE IF NOT EXISTS custom_field_values (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		field_id INTEGER NOT NULL,
		task_id INTEGER NOT NULL,
		value_text TEXT,
		value_number REAL,
		value_boolean INTEGER,
		value_date DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		UNIQUE(field_id, task_id),
		FOREIGN KEY (field_id) REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(customValuesQuery)
	if err != nil {
		return err
	}

	attachmentsQuery := `
	CREATE TABLE IF NOT EXISTS attachments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		workspace_id INTEGER NOT NULL,
		filename TEXT NOT NULL,
		original_name TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		content_type TEXT NOT NULL,
		storage_path TEXT NOT NULL,
		checksum_sha256 TEXT NOT NULL,
		download_count INTEGER NOT NULL DEFAULT 0,
		is_archived INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(attachmentsQuery)
	if err != nil {
		return err
	}

	templatesQuery := `
	CREATE TABLE IF NOT EXISTS task_templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		created_by_user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT 'general',
		default_priority TEXT NOT NULL DEFAULT 'medium',
		estimated_minutes INTEGER NOT NULL DEFAULT 0,
		tags TEXT NOT NULL DEFAULT '[]',
		checklist_items TEXT NOT NULL DEFAULT '[]',
		subtasks TEXT NOT NULL DEFAULT '[]',
		usage_count INTEGER NOT NULL DEFAULT 0,
		is_archived INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		last_used_at DATETIME,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(templatesQuery)
	if err != nil {
		return err
	}

	savedViewsQuery := `
	CREATE TABLE IF NOT EXISTS saved_views (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		icon TEXT NOT NULL DEFAULT 'bookmark',
		color TEXT NOT NULL DEFAULT '#4F46E5',
		filter_criteria TEXT NOT NULL DEFAULT '{}',
		sort_by TEXT NOT NULL DEFAULT 'created_at',
		sort_order TEXT NOT NULL DEFAULT 'desc',
		group_by TEXT NOT NULL DEFAULT 'none',
		is_shared INTEGER NOT NULL DEFAULT 0,
		is_pinned INTEGER NOT NULL DEFAULT 0,
		position INTEGER NOT NULL DEFAULT 0,
		usage_count INTEGER NOT NULL DEFAULT 0,
		last_used_at DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(savedViewsQuery)
	if err != nil {
		return err
	}

	automationRulesQuery := `
	CREATE TABLE IF NOT EXISTS automation_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		created_by_user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		trigger_type TEXT NOT NULL,
		conditions_json TEXT NOT NULL DEFAULT '[]',
		actions_json TEXT NOT NULL DEFAULT '[]',
		is_active INTEGER NOT NULL DEFAULT 1,
		execution_count INTEGER NOT NULL DEFAULT 0,
		last_triggered_at DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(automationRulesQuery)
	if err != nil {
		return err
	}

	automationExecutionLogsQuery := `
	CREATE TABLE IF NOT EXISTS automation_execution_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		rule_id INTEGER NOT NULL,
		workspace_id INTEGER NOT NULL,
		trigger_event TEXT NOT NULL,
		status TEXT NOT NULL,
		actions_executed TEXT NOT NULL DEFAULT '[]',
		error_message TEXT NOT NULL DEFAULT '',
		duration_ms INTEGER NOT NULL DEFAULT 0,
		executed_at DATETIME NOT NULL,
		FOREIGN KEY (rule_id) REFERENCES automation_rules(id) ON DELETE CASCADE,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(automationExecutionLogsQuery)
	if err != nil {
		return err
	}

	releasesQuery := `
	CREATE TABLE IF NOT EXISTS releases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		project_id INTEGER,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'planning',
		target_date DATETIME,
		released_at DATETIME,
		created_by_user_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(releasesQuery)
	if err != nil {
		return err
	}

	releaseItemsQuery := `
	CREATE TABLE IF NOT EXISTS release_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		task_id INTEGER NOT NULL,
		category TEXT NOT NULL DEFAULT 'feature',
		include_in_changelog INTEGER NOT NULL DEFAULT 1,
		custom_notes TEXT NOT NULL DEFAULT '',
		added_by_user_id TEXT NOT NULL,
		added_at DATETIME NOT NULL,
		UNIQUE(release_id, task_id),
		FOREIGN KEY (release_id) REFERENCES releases(id) ON DELETE CASCADE,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (added_by_user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(releaseItemsQuery)
	if err != nil {
		return err
	}

	intakeFormsQuery := `
	CREATE TABLE IF NOT EXISTS intake_forms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		project_id INTEGER,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		slug TEXT UNIQUE NOT NULL,
		status TEXT NOT NULL DEFAULT 'draft',
		is_published INTEGER NOT NULL DEFAULT 0,
		allow_anonymous INTEGER NOT NULL DEFAULT 1,
		submit_button_text TEXT NOT NULL DEFAULT 'Submit',
		success_message TEXT NOT NULL DEFAULT 'Thank you! Your submission has been received.',
		redirect_url TEXT NOT NULL DEFAULT '',
		default_priority TEXT NOT NULL DEFAULT 'medium',
		default_tags TEXT NOT NULL DEFAULT '[]',
		default_assignee_id TEXT,
		submission_count INTEGER NOT NULL DEFAULT 0,
		created_by_user_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
		FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (default_assignee_id) REFERENCES users(id) ON DELETE SET NULL
	);`

	_, err = db.Exec(intakeFormsQuery)
	if err != nil {
		return err
	}

	intakeFormFieldsQuery := `
	CREATE TABLE IF NOT EXISTS intake_form_fields (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		form_id INTEGER NOT NULL,
		field_key TEXT NOT NULL,
		field_type TEXT NOT NULL,
		target_field TEXT NOT NULL,
		label TEXT NOT NULL,
		placeholder TEXT NOT NULL DEFAULT '',
		help_text TEXT NOT NULL DEFAULT '',
		is_required INTEGER NOT NULL DEFAULT 0,
		position INTEGER NOT NULL DEFAULT 0,
		options TEXT NOT NULL DEFAULT '[]',
		default_value TEXT NOT NULL DEFAULT '',
		UNIQUE(form_id, field_key),
		FOREIGN KEY (form_id) REFERENCES intake_forms(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(intakeFormFieldsQuery)
	if err != nil {
		return err
	}

	intakeFormSubmissionsQuery := `
	CREATE TABLE IF NOT EXISTS intake_form_submissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		form_id INTEGER NOT NULL,
		task_id INTEGER,
		submitter_email TEXT NOT NULL DEFAULT '',
		submitter_ip TEXT NOT NULL DEFAULT '',
		user_agent TEXT NOT NULL DEFAULT '',
		responses TEXT NOT NULL DEFAULT '{}',
		submitted_at DATETIME NOT NULL,
		FOREIGN KEY (form_id) REFERENCES intake_forms(id) ON DELETE CASCADE,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL
	);`

	_, err = db.Exec(intakeFormSubmissionsQuery)
	if err != nil {
		return err
	}

	slaPoliciesQuery := `
	CREATE TABLE IF NOT EXISTS sla_policies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		priority_tiers TEXT NOT NULL DEFAULT '[]',
		target_response_minutes INTEGER NOT NULL DEFAULT 60,
		target_resolution_minutes INTEGER NOT NULL DEFAULT 1440,
		business_hours_only INTEGER NOT NULL DEFAULT 0,
		is_default INTEGER NOT NULL DEFAULT 0,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_by_user_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(slaPoliciesQuery)
	if err != nil {
		return err
	}

	slaEscalationRulesQuery := `
	CREATE TABLE IF NOT EXISTS sla_escalation_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		policy_id INTEGER NOT NULL,
		trigger_event TEXT NOT NULL,
		threshold_percentage INTEGER NOT NULL DEFAULT 0,
		action_type TEXT NOT NULL,
		target_user_id TEXT NOT NULL DEFAULT '',
		target_priority TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		FOREIGN KEY (policy_id) REFERENCES sla_policies(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(slaEscalationRulesQuery)
	if err != nil {
		return err
	}

	taskSlaRecordsQuery := `
	CREATE TABLE IF NOT EXISTS task_sla_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER UNIQUE NOT NULL,
		policy_id INTEGER NOT NULL,
		workspace_id INTEGER NOT NULL,
		response_deadline DATETIME NOT NULL,
		resolution_deadline DATETIME NOT NULL,
		responded_at DATETIME,
		resolved_at DATETIME,
		response_status TEXT NOT NULL DEFAULT 'pending',
		resolution_status TEXT NOT NULL DEFAULT 'pending',
		is_paused INTEGER NOT NULL DEFAULT 0,
		paused_at DATETIME,
		total_paused_minutes INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (policy_id) REFERENCES sla_policies(id) ON DELETE CASCADE,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(taskSlaRecordsQuery)
	if err != nil {
		return err
	}

	slaBreachLogsQuery := `
	CREATE TABLE IF NOT EXISTS sla_breach_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		policy_id INTEGER NOT NULL,
		workspace_id INTEGER NOT NULL,
		breach_type TEXT NOT NULL,
		breached_at DATETIME NOT NULL,
		minutes_overdue INTEGER NOT NULL DEFAULT 0,
		escalation_triggered INTEGER NOT NULL DEFAULT 0,
		escalation_action TEXT NOT NULL DEFAULT '',
		details TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (policy_id) REFERENCES sla_policies(id) ON DELETE CASCADE,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(slaBreachLogsQuery)
	if err != nil {
		return err
	}

	goalsQuery := `
	CREATE TABLE IF NOT EXISTS goals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workspace_id INTEGER NOT NULL,
		owner_user_id TEXT NOT NULL,
		parent_goal_id INTEGER,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		timeframe TEXT NOT NULL DEFAULT '',
		start_date DATETIME,
		end_date DATETIME,
		status TEXT NOT NULL DEFAULT 'draft',
		progress_percentage REAL NOT NULL DEFAULT 0.0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
		FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_goal_id) REFERENCES goals(id) ON DELETE SET NULL
	);`

	_, err = db.Exec(goalsQuery)
	if err != nil {
		return err
	}

	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_time_entries_user_time ON time_entries(user_id, started_at);
	CREATE INDEX IF NOT EXISTS idx_time_entries_task_id ON time_entries(task_id);
	CREATE INDEX IF NOT EXISTS idx_task_dependencies_task_id ON task_dependencies(task_id);
	CREATE INDEX IF NOT EXISTS idx_task_dependencies_depends_on ON task_dependencies(depends_on_task_id);
	CREATE INDEX IF NOT EXISTS idx_activities_user_created ON activities(user_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_activities_task_created ON activities(task_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_activities_action ON activities(action);
	CREATE INDEX IF NOT EXISTS idx_workspaces_owner_id ON workspaces(owner_id);
	CREATE INDEX IF NOT EXISTS idx_workspace_members_user ON workspace_members(user_id);
	CREATE INDEX IF NOT EXISTS idx_workspace_members_ws ON workspace_members(workspace_id, role);
	CREATE INDEX IF NOT EXISTS idx_workspace_invites_token ON workspace_invites(token);
	CREATE INDEX IF NOT EXISTS idx_workspace_invites_ws ON workspace_invites(workspace_id, status);
	CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read);
	CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_webhook_subs_workspace ON webhook_subscriptions(workspace_id, is_active);
	CREATE INDEX IF NOT EXISTS idx_webhook_subs_user ON webhook_subscriptions(user_id);
	CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_sub ON webhook_deliveries(subscription_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status);
	CREATE INDEX IF NOT EXISTS idx_sprints_ws_proj ON sprints(workspace_id, project_id);
	CREATE INDEX IF NOT EXISTS idx_sprints_status ON sprints(project_id, status);
	CREATE INDEX IF NOT EXISTS idx_sprint_tasks_sprint ON sprint_tasks(sprint_id);
	CREATE INDEX IF NOT EXISTS idx_sprint_tasks_task ON sprint_tasks(task_id);
	CREATE INDEX IF NOT EXISTS idx_custom_fields_ws ON custom_field_definitions(workspace_id, project_id);
	CREATE INDEX IF NOT EXISTS idx_custom_fields_key ON custom_field_definitions(workspace_id, field_key);
	CREATE INDEX IF NOT EXISTS idx_custom_values_field ON custom_field_values(field_id);
	CREATE INDEX IF NOT EXISTS idx_custom_values_task ON custom_field_values(task_id);
	CREATE INDEX IF NOT EXISTS idx_custom_values_text ON custom_field_values(field_id, value_text);
	CREATE INDEX IF NOT EXISTS idx_custom_values_num ON custom_field_values(field_id, value_number);
	CREATE INDEX IF NOT EXISTS idx_attachments_task ON attachments(task_id);
	CREATE INDEX IF NOT EXISTS idx_attachments_ws ON attachments(workspace_id);
	CREATE INDEX IF NOT EXISTS idx_attachments_checksum ON attachments(workspace_id, checksum_sha256);
	CREATE INDEX IF NOT EXISTS idx_attachments_user ON attachments(user_id);
	CREATE INDEX IF NOT EXISTS idx_task_templates_ws ON task_templates(workspace_id);
	CREATE INDEX IF NOT EXISTS idx_task_templates_category ON task_templates(workspace_id, category);
	CREATE INDEX IF NOT EXISTS idx_task_templates_usage ON task_templates(workspace_id, usage_count DESC);
	CREATE INDEX IF NOT EXISTS idx_saved_views_ws ON saved_views(workspace_id);
	CREATE INDEX IF NOT EXISTS idx_saved_views_user ON saved_views(user_id);
	CREATE INDEX IF NOT EXISTS idx_saved_views_pinned ON saved_views(workspace_id, is_pinned);
	CREATE INDEX IF NOT EXISTS idx_saved_views_usage ON saved_views(workspace_id, usage_count DESC);
	CREATE INDEX IF NOT EXISTS idx_automation_rules_ws ON automation_rules(workspace_id, is_active);
	CREATE INDEX IF NOT EXISTS idx_automation_rules_trigger ON automation_rules(workspace_id, trigger_type);
	CREATE INDEX IF NOT EXISTS idx_automation_logs_rule ON automation_execution_logs(rule_id, executed_at DESC);
	CREATE INDEX IF NOT EXISTS idx_automation_logs_ws ON automation_execution_logs(workspace_id, status);
	CREATE INDEX IF NOT EXISTS idx_releases_ws_status ON releases(workspace_id, status);
	CREATE INDEX IF NOT EXISTS idx_releases_proj ON releases(workspace_id, project_id);
	CREATE INDEX IF NOT EXISTS idx_releases_version ON releases(workspace_id, version);
	CREATE INDEX IF NOT EXISTS idx_release_items_release ON release_items(release_id);
	CREATE INDEX IF NOT EXISTS idx_release_items_task ON release_items(task_id);
	CREATE INDEX IF NOT EXISTS idx_release_items_cat ON release_items(release_id, category);
	CREATE INDEX IF NOT EXISTS idx_intake_forms_ws ON intake_forms(workspace_id);
	CREATE INDEX IF NOT EXISTS idx_intake_forms_slug ON intake_forms(slug);
	CREATE INDEX IF NOT EXISTS idx_intake_forms_status ON intake_forms(workspace_id, status);
	CREATE INDEX IF NOT EXISTS idx_intake_form_fields_form ON intake_form_fields(form_id, position);
	CREATE INDEX IF NOT EXISTS idx_intake_form_submissions_form ON intake_form_submissions(form_id, submitted_at DESC);
	CREATE INDEX IF NOT EXISTS idx_intake_form_submissions_task ON intake_form_submissions(task_id);
	CREATE INDEX IF NOT EXISTS idx_sla_policies_ws ON sla_policies(workspace_id, is_active);
	CREATE INDEX IF NOT EXISTS idx_sla_escalation_rules_policy ON sla_escalation_rules(policy_id, trigger_event);
	CREATE INDEX IF NOT EXISTS idx_task_sla_records_task ON task_sla_records(task_id);
	CREATE INDEX IF NOT EXISTS idx_task_sla_records_ws_status ON task_sla_records(workspace_id, resolution_status);
	CREATE INDEX IF NOT EXISTS idx_task_sla_records_deadlines ON task_sla_records(response_deadline, resolution_deadline);
	CREATE INDEX IF NOT EXISTS idx_sla_breach_logs_ws ON sla_breach_logs(workspace_id, breached_at DESC);
	CREATE INDEX IF NOT EXISTS idx_sla_breach_logs_task ON sla_breach_logs(task_id, breach_type);
	`

	_, err = db.Exec(indexQuery)
	return err
}
