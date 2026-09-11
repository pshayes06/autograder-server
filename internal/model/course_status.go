package model

import (
	"strings"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// Represents the initiator of a course status, which is either a user or automated process.
type StatusSource int

const (
	// The zero value, which is not a valid source.
	StatusSourceUnknown StatusSource = 0

	// For users with admin/owner privileges in the course.
	StatusSourceCourse = 10

	// For users with admin/owner privileges in the server.
	StatusSourceServer = 20

	// For automated processes.
	StatusSourceAutomated = 30

	// For users with root privileges.
	StatusSourceRoot = 40
)

type CourseStatus struct {
	Active bool         `json:"active"`
	Source StatusSource `json:"source"`

	// User who set the status.
	Owner string `json:"owner"`

	// Optional message to describe the status reason.
	Message string `json:"message,omitempty"`

	SetTime timestamp.Timestamp `json:"set-time"`
}

var statusSourceToString = map[StatusSource]string{
	StatusSourceUnknown:   "unknown",
	StatusSourceCourse:    "course",
	StatusSourceServer:    "server",
	StatusSourceAutomated: "automated",
	StatusSourceRoot:      "root",
}

var stringToStatusSource = util.MapReverse(statusSourceToString)

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
// Ties are awarded to the most recent SetTime, and if those are equal, the unique owner string are compared lexicographically.
func (this *CourseStatus) compareTo(other *CourseStatus) int {
	if this.Source != other.Source {
		if this.Source > other.Source {
			return 1
		}
		return -1
	}

	if this.SetTime > other.SetTime {
		return 1
	}

	if this.SetTime < other.SetTime {
		return -1
	}

	return strings.Compare(this.Owner, other.Owner)
}
