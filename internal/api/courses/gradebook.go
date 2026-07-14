package courses

import (
	"fmt"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type GradebookRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	TargetUsers       []model.CourseUserReference `json:"target-users"`
	TargetAssignments []string                    `json:"target-assignments"`
}

type GradebookResponse struct {
	Gradebook map[string]map[string]*model.SubmissionHistoryItem `json:"gradebook"`
}

// Get gradebook (most recent score for each user on each assignment) for a course.
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
			return nil, core.NewBadRequestError("-645", request,
				fmt.Sprintf("Could not find assignment: '%s'.", id))

		}

		submissionInfos, err := db.GetRecentSubmissionSurvey(assignment, reference)
		if err != nil {
			return nil, core.NewInternalError("-646", request, "Failed to get submission summaries.").Err(err)
		}

		gradebook[assignment.GetID()] = submissionInfos
	}

	return &GradebookResponse{gradebook}, nil
}
