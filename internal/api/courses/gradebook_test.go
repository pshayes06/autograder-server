package courses

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestGradebook(test *testing.T) {
	fullGradebook := map[string]map[string]string{
		"hw0": {
			"course-other@test.edulinq.org":   "",
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
			"course-grader@test.edulinq.org":  "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		},
	}

	studentOnlyGradebook := map[string]map[string]string{
		"hw0": {
			"course-student@test.edulinq.org": "course101::hw0::course-student@test.edulinq.org::1697406272",
		},
	}

	testCases := []struct {
		email             string
		targetUsers       []model.CourseUserReference
		targetAssignments []string
		locator           string
		expected          map[string]map[string]string
	}{
		// Valid Permissions
		{
			"course-grader",
			nil,
			nil,
			"",
			fullGradebook,
		},
		{
			"course-admin",
			[]model.CourseUserReference{},
			[]string{},
			"",
			fullGradebook,
		},
		{
			"course-owner",
			[]model.CourseUserReference{"*"},
			nil,
			"",
			fullGradebook,
		},

		// User Filtering
		{
			"course-grader",
			[]model.CourseUserReference{"student"},
			nil,
			"",
			studentOnlyGradebook,
		},
		{
			"course-admin",
			[]model.CourseUserReference{"-*"},
			nil,
			"",
			map[string]map[string]string{
				"hw0": {},
			},
		},

		// Assignment Filtering
		{
			"course-grader",
			nil,
			[]string{"hw0"},
			"",
			fullGradebook,
		},

		// Assignment ID Normalization Check
		{
			"course-grader",
			nil,
			[]string{"HW0"},
			"",
			fullGradebook,
		},

		// Invalid Permissions
		{
			"course-student",
			[]model.CourseUserReference{"*"},
			nil,
			"-020",
			nil,
		},
		{
			"course-other",
			[]model.CourseUserReference{"*"},
			nil,
			"-020",
			nil,
		},

		// Failure on malformed role
		{
			"course-grader",
			[]model.CourseUserReference{"ZZZ"},
			nil,
			"-644",
			nil,
		},

		// Ignored unknown/malformed assignment
		{
			"course-admin",
			nil,
			[]string{"ZZZ"},
			"",
			map[string]map[string]string{},
		},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"target-users":       testCase.targetUsers,
			"target-assignments": testCase.targetAssignments,
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

		if !reflect.DeepEqual(testCase.expected, actual) {
			test.Errorf("Case %d: Unexpected gradebook. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(actual))
		}
	}
}
