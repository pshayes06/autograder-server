package status

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type RemoveRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	// Email of status owner to remove. Defaults to the caller.
	TargetOwner string `json:"target-owner"`

	// If true, remove all statuses the caller has permission to remove.
	Clear bool `json:"clear"`

	// Optional log message to include with status removal.
	Message string `json:"message"`
}

type RemoveResponse struct {
	Removed []string `json:"removed"`
}

func HandleRemove(request *RemoveRequest) (*RemoveResponse, *core.APIError) {

	callerSource := determineSource(request.ServerUser.Role)
	courseStatuses := request.Course.Statuses

	deletionMap := make(map[string]*model.CourseStatus)

	if request.Clear {
		for owner, status := range courseStatuses {
			if callerSource >= status.Source {
				deletionMap[owner] = nil
			}
		}
	} else if request.TargetOwner != "" {
		targetStatus, ok := courseStatuses[request.TargetOwner]
		if ok {
			if callerSource < targetStatus.Source {
				return nil, core.NewPermissionsError("-648", request, targetStatus.Source, callerSource, "Cannot remove a status with a higher source.")
			}

			deletionMap[request.TargetOwner] = nil
		}
	} else {
		_, ok := courseStatuses[request.ServerUser.Email]
		if ok {
			deletionMap[request.ServerUser.Email] = nil
		}
	}

	if len(deletionMap) > 0 {
		err := db.UpsertCourseStatuses(request.Course, deletionMap)
		if err != nil {
			return nil, core.NewInternalError("-649", request, "Failed to upsert status.").Err(err)
		}
	}

	removed := []string{}

	for owner := range deletionMap {
		removed = append(removed, owner)
	}

	log.Info(
		"Course status removal.",
		request.Course,
		request.ServerUser,
		log.NewAttr("message", request.Message),
		log.NewAttr("removed", removed),
	)

	return &RemoveResponse{Removed: removed}, nil
}
