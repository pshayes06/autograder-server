package db

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestUpsertStatus(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	course := MustGetTestCourse()

	MustUpsertCourseStatuses(course, testStatuses)

	statuses, err := GetCourseStatuses(course)
	if err != nil {
		test.Fatalf("Failed to fetch initial statuses: '%v'.", err)
	}

	if !reflect.DeepEqual(testStatuses, statuses) {
		test.Fatalf("Initial statuses are not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(testStatuses), util.MustToJSONIndent(statuses))
	}

	statuses["userA"] = nil

	statuses["userB"].Active = true

	delete(statuses, "userC")

	MustUpsertCourseStatuses(course, statuses)

	newStatuses, err := GetCourseStatuses(course)
	if err != nil {
		test.Fatalf("Failed to fetch new statuses: '%v'.", err)
	}

	if len(newStatuses) != (len(testStatuses) - 1) {
		test.Fatalf("New statuses hand unexpected count. Expected: %d, Actual: %d.", len(newStatuses)-1, len(newStatuses))
	}

	_, exists := newStatuses["userA"]
	if exists {
		test.Fatalf("Found status that should have been removed.")
	}

	if !newStatuses["userB"].Active {
		test.Fatalf("Expected user B's status to be active after update, got '%v'.", newStatuses["B"].Active)
	}
}

var testStatuses = map[string]*model.CourseStatus{
	"userA": {
		Active: false,
	},
	"userB": {
		Active: false,
	},
	"userC": {
		Active: false,
	},
}
