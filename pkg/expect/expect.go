// expect.go
package expect

import (
	"github.com/jjuliano/runner/pkg/expect/check"
	"github.com/jjuliano/runner/pkg/expect/process"
)

var (
	ProcessExpectations = process.ProcessExpectations
	CheckExpectations   = check.CheckExpectations
)
