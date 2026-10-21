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

## Current limitations

State is held in process memory and is reset when the server restarts. Run a single server instance; this implementation is not durable or shared across replicas. The API does not yet authenticate callers or isolate workspaces, so it should not be exposed to an untrusted network. Persistent storage and an identity/authorization model should be selected before using it for multi-user or production data.

## Validate

```sh
go test ./...
go test -race ./...
```