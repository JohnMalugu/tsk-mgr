package model

type TaskTimeSummaryDB struct {
	TaskID           int   `json:"task_id"`
	EstimatedSeconds int64 `json:"estimated_seconds"`
	ActualSeconds    int64 `json:"actual_seconds"`
	VarianceSeconds  int64 `json:"variance_seconds"`
}

type TaskTimeTotalDB struct {
	TaskID       int    `json:"task_id"`
	TaskTitle    string `json:"task_title"`
	TotalSeconds int64  `json:"total_seconds"`
}

type TimeReportDB struct {
	TotalSeconds int64             `json:"total_seconds"`
	EntryCount   int               `json:"entry_count"`
	ActiveCount  int               `json:"active_count"`
	ByTask       []TaskTimeTotalDB `json:"by_task"`
}
