package courses

import (
	"maps"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestGradebook(test *testing.T) {

	allUsers := map[string]map[string]string{
		"hw0": {
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		},
	}

	testCases := []struct {
		email       string
		targetUsers []model.CourseUserReference
		locator     string
		expected    map[string]map[string]string
	}{

		// Valid Permissions
		{"course-grader", nil, "", allUsers},
		{"course-admin", []model.CourseUserReference{}, "", allUsers},
		{"course-owner", []model.CourseUserReference{"*"}, "", allUsers},

		// Filtering
		{"course-grader", []model.CourseUserReference{"student"}, "", map[string]map[string]string{
			"hw0": {
				"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			},
		}},
		{"course-admin", []model.CourseUserReference{"-*"}, "", map[string]map[string]string{
			"hw0": {},
		}},

		// Valid Permissions, Role Escalation
		{"server-admin", []model.CourseUserReference{"*"}, "", allUsers},
		{"server-admin", []model.CourseUserReference{"student"}, "", map[string]map[string]string{
			"hw0": {
				"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			},
		}},

		// Invalid Permissions
		{"course-student", []model.CourseUserReference{"*"}, "-020", nil},
		{"course-other", []model.CourseUserReference{"*"}, "-020", nil},

		// Invalid Permissions, Role Escalation
		{"server-user", []model.CourseUserReference{"*"}, "-040", nil},
		{"server-creator", []model.CourseUserReference{"*"}, "-040", nil},

		// Invalid Inputs
		{"course-grader", []model.CourseUserReference{"ZZZ"}, "-644", nil},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-users": testCase.targetUsers,
		}

		response := core.SendTestAPIRequestFull(test, `courses/gradebook`, fields, nil, testCase.email)
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

		var responseContent GradebookResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		// Compare submission IDs rather than whole structs.
		actual := make(map[string]map[string]string, len(responseContent.Gradebook))
		for assignmentID, submissionInfos := range responseContent.Gradebook {
			ids := make(map[string]string, len(submissionInfos))
			for email, info := range submissionInfos {
				id := ""
				if info != nil {
					id = info.ID
				}

				ids[email] = id
			}

			actual[assignmentID] = ids
		}

		if len(actual) != len(testCase.expected) {
			test.Errorf("Case %d: Unexpected number of assignments. Expected: %d, actual: %d.",
				i, len(testCase.expected), len(actual))
			continue
		}

		for assignmentID, expectedIDs := range testCase.expected {
			actualIDs, ok := actual[assignmentID]
			if !ok {
				test.Errorf("Case %d: Missing assignment '%s'.", i, assignmentID)
				continue
			}

			if !maps.Equal(expectedIDs, actualIDs) {
				test.Errorf("Case %d: Submission IDs do not match for assignment '%s'. Expected: '%+v', actual: '%+v'.",
					i, assignmentID, expectedIDs, actualIDs)
			}
		}
	}
}
