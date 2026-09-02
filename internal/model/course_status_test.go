package model

import (
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
)

func TestDetermineStatus(test *testing.T) {
	statuses := map[string]*CourseStatus{
		"course-admin@test.edulinq.org": {
			Active:  false,
			Source:  StatusSourceCourse,
			Owner:   "course-admin@test.edulinq.org",
			SetTime: timestamp.Zero(),
		},
		"server-owner@test.edulinq.org": {
			Active:  true,
			Source:  StatusSourceServer,
			Owner:   "server-owner@test.edulinq.org",
			SetTime: timestamp.Zero(),
		},
	}

	if determineStatus(statuses).Owner != "server-owner@test.edulinq.org" {
		test.Fatalf("Expected server-owner to win (higher source), got '%s'.",
			determineStatus(statuses).Owner)
	}

	// Adding a more recent StatusSourceServer
	statuses["server-admin@test.edulinq.org"] = &CourseStatus{
		Active:  false,
		Source:  StatusSourceServer,
		Owner:   "server-admin@test.edulinq.org",
		SetTime: timestamp.FromMSecs(100),
	}

	if determineStatus(statuses).Owner != "server-admin@test.edulinq.org" {
		test.Fatalf("Expected server-admin to win (more recent), got '%s'.",
			determineStatus(statuses).Owner)
	}
}
