package data

type ConsumerResult struct {
	ConsumerQueue  string
	AssertionError error
}

type AssertionResult struct {
	Success          bool
	AssertionMessage string
}

type ValidationResult struct {
	TestName        string
	ValidationError error
}
