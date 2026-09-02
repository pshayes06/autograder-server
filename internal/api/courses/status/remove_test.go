package status

import (
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestRemove(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	statuses := map[string]*model.CourseStatus{
		"course-admin@test.edulinq.org": {
			Active:  true,
			Source:  model.StatusSourceCourse,
			Owner:   "course-admin@test.edulinq.org",
			SetTime: timestamp.Now(),
		},
		"course-owner@test.edulinq.org": {
			Active:  false,
			Source:  model.StatusSourceCourse,
			Owner:   "course-owner@test.edulinq.org",
			SetTime: timestamp.Now(),
		},
		"server-admin@test.edulinq.org": {
			Active:  true,
			Source:  model.StatusSourceServer,
			Owner:   "server-admin@test.edulinq.org",
			SetTime: timestamp.Now(),
		},
		"server-owner@test.edulinq.org": {
			Active:  false,
			Source:  model.StatusSourceServer,
			Owner:   "server-owner@test.edulinq.org",
			SetTime: timestamp.Now(),
		},
	}

	db.MustUpsertCourseStatuses("course101", statuses)

	testCases := []struct {
		email    string
		target   string
		clear    bool
		locator  string
		expected []string
	}{
		// Basic deletion
		{
			"course-owner",
			"",
			false,
			"",
			[]string{"course-owner@test.edulinq.org"},
		},

		// Finding nothing to delete
		{
			"course-owner",
			"",
			false,
			"",
			[]string{},
		},
		{
			"course-owner",
			"fake-user@test.edulinq.org",
			false,
			"",
			[]string{},
		},

		// Deleting and clearing targets with higher status
		{
			"course-owner",
			"server-admin@test.edulinq.org",
			false,
			"-648",
			nil,
		},
		{
			"course-admin",
			"",
			true,
			"",
			[]string{"course-admin@test.edulinq.org"},
		},

		// True clear when caller has highest StatusSource
		{
			"server-admin",
			"",
			true,
			"",
			[]string{"server-admin@test.edulinq.org", "server-owner@test.edulinq.org"},
		},

		// Clearing when empty
		{
			"server-owner",
			"",
			true,
			"",
			[]string{},
		},

		// Invalid Permissions
		{
			"course-grader",
			"",
			false,
			"-020",
			nil,
		},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"course-id":    "course101",
			"target-owner": testCase.target,
			"clear":        testCase.clear,
		}

		response := core.SendTestAPIRequestFull(test, `courses/status/remove`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.locator != "" {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.locator != "" {
			test.Errorf("Case %d: Did not get an expected error: '%s'.", i, testCase.locator)
			continue
		}

		var responseContent RemoveResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		removed := responseContent.Removed
		slices.Sort(removed)

		if !slices.Equal(removed, testCase.expected) {
			test.Errorf("Case %d: Expected '%v', found '%v'.", i, testCase.expected, removed)
		}
	}

	finalStatuses, err := db.GetCourseStatuses(db.MustGetCourse("course101"))
	if err != nil {
		test.Fatalf("Failed to get final statuses: '%v'.", err)
	}
	if len(finalStatuses) != 0 {
		test.Fatalf("Expected all statuses removed, found %d.", len(finalStatuses))
	}
}
