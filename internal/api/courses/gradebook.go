package courses

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type GradebookRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	// The users to include in the gradebook.
	// If nothing is specified, all course users will be included.
	// Unknown users are ignored and will not raise an error.
	TargetUsers []model.CourseUserReference `json:"target-users"`

	// The assignments to include in the gradebook.
	// If nothing is specified, all course assignments will be included.
	// Unknown and malformed assignments are ignored and will not raise an error.
	TargetAssignments []string `json:"target-assignments"`
}

type GradebookResponse struct {
	Gradebook map[string]map[string]*model.SubmissionHistoryItem `json:"gradebook"`
}

// Get a gradebook (most recent score for each user on each assignment) for a course.
func HandleGradebook(request *GradebookRequest) (*GradebookResponse, *core.APIError) {
	if len(request.TargetUsers) == 0 {
		request.TargetUsers = model.NewAllCourseUserReference()
	}

	reference, err := model.ParseCourseUserReferences(request.TargetUsers)
	if err != nil {
		return nil, core.NewBadRequestError("-644", request, "Failed to parse target users.").Err(err)
	}

	if len(request.TargetAssignments) == 0 {
		for id := range request.Course.GetAssignments() {
			request.TargetAssignments = append(request.TargetAssignments, id)
		}
	}

	gradebook := make(map[string]map[string]*model.SubmissionHistoryItem, len(request.TargetAssignments))
	for _, id := range request.TargetAssignments {
		assignment := request.Course.GetAssignment(id)
		if assignment == nil {
			continue
		}

		submissionInfos, err := db.GetRecentSubmissionSurvey(assignment, reference)
		if err != nil {
			return nil, core.NewInternalError("-645", request, "Failed to get submission summaries.").Err(err)
		}

		gradebook[assignment.GetID()] = submissionInfos
	}

	return &GradebookResponse{gradebook}, nil
}
