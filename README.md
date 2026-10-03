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
GET /tasks?q=release&completed=false&overdue=false&priority=high&tag=release&dueAfter=2030-01-01T00:00:00Z&dueBefore=2030-01-31T23:59:59Z&offset=0&limit=20&sort=dueDate&order=asc
```

Search matches title, description, tags, and checklist text. Sort fields include `id`, `title`, `dueDate`, `priority`, `completed`, `estimateMinutes`, `updatedAt`, and `status`.

Other task workflows include:

- `PATCH /tasks/{id}` to update fields, including `status`.
- `PATCH /tasks/{id}/complete` to complete or reopen a task.
- `GET /tasks/ready`, `/tasks/blocked`, and `/tasks/upcoming?days=7` for actionable queue views.
- `GET /tasks/summary` for status counts, priority distribution, upcoming work, estimates, and tracked time.
- `GET /tags` for a paginated usage-ranked tag catalog.
- `GET|POST /tasks/{id}/comments` and `PATCH|DELETE /tasks/{id}/comments/{commentId}` for task discussion.
- `GET|POST /tasks/{id}/checklist`, plus checklist item completion and ordering routes.
- Dependency, recurrence preview/skip, activity, and timer routes under `/tasks/{id}`.
- `POST /tasks/bulk/complete`, `/tasks/bulk/delete`, `/tasks/bulk/priority`, `/tasks/bulk/due-date`, `/tasks/bulk/tags`, `/tasks/bulk/status`, and `/tasks/bulk/estimate` for batch workflows.

Task creation, update, and comment requests reject unknown fields and enforce request-size limits. List pagination is bounded to 100 records per request.

## Current limitations

State is held in process memory and is reset when the server restarts. Run a single server instance; this implementation is not durable or shared across replicas. The API does not yet authenticate callers or isolate workspaces, so it should not be exposed to an untrusted network. Persistent storage and an identity/authorization model should be selected before using it for multi-user or production data.

## Validate

```sh
go test ./...
go test -race ./...
```