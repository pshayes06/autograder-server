package status

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestSet(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		email          string
		active         bool
		force          bool
		locator        string
		expectedSource model.StatusSource
	}{
		// Basic status sets
		{
			"course-owner",
			false,
			false,
			"",
			model.StatusSourceCourse,
		},
		{
			"course-admin",
			true,
			false,
			"",
			model.StatusSourceCourse,
		},

		// Trying to overwrite without force
		{
			"course-admin",
			false,
			false,
			"-646",
			0,
		},

		// Overwrite with force
		{
			"course-admin",
			false,
			true,
			"",
			model.StatusSourceCourse,
		},

		// Invalid Permissions
		{
			"course-grader",
			false,
			false,
			"-020",
			0,
		},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"course-id": "course101",
			"active":    testCase.active,
			"force":     testCase.force,
		}

		response := core.SendTestAPIRequestFull(test, `courses/status/set`, fields, nil, testCase.email)
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

		var responseContent SetResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if responseContent.Status.Source != testCase.expectedSource {
			test.Errorf("Case %d: Wrong source. Expected '%s', found '%s'.", i, testCase.expectedSource, responseContent.Status.Source)
		}
	}
}
