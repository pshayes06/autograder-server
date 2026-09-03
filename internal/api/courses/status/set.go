package status

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type SetRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	// Indicates whether course should be active or inactive.
	Active bool `json:"active"`

	// Optional message to include with status change.
	Message string `json:"message"`

	// If status already exists for user, allows overwrite for existing status.
	Force bool `json:"force"`
}

type SetResponse struct {
	Status *model.CourseStatus `json:"status"`
}

func HandleSet(request *SetRequest) (*SetResponse, *core.APIError) {

	owner := request.ServerUser.Email

	_, ok := request.Course.Statuses[owner]
	if ok && !request.Force {
		return nil, core.NewBadRequestError("-646", request, "Existing status for owner (must force to overwrite).")
	}

	status := &model.CourseStatus{
		Active:  request.Active,
		Source:  determineSource(request.ServerUser.Role),
		Owner:   owner,
		Message: request.Message,
		SetTime: request.Timestamp,
	}

	err := db.UpsertCourseStatuses(request.Course, map[string]*model.CourseStatus{owner: status})
	if err != nil {
		return nil, core.NewInternalError("-647", request, "Failed to upsert status.").Err(err)
	}

	log.Info(
		"Course status set.",
		request.Course,
		request.ServerUser,
		log.NewAttr("message", request.Message),
		log.NewAttr("active", request.Active),
		log.NewAttr("source", status.Source),
	)

	return &SetResponse{Status: status}, nil
}

// Finds the highest StatusSource using the user's server role
func determineSource(serverRole model.ServerUserRole) model.StatusSource {
	if serverRole >= model.ServerRoleRoot {
		return model.StatusSourceRoot
	}

	if serverRole >= model.ServerRoleAdmin {
		return model.StatusSourceServer
	}

	return model.StatusSourceCourse
}
