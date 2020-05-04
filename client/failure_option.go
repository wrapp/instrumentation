package client

import "net/http"

// FailManager is a failure handler.
type FailManager interface {
	Check(*http.Response) error
}

func contains(status int, statusList []int) bool {
	for _, s := range statusList {
		if s == status {
			return true
		}
	}

	return false
}

// statusChecker contains the implementation where the status are matching a given
// list.
type statusChecker struct {
	statusList  []int
	raiseErr    error
	mustContain bool
}

func (checker statusChecker) Check(resp *http.Response) error {
	if contains(resp.StatusCode, checker.statusList) == checker.mustContain {
		return nil
	}
	return checker.raiseErr
}

// StatusChecker creates a FailManager that fails when some status are returned.
func StatusChecker(err error, statusList ...int) FailManager {
	return statusChecker{raiseErr: err, statusList: statusList}
}

// StatusBetween creates a FailManager that fails when the status is between two
// status.
func StatusBetween(err error, min, max int) FailManager {
	invalidStatus := make([]int, 0, max-min+1)
	for i := min; i <= max; i++ {
		invalidStatus = append(invalidStatus, i)
	}
	return statusChecker{raiseErr: err, statusList: invalidStatus}
}

// StatusIsNot creates a FailManager that fails if the status is not in the list.
func StatusIsNot(err error, statusList ...int) FailManager {
	return statusChecker{raiseErr: err, statusList: statusList, mustContain: true}
}
