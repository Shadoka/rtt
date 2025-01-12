package data

type ApplicationResult struct {
	AssertionError error
}

type ValidationResult struct {
	TestName        string
	ValidationError error
}
