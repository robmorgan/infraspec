package aws

// AWSAsserter implements assertions for AWS resources
type AWSAsserter struct{}

// NewAWSAsserter creates a new AWSAsserter instance
func NewAWSAsserter() *AWSAsserter {
	return &AWSAsserter{}
}

// GetName returns the name of the asserter
func (a *AWSAsserter) GetName() string {
	return "aws"
}
