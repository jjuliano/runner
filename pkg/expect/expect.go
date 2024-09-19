// expect.go
package expect

import (
	"github.com/runner/pkg/plugins/expect/check"
	"github.com/runner/pkg/plugins/expect/process"
)

var (
	ProcessExpectations = process.ProcessExpectations
	CheckExpectations   = check.CheckExpectations
)
