package model

import (
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type CourseStatus struct {
	Active bool         `json:"active"`
	Source StatusSource `json:"source"`

	// User who set the status
	Owner string `json:"owner"`

	// Optional message to describe the status reason
	Message string `json:"message,omitempty"`

	SetTime timestamp.Timestamp `json:"set-time"`
}

// Represents the source of a course status, which is either a user or automated process.
type StatusSource int

// StatusSourceUnknown is the zero value and is not a valid source.
// StatusSourceCourse is for users with admin/owner privileges in the course.
// StatusSourceServer is for users with admin/owner privileges in the server.
// StatusSourceAutomated is for automated processes.
// StatusSourceRoot is for users with root privileges.
const (
	StatusSourceUnknown   StatusSource = 0
	StatusSourceCourse                 = 10
	StatusSourceServer                 = 20
	StatusSourceAutomated              = 30
	StatusSourceRoot                   = 40
)

var statusSourceToString = map[StatusSource]string{
	StatusSourceUnknown:   "unknown",
	StatusSourceCourse:    "course",
	StatusSourceServer:    "server",
	StatusSourceAutomated: "automated",
	StatusSourceRoot:      "root",
}

var stringToStatusSource = map[string]StatusSource{
	"unknown":   StatusSourceUnknown,
	"course":    StatusSourceCourse,
	"server":    StatusSourceServer,
	"automated": StatusSourceAutomated,
	"root":      StatusSourceRoot,
}

func (this StatusSource) String() string {
	return statusSourceToString[this]
}

func (this StatusSource) MarshalJSON() ([]byte, error) {
	return util.MarshalEnum(this, statusSourceToString)
}

func (this *StatusSource) UnmarshalJSON(data []byte) error {
	value, err := util.UnmarshalEnum(data, stringToStatusSource, true)
	if err == nil {
		*this = *value
	}

	return err
}

// Checks if a status has priority over another status, which is determined through the StatusSource hierarchy.
// Ties are broken by the SetTime, giving priority to the most recently set status.
func (this *CourseStatus) hasPriorityOver(other *CourseStatus) bool {
	if this.Source != other.Source {
		return this.Source > other.Source
	}
	return this.SetTime > other.SetTime
}

// Iterates through every status to determine which should actually be used
func determineStatus(statuses map[string]*CourseStatus) *CourseStatus {
	var best *CourseStatus
	for _, status := range statuses {
		if best == nil || status.hasPriorityOver(best) {
			best = status
		}
	}
	return best
}
