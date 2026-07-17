package courses

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// Flatten a gradebook response into submission IDs rather than comparing whole structs.
func flattenGradebook(content any) map[string]map[string]string {
	var responseContent GradebookResponse
	util.MustJSONFromString(util.MustToJSON(content), &responseContent)

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

	return actual
}

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
		// Full Gradebook
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

		// Assignment ID Normalization
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

		// Invalid User (error)
		{
			"course-grader",
			[]model.CourseUserReference{"ZZZ"},
			nil,
			"-644",
			nil,
		},

		// Unknown Assignment (ignored)
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

		actual := flattenGradebook(response.Content)

		if !reflect.DeepEqual(testCase.expected, actual) {
			test.Errorf("Case %d: Unexpected gradebook. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(actual))
		}
	}
}

func TestGradebookAssignmentFiltering(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testSubmissions := []*model.GradingResult{
		{Info: &model.GradingInfo{
			ID:           "course-languages::bash::course-grader@test.edulinq.org::1234567890",
			ShortID:      "1234567890",
			CourseID:     "course-languages",
			AssignmentID: "bash",
			User:         "course-grader@test.edulinq.org",
		}},
		{Info: &model.GradingInfo{
			ID:           "course-languages::cpp::course-student@test.edulinq.org::1234567890",
			ShortID:      "1234567890",
			CourseID:     "course-languages",
			AssignmentID: "cpp",
			User:         "course-student@test.edulinq.org",
		}},
	}

	for _, submission := range testSubmissions {
		assignment := db.MustGetAssignment(submission.Info.CourseID, submission.Info.AssignmentID)
		if err := db.SaveSubmission(assignment, submission); err != nil {
			test.Fatalf("Failed to insert test submission: '%v'.", err)
		}
	}

	fullLanguagesGradebook := map[string]map[string]string{
		"bash": {
			"course-student@test.edulinq.org": "course-languages::bash::course-student@test.edulinq.org::1768603685",
			"course-grader@test.edulinq.org":  "course-languages::bash::course-grader@test.edulinq.org::1234567890",
			"course-other@test.edulinq.org":   "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		},
		"cpp": {
			"course-student@test.edulinq.org": "course-languages::cpp::course-student@test.edulinq.org::1234567890",
			"course-grader@test.edulinq.org":  "",
			"course-other@test.edulinq.org":   "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		},
		"java": {
			"course-student@test.edulinq.org": "",
			"course-grader@test.edulinq.org":  "",
			"course-other@test.edulinq.org":   "",
			"course-admin@test.edulinq.org":   "",
			"course-owner@test.edulinq.org":   "",
		},
	}

	testCases := []struct {
		targetAssignments []string
		expected          map[string]map[string]string
	}{
		// All Assignments
		{
			nil,
			fullLanguagesGradebook,
		},

		// Select Assignments
		{
			[]string{"bash"},
			map[string]map[string]string{
				"bash": fullLanguagesGradebook["bash"],
			},
		},
		{
			[]string{"bash", "cpp"},
			map[string]map[string]string{
				"bash": fullLanguagesGradebook["bash"],
				"cpp":  fullLanguagesGradebook["cpp"],
			},
		},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"course-id":          "course-languages",
			"target-assignments": testCase.targetAssignments,
		}

		response := core.SendTestAPIRequestFull(test, `courses/gradebook`, fields, nil, "course-grader")
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			continue
		}

		actual := flattenGradebook(response.Content)

		if !reflect.DeepEqual(testCase.expected, actual) {
			test.Errorf("Case %d: Unexpected gradebook. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(actual))
		}
	}
}
