package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

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

// HasValidationErrors creates a FailManager that fails if the status is StatusBadRequest
// and the body includes validation errors.
// The error returned will be of type ValidationErrors wrapped under the given error.
func HasValidationErrors(err error) FailManager {
	return validationErrorsChecker{raiseErr: err}
}

type validationErrorsChecker struct {
	raiseErr error
}

func (checker validationErrorsChecker) Check(resp *http.Response) error {
	if resp.StatusCode != http.StatusBadRequest {
		return nil
	}
	// Retrieve the body
	body, err := ioutil.ReadAll(resp.Body)
	defer func(bdy []byte) {
		// reinject the body because it has been consumed previously, in case someone wants to use it.
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bdy))
	}(body)
	if err != nil {
		return fmt.Errorf("%s: %w", checker.raiseErr.Error(), fmt.Errorf("Failed to read body to check for validation errors: %w", err))
	}
	validationErrors, err := ParseValidationErrors(body, checker.raiseErr.Error())
	if err != nil {
		return fmt.Errorf("%s: %w", checker.raiseErr.Error(), fmt.Errorf("Falied to parse validation errors from response body: %w", err))
	}
	// In the case the validation result is valid.
	if validationErrors.Valid {
		return nil
	}
	return validationErrors
}
