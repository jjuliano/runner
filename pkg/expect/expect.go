// expect.go
package expect

import (
	"github.com/runner/pkg/expect/check"
	"github.com/runner/pkg/expect/process"
)

var (
	ProcessExpectations = process.ProcessExpectations
	CheckExpectations   = check.CheckExpectations
)
