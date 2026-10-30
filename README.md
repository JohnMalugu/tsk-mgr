# Task Manager API

A Go HTTP API for managing tasks, dependencies, checklists, recurring work, comments, activity, and time tracking.

## Run

```sh
go run ./cmd/server
```

The server listens on `http://localhost:8080`.

## Task workflow

Create a task with a due date and optional status, priority, tags, estimate, checklist, recurrence rule, or prerequisites:

```sh
curl -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Prepare release","description":"Review the release plan","dueDate":"2030-01-15T09:00:00Z","status":"in_progress","priority":"high","tags":["release"],"estimateMinutes":90}'
```

Task status values are `todo`, `in_progress`, `completed`, and `canceled`. The `completed` boolean remains available for older clients and is synchronized with status transitions. Completion still honors task dependencies.

List tasks with optional filters and sorting:

```text
GET /tasks?q=release&status=in_progress&completed=false&overdue=false&priority=high&tag=release&dueAfter=2030-01-01T00:00:00Z&dueBefore=2030-01-31T23:59:59Z&updatedAfter=2030-01-01T00:00:00Z&updatedBefore=2030-01-31T23:59:59Z&minEstimate=30&maxEstimate=120&checklistComplete=false&offset=0&limit=20&sort=dueDate&order=asc
```

Search matches title, description, tags, and checklist text. Sort fields include `id`, `title`, `dueDate`, `priority`, `completed`, `estimateMinutes`, `updatedAt`, and `status`.

Other task workflows include:

- `PATCH /tasks/{id}` to update fields, including `status`.
- `PATCH /tasks/{id}/complete` to complete or reopen a task.
- `POST /tasks/{id}/reopen` to return a canceled task to the todo queue.
- `POST /tasks/{id}/duplicate` to create a fresh copy of one task.
- `GET /tasks/ready`, `/tasks/blocked`, and `/tasks/upcoming?days=7` for actionable queue views.
- `GET /tasks/summary` for status counts, priority distribution, upcoming work, estimates, and tracked time.
- `GET /tags` for a paginated usage-ranked tag catalog.
- `GET|POST /tasks/{id}/comments` and `PATCH|DELETE /tasks/{id}/comments/{commentId}` for task discussion.
- `GET|POST /tasks/{id}/checklist`, plus checklist item completion and ordering routes.
- Dependency, recurrence preview/skip, activity, and timer routes under `/tasks/{id}`.
- `POST /tasks/bulk/complete`, `/tasks/bulk/delete`, `/tasks/bulk/duplicate`, `/tasks/bulk/priority`, `/tasks/bulk/due-date`, `/tasks/bulk/tags`, `/tasks/bulk/status`, and `/tasks/bulk/estimate` for batch workflows.

Task creation, update, and comment requests reject unknown fields and enforce request-size limits. List pagination is bounded to 100 records per request.
Live timers and manual time entries are accepted only for tasks that are not completed or canceled.

## Persistent V2 API & Time Tracking (SQLite)

The API includes authenticated, multi-tenant SQLite persistence under `/v2` with JWT authentication and Argon2id password hashing:

- `POST /v2/auth/register` & `POST /v2/auth/login`
- `GET|POST /v2/tasks`, `GET|PUT|DELETE /v2/tasks/{id}`, `PATCH /v2/tasks/{id}/complete`
- `GET|POST /v2/tasks/{id}/subtasks`, `PATCH|DELETE /v2/subtasks/{id}`
- `GET|POST /v2/tasks/{id}/comments`, `PATCH|DELETE /v2/comments/{id}`
- `POST /v2/tasks/{id}/timer/start` to start an active timer (ensures single concurrent timer per user)
- `POST /v2/timer/stop` to stop the current active timer and record elapsed seconds
- `GET /v2/timer` to fetch current active timer with live duration
- `GET|POST /v2/tasks/{id}/time` to list time entries or log manual time with start/end timestamps
- `GET /v2/tasks/{id}/time/summary` for actual vs. estimated variance analysis
- `GET /v2/time/report?from={start}&to={end}` for date-range aggregated time breakdown across tasks
- `DELETE /v2/time/{id}` to remove a time entry
- `POST /v2/tasks/{id}/dependencies` to link prerequisite tasks (enforces user isolation and DAG cycle prevention)
- `DELETE /v2/tasks/{id}/dependencies/{dependencyId}` to remove a prerequisite relationship
- `GET /v2/tasks/{id}/dependencies` to retrieve prerequisite tasks with enriched task details
- `GET /v2/tasks/{id}/dependents` to retrieve downstream dependent tasks
- `GET /v2/tasks/{id}/dag` for complete task DAG summary including blocked status and blocking task IDs
- `GET /v2/tasks/blocked` for queue of tasks blocked by incomplete prerequisites
- `GET /v2/tasks/ready` for actionable queue of unblocked tasks ready for work
- `GET /v2/activity?offset={n}&limit={n}&action={action}&from={timestamp}&to={timestamp}&taskId={id}` for paginated audit log feed with multi-criteria filtering
- `GET /v2/activity/summary?from={timestamp}&to={timestamp}` for aggregated activity metrics, action breakdown, and recent event stream
- `POST /v2/activity/purge` to execute audit log retention cleanup with configurable day cutoff
- `GET /v2/tasks/{id}/activity?offset={n}&limit={n}` for task-specific immutable changelog and audit trail
- `GET|POST /v2/workspaces` to list user workspaces or create a new workspace with automatic owner assignment
- `GET|PUT|DELETE /v2/workspaces/{id}` to fetch workspace details, update metadata, or delete workspace (with owner guard)
- `GET|POST /v2/workspaces/{id}/members` to list workspace members or add a member with role assignment (`owner`, `admin`, `member`, `viewer`)
- `PUT|DELETE /v2/workspaces/{id}/members/{userId}` to update member role or remove members (with owner protection and self-leave support)
- `POST /v2/workspaces/{id}/invites` to generate secure 7-day invitation tokens for team onboarding
- `POST /v2/invites/accept` to accept an invitation token and join the workspace with the designated role
- `GET /v2/notifications?offset={n}&limit={n}&type={type}&unreadOnly={true|false}` to list user notifications with counters and pagination
- `POST /v2/notifications` to trigger an in-app alert (enforces user topic and channel preferences)
- `GET /v2/notifications/unread-count` for real-time badge counters
- `POST /v2/notifications/read-all` to mark all unread notifications as read
- `POST|PATCH /v2/notifications/{id}/read` to mark an individual notification as read
- `DELETE /v2/notifications/{id}` to remove a notification record
- `GET|PUT|PATCH /v2/notifications/preferences` to retrieve or customize per-topic and channel notification preferences
- `GET /v2/webhooks?workspace_id={id}&is_active={true|false}&event_type={type}` to list registered webhook endpoints
- `POST /v2/webhooks` to register a webhook with event filtering and auto-generated HMAC-SHA256 secret token
- `GET|PUT|PATCH /v2/webhooks/{id}` to fetch subscription details, modify target URL, or toggle active state
- `DELETE /v2/webhooks/{id}` to remove a webhook subscription
- `POST /v2/webhooks/{id}/ping` to trigger an immediate live test delivery with `X-Webhook-Signature-256` HMAC validation
- `GET /v2/webhooks/{id}/deliveries?status={success|failed}` for chronological audit log of HTTP webhook dispatch attempts
- `GET /v2/webhooks/{id}/stats` for delivery success rate, average response latency, and health analytics
- `POST /v2/webhooks/deliveries/{deliveryId}/retry` to re-attempt an existing webhook delivery
- `GET|POST /v2/sprints` to query sprints by workspace/project/status or create a new sprint milestone
- `GET|PUT|DELETE /v2/sprints/{id}` to inspect sprint details, update metadata, or delete non-active sprints
- `POST /v2/sprints/{id}/start` to start a sprint milestone (enforces single-active sprint per project)
- `POST /v2/sprints/{id}/complete` to complete an active sprint and record completion timestamp
- `POST /v2/sprints/{id}/close` to archive a completed sprint
- `GET|POST /v2/sprints/{id}/tasks` to view sprint task backlog or scope a task into a sprint with Agile story points
- `PUT|PATCH /v2/sprints/{id}/tasks/{taskId}` to update task story points estimation
- `DELETE /v2/sprints/{id}/tasks/{taskId}` to remove a task from the sprint backlog
- `GET /v2/sprints/{id}/metrics` to calculate live burndown metrics, velocity, story points, and status breakdown
- `GET|POST /v2/custom-fields` to query custom field schemas by workspace/project/type or define a new dynamic attribute schema (`text`, `number`, `dropdown`, `boolean`, `date`, `url`)
- `GET|PUT|DELETE /v2/custom-fields/{id}` to fetch field schema, modify metadata and dropdown choices, or delete custom field
- `POST /v2/custom-fields/bulk` to bulk assign custom field values across multiple tasks in a single operation
- `GET /v2/custom-fields/search?workspace_id={id}&field_key={key}&query_value={val}` to search and filter tasks by specific custom field value
- `GET|POST|PUT /v2/tasks/{id}/custom-fields` to retrieve enriched custom fields for a task or assign a typed custom field value
- `DELETE /v2/tasks/{id}/custom-fields/{fieldId}` to remove a custom field assignment from a task
- `POST /v2/tasks/{id}/attachments` to attach a file to a task (supports multipart/form-data and JSON base64 payloads with SHA-256 deduplication and blocked executable security rules)
- `GET /v2/tasks/{id}/attachments` to list all file attachments associated with a specific task
- `GET /v2/attachments?workspace_id={id}&content_type={type}&is_archived={bool}&limit={n}&offset={n}` to filter attachments across a workspace
- `GET /v2/attachments/{id}` to retrieve attachment metadata, file size, MIME type, and checksum
- `PUT|PATCH /v2/attachments/{id}` to update attachment metadata (filename, archive status)
- `DELETE /v2/attachments/{id}` to remove an attachment record and delete unreferenced physical file assets from disk
- `GET /v2/attachments/{id}/download` to stream raw binary files with HTTP conditional caching (`ETag`, `If-None-Match`) and download tracking
- `GET /v2/attachments/metrics?workspace_id={id}` for workspace file storage quota, total bytes, download analytics, and MIME type breakdown
- `GET|POST /v2/templates` to search and filter templates by category/workspace or create a new template with subtask and checklist blueprints
- `GET|PUT|DELETE /v2/templates/{id}` to view template specification, modify metadata and blueprints, or delete a template
- `POST /v2/templates/{id}/duplicate` to clone an existing template with its subtask and checklist blueprints
- `POST /v2/templates/{id}/instantiate` to instantiate a new task from the template blueprint with automatic subtask generation and usage tracking
- `GET /v2/templates/{id}/export` to export a portable template bundle JSON packet
- `GET /v2/templates/stats?workspace_id={id}` for template usage analytics, category distribution, and most popular blueprints
- `GET|POST /v2/views` to list workspace views (with `pinned=true` filter) or register a new custom saved view with filter criteria, sorting, and grouping
- `GET|PUT|DELETE /v2/views/{id}` to fetch view details, update view filter criteria and styling, or delete a saved view
- `POST /v2/views/{id}/duplicate` to clone an existing saved view into a private copy or copy across workspaces
- `POST|PATCH /v2/views/{id}/pin` to toggle the pinned favorite status of a saved view
- `GET|POST /v2/views/{id}/execute` to execute a saved view with dynamic multi-dimensional task filtering, sorting, and group count aggregations
- `GET|POST /v2/automations` to list rules with workspace and trigger filters or create an event-driven automation rule (`WHEN trigger [IF conditions] THEN actions`)
- `GET|PUT|DELETE /v2/automations/{id}` to fetch rule specification, modify trigger conditions and actions, or delete an automation rule
- `POST|PATCH /v2/automations/{id}/toggle` to activate or deactivate an automation rule
- `POST /v2/automations/{id}/duplicate` to clone an automation rule with customizable name and target workspace
- `POST /v2/automations/{id}/dry-run` to simulate condition evaluation and preview estimated task mutations without altering persistent database records
- `GET /v2/automations/{id}/logs` to inspect chronological execution logs for a specific automation rule
- `GET /v2/automations/logs?workspace_id={id}&status={success|skipped|failed}` to query and filter execution audit trails across a workspace
- `GET /v2/automations/metrics?workspace_id={id}` for workspace automation analytics, execution success rates, and top triggered rules
- `POST /v2/automations/dispatch` to dispatch manual trigger events for integration testing and automated workflow verification
- `GET|POST /v2/releases` to query releases by workspace/project/status or provision a new release milestone with semver and target date
- `GET|PUT|DELETE /v2/releases/{id}` to fetch release metadata, modify target date and description, or delete a release
- `POST|PATCH /v2/releases/{id}/status` to transition release lifecycle (`planning` -> `active` -> `code_freeze` -> `released` -> `archived`) with automatic release timestamping
- `GET|POST /v2/releases/{id}/tasks` to view scoped release task backlog or attach tasks with categorization (`feature`, `bugfix`, `improvement`, `security`, `chore`)
- `PUT|PATCH|DELETE /v2/releases/{id}/tasks/{taskId}` to update changelog notes and category or detach a task from the release
- `GET|POST /v2/releases/{id}/changelog` to generate automated markdown or JSON release notes grouped by category
- `GET /v2/releases/{id}/metrics` to compute release burndown, completion rate, category breakdown, blockers, and health status (`on_track`, `at_risk`, `delayed`, `completed`)
- `POST /v2/releases/{id}/clone` to clone a release milestone into a new version with optional incomplete task rollover
- `GET|POST /v2/forms` to list workspace forms or provision a new task intake portal with custom vanity slug, default tags, priority, and assignees
- `GET|PUT|DELETE /v2/forms/{id}` to fetch form details, update form settings and task blueprints, or delete an intake form
- `POST /v2/forms/{id}/publish` to toggle public availability (`publish: true|false`) of an intake form
- `GET|PUT|POST /v2/forms/{id}/fields` to inspect field schema or batch update custom input fields with task attribute targets
- `GET /v2/forms/public/{slug}` for unauthenticated rendering of public forms and active input schemas
- `POST /v2/forms/public/{slug}/submit` to submit public responses with automatic validation and background task provisioning
- `GET /v2/forms/{id}/submissions?limit={limit}&offset={offset}` to inspect chronological submission logs and response payloads
- `GET /v2/forms/{id}/analytics` for form submission analytics, 7-day/30-day velocity, and daily intake volume breakdown
- `GET|POST /v2/slas` to list SLA policies with workspace and status filters or configure new service level agreements with priority tiers, target response/resolution minutes, and business hour constraints
- `GET|PUT|PATCH|DELETE /v2/slas/{id}` to fetch policy specifications, modify target response and resolution durations, or delete an SLA policy
- `POST /v2/slas/{id}/toggle` to activate or deactivate an SLA policy
- `GET|POST /v2/slas/{id}/escalations` to inspect or replace automated escalation rules (`at_risk` threshold percentage or `breached` triggers with priority escalation, task reassignment, or notification actions)
- `DELETE /v2/slas/escalations/{ruleId}` to remove a specific escalation rule
- `POST /v2/slas/tasks/apply` to attach an SLA policy to a task and compute response/resolution deadlines with business hours calculation
- `GET /v2/slas/tasks/{taskId}` to view live SLA target countdowns, pause state, and fulfillment stages (`pending`, `met`, `breached`)
- `POST /v2/slas/tasks/{taskId}/pause` to pause active SLA countdown timers during customer or external dependency holds
- `POST /v2/slas/tasks/{taskId}/resume` to resume paused SLA timers and forward-adjust response and resolution deadlines
- `POST /v2/slas/tasks/{taskId}/evaluate` to trigger on-demand SLA evaluation, transition stage fulfillment, execute escalation mitigations, and record breach events
- `GET /v2/slas/breaches?workspace_id={id}&task_id={id}&breach_type={response|resolution}` to query chronological breach audit trail logs
- `GET /v2/slas/metrics?workspace_id={id}` for workspace compliance rates, met vs. breached distribution, and response time analytics
- `GET|POST /v2/goals` to list goals with multi-criteria filters (`workspace_id`, `owner_user_id`, `status`, `timeframe`, `parent_goal_id`) or provision a new strategic objective
- `GET|PUT|PATCH|DELETE /v2/goals/{id}` to fetch goal details (with embedded key results and latest check-in), update goal parameters, or delete a goal
- `GET /v2/goals/{id}/tree` to retrieve the complete cascading hierarchical goal tree with nested sub-goals and child key results
- `GET|POST /v2/goals/{id}/key-results` to list key results or define a new measurable key result with typed metrics (`numerical`, `percentage`, `currency`, `boolean`), start/target/current values, and weights
- `GET|PUT|PATCH|DELETE /v2/goals/key-results/{id}` to inspect key result details, update values/metrics (triggering dynamic goal progress recalculation), or delete a key result
- `GET|POST /v2/goals/key-results/{id}/tasks` to view linked tasks or link a task to a key result with automatic task completion progress synchronization
- `DELETE /v2/goals/key-results/{id}/tasks/{taskId}` to unlink a task from a key result and re-sync progress
- `GET|POST /v2/goals/{id}/check-ins` to list chronological progress check-ins or submit a new status update with confidence ratings (`high`, `medium`, `low`), notes, and progress percentage
- `GET /v2/goals/metrics?workspace_id={id}` for workspace OKR health analytics, achievement rates, status distribution (`draft`, `active`, `on_track`, `behind`, `at_risk`, `achieved`, `canceled`), and average progress

## Current limitations

State is held in process memory and is reset when the server restarts. Run a single server instance; this implementation is not durable or shared across replicas. The API does not yet authenticate callers or isolate workspaces, so it should not be exposed to an untrusted network. Persistent storage and an identity/authorization model should be selected before using it for multi-user or production data.

## Validate

```sh
go test ./...
go test -race ./...
```