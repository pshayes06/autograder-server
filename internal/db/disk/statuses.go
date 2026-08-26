package disk

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const DISK_DB_STATUSES_FILENAME = "statuses.json"

func (this *backend) GetCourseStatuses(course *model.Course) (map[string]*model.CourseStatus, error) {
	this.statusLock.RLock()
	defer this.statusLock.RUnlock()

	allStatuses, err := this.getStatuses(course.ID)
	if err != nil {
		return nil, err
	}

	return allStatuses, nil
}

func (this *backend) UpsertCourseStatuses(courseID string, upsertStatuses map[string]*model.CourseStatus) error {
	this.statusLock.Lock()
	defer this.statusLock.Unlock()

	statuses, err := this.getStatuses(courseID)
	if err != nil {
		return err
	}

	for owner, status := range upsertStatuses {
		if status == nil {
			delete(statuses, owner)
		} else {
			statuses[owner] = status
		}
	}

	return this.writeStatuses(courseID, statuses)
}

func (this *backend) getStatusesPath(courseID string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), DISK_DB_STATUSES_FILENAME)
}

func (this *backend) getStatuses(courseID string) (map[string]*model.CourseStatus, error) {
	statuses := make(map[string]*model.CourseStatus, 0)

	path := this.getStatusesPath(courseID)
	if !util.PathExists(path) {
		return statuses, nil
	}

	err := util.JSONFromFile(path, &statuses)
	if err != nil {
		return nil, fmt.Errorf("Failed to read statuses file '%s': '%w'.", path, err)
	}

	return statuses, nil
}

func (this *backend) writeStatuses(courseID string, statuses map[string]*model.CourseStatus) error {
	path := this.getStatusesPath(courseID)

	err := util.MkDir(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("Failed to create directory for statuses file '%s': '%w'.", path, err)
	}

	err = util.ToJSONFileIndent(statuses, path)
	if err != nil {
		return fmt.Errorf("Failed to write statuses file '%s': '%w'.", path, err)
	}

	return nil
}
