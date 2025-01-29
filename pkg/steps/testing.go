package steps

import (
	"github.com/gruntwork-io/terratest/modules/testing"
)

func GetT() testing.TestingT {
	return &specTestingT{}
}

type specTestingT struct {
	testing.TestingT
}

// Extends specTestingT to have #Name() method, that is compatible with testing.TestingT
func (t *specTestingT) Name() string {
	return "[InfraSpec]"
}
