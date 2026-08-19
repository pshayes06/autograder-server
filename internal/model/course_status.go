package model

import (
	"github.com/edulinq/autograder/internal/timestamp"
)

type CourseStatus struct {
	Active bool         `json:"active"`
	Source StatusSource `json:"source"`

	// Email of the user who set the status, if applicable
	Owner string `json:"owner,omitempty"`

	// Optional message to describe the status reason
	Message string `json:"message,omitempty"`

	SetTime timestamp.Timestamp `json:"set-time"`
}

// Represents the source of a course status, which is either a user or automated process.
type StatusSource int

// SourceUnknown is the zero value and is not a valid source.
// SourceCourse is for users with admin/owner privileges in the course.
// SourceServer is for users with admin/owner privileges in the server.
// SourceAutomated is for automated processes.
// SourceRoot is for users with root privileges.
const (
	SourceUnknown   StatusSource = 0
	SourceCourse                 = 10
	SourceServer                 = 20
	SourceAutomated              = 30
	SourceRoot                   = 40
)

// Checks if a status has priority over another status, which is determined through the StatusSource hierarchy.
// Ties are broken by the SetTime, giving priority to the most recently set status.
func (this *CourseStatus) hasPriorityOver(other *CourseStatus) bool {
	if this.Source != other.Source {
		return this.Source > other.Source
	}
	return this.SetTime > other.SetTime
}

// Iterates through every status to determine which should actually be used
func determineStatus(statuses []*CourseStatus) *CourseStatus {
	var best *CourseStatus
	for _, status := range statuses {
		if best == nil || status.hasPriorityOver(best) {
			best = status
		}
	}
	return best
}
