package utils

import "fmt"

func Invariant(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

// Invariantf is like Invariant but with a formatted message
func Invariantf(cond bool, msg string, args ...interface{}) {
	if !cond {
		panic(fmt.Sprintf(msg, args...))
	}
}

func InvariantErr(cond bool, msg string) error {
	if !cond {
		return fmt.Errorf(msg)
	}
	return nil
}
