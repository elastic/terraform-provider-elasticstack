package config

// TestStepConfigFunc is a function that returns a config string for a test step.
type TestStepConfigFunc func(TestStepConfigRequest) string

// TestStepConfigRequest holds information about the test step.
type TestStepConfigRequest struct{}

// TestNameDirectory returns a function that returns the test name directory.
func TestNameDirectory() TestStepConfigFunc {
	return func(_ TestStepConfigRequest) string { return "" }
}
