package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

func GetCourseStatuses(course *model.Course) (map[string]*model.CourseStatus, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetCourseStatuses(course)
}

func UpsertCourseStatuses(courseID string, statuses map[string]*model.CourseStatus) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.UpsertCourseStatuses(courseID, statuses)
}

func MustUpsertCourseStatuses(courseID string, statuses map[string]*model.CourseStatus) {
	err := UpsertCourseStatuses(courseID, statuses)
	if err != nil {
		log.Fatal("Failed to upsert course statuses.", err)
	}
}
