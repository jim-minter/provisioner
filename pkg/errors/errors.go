package errors

import (
	"fmt"
)

type StatusCodeError int

func (err StatusCodeError) Error() string {
	return fmt.Sprintf("status code %d", err)
}
