package main

import (
	"fmt"
	"time"
)



func IsOverdue(task Task) bool {

	return task.DueDate.Before(time.Now())
}

func GetTaskStatus(task Task) string {

	if task.Completed {
		return "Task Completed"
	}

	if IsOverdue(task) {
		return "OverDue"
	}

	return "On track"
}
//git
func main() {
	task := Task {
		ID:			1,
		Title: 		"Buy gloceries",
		DueDate: 	time.Now().AddDate(0, 0, 3),
		Completed: 	false,

	}

	fmt.Println(GetTaskStatus(task))
}