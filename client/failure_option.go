package client

import "net/http"

// FailManager is a failure handler.
type FailManager interface {
	Check(*http.Response) error
}

type statusChecker struct {
	statusList []int
	raiseErr   error
}

func (checker statusChecker) Check(resp *http.Response) error {
	for _, status := range checker.statusList {
		if resp.StatusCode == status {
			return checker.raiseErr
		}
	}
	return nil
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
