package courses

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type GradebookRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	TargetUsers []model.CourseUserReference `json:"target-users"`
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

	gradebook := make(map[string]map[string]*model.SubmissionHistoryItem)
	for id, assignment := range request.Course.GetAssignments() {
		submissionInfos, err := db.GetRecentSubmissionSurvey(assignment, reference)
		if err != nil {
			return nil, core.NewInternalError("-645", request, "Failed to get submission summaries.").Err(err)
		}
		gradebook[id] = submissionInfos
	}

	return &GradebookResponse{gradebook}, nil
}
