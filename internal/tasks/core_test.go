package tasks

import (
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestTaskCoreRunOneTask(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	resetTestTaskCalls()
	defer resetTestTaskCalls()

	enableTaskEngine = true
	defer func() {
		enableTaskEngine = false
	}()

	task := &model.FullScheduledTask{
		UserTaskInfo: model.UserTaskInfo{
			Type: model.TaskTypeTest,
			When: &util.ScheduledTime{
				Daily: "0:00",
			},
		},
		SystemTaskInfo: model.SystemTaskInfo{
			Source:      model.TaskSourceTest,
			LastRunTime: timestamp.Zero(),
			NextRunTime: timestamp.Zero(),
			Hash:        "ABC",
		},
	}

	db.MustUpsertActiveTask(task)

	if testTaskCalls != 0 {
		test.Fatalf("Intial value for test task is wrong. Expected: %d, Actual: %d.", 0, testTaskCalls)
	}

	runNextTask()

	if testTaskCalls != 1 {
		test.Fatalf("Final value for test task is wrong. Expected: %d, Actual: %d.", 1, testTaskCalls)
	}
}

func TestTaskCoreSkipInactiveCourse(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	resetTestTaskCalls()
	defer resetTestTaskCalls()

	enableTaskEngine = true
	defer func() {
		enableTaskEngine = false
	}()

	db.MustUpsertCourseStatuses(db.MustGetCourse("course101"), map[string]*model.CourseStatus{"tester": {Active: false}})

	task := &model.FullScheduledTask{
		UserTaskInfo: model.UserTaskInfo{
			Type: model.TaskTypeTest,
			When: &util.ScheduledTime{
				Daily: "0:00",
			},
		},
		SystemTaskInfo: model.SystemTaskInfo{
			Source:      model.TaskSourceTest,
			LastRunTime: timestamp.Zero(),
			NextRunTime: timestamp.Zero(),
			Hash:        "ABC",
			CourseID:    "course101",
		},
	}

	db.MustUpsertActiveTask(task)

	if testTaskCalls != 0 {
		test.Fatalf("Intial value for test task is wrong. Expected: %d, Actual: %d.", 0, testTaskCalls)
	}

	runNextTask()

	if testTaskCalls != 0 {
		test.Fatalf("Final value for test task is wrong. Expected: %d, Actual: %d.", 0, testTaskCalls)
	}
}
