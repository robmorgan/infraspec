package awshelpers

import (
	"os"
	"testing"

	"github.com/robmorgan/infraspec/test/testhelpers"
)

func TestMain(m *testing.M) {
	testhelpers.SetupAWSTestsAndConfig()
	code := m.Run()
	testhelpers.CleanupAwsTestEnvironment()
	os.Exit(code)
}
