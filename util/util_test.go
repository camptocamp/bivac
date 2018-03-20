package util

import (
	"errors"
)

func ExampleCheckErr_nil() {
	CheckErr(nil, "test", "fatal")
	// Output:
}

func ExampleCheckErr_noExit() {
	fakeErr := errors.New("Fake error")
	CheckErr(fakeErr, "test: %v", "error")
	// Output:
}

func ExampleCheckErr_exit() {
	// fakeErr := errors.New("Fake error")
	// How do we test the output and the os.Exit(1)?"
	// CheckErr(fakeErr, "test: %v", "fatal")
}
