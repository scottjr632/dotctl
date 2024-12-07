package result

import "errors"

type Result[T any] struct {
	value T
	err   error
}

// NewResult creates a new Result[T] with the given value and error
func NewResult[T any](value T, err error) Result[T] {
	return Result[T]{value, err}
}

// New creates a new Result[T] with the given value and error
func New[T any](value T, err error) Result[T] {
	return NewResult(value, err)
}

// Ok creates a new Result[T] with the given value and no error
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a new Result[T] with the given error and no value
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

func (r Result[T]) Value() T {
	return r.value
}

func (r Result[T]) UnwrapErr() error {
	return r.err
}

func (r Result[T]) IsOk() bool {
	return r.err == nil
}

func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Must unraps the result and panics if there is an error
func (r Result[T]) Must() T {
	if r.err != nil {
		panic(resultError{r.err})
	}
	return r.value
}

// Unwrap returns the value and error
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

type resultError struct {
	error
}

// Recover recovers a panicked Result[T] and returns the error
// to the caller. This call **must** be used in a defer statement
func Recover[T any](res *Result[T]) {
	if r := recover(); r != nil {
		err, ok := r.(resultError)
		if !ok {
			panic(r)
		}
		*res = Result[T]{err: err.error}
	}
}

func (r Result[T]) RecoverToError(callignErr *error) {
	if r := recover(); r != nil {
		err, ok := r.(resultError)
		if !ok {
			panic(r)
		}
		*callignErr = err.error
	}
}

// MustUnwrapErr returns the Err value or panics if there is no error.
func (r Result[T]) MustUnwrapErr() error {
	if r.err == nil {
		panic("expected the result to contain error")
	}
	return r.err
}

// Example: taken from https://github.com/olevski/eh/blob/main/eh.go
//
//	func main() {
//		var WhenGetFromDBError error
//		var WhenGetFromRemoteError error
//		var WhenGetFromInMemoryError error
//		func Example() (r eh.Result[string]) {
//				// Escape if error is not handled
//				defer eh.EscapeHatch(&r)
//				// Run error handler when previous `defer` rethrows error
//				defer eh.CatchError(&r, func(_ error) {
//					return eh.NewResult(GetFromRemote()).Eh()
//				}, WhenGetFromDBError)
//				// Run error handler when error is `WhenGetFromInMemoryError`
//				defer eh.CatchError(&r, func(_ error) {
//					return eh.NewResult(GetFromDb()).Eh()
//				}, WhenGetFromInMemoryError)
//				successVal := eh.NewResult(GetFromInMemory()).Eh()
//				return eh.Result[string]{Ok: successVal}
//			}
//	}
func CatchError[T any](res *Result[T], catcher func(error) T, when ...error) {
	defer func() {
		if res.IsOk() {
			return
		}

		err := res.MustUnwrapErr()
		if when == nil {
			*res = Result[T]{value: catcher(err)}
			return
		}
		for _, target := range when {
			if !errors.Is(err, target) {
				continue
			}
			*res = Result[T]{value: catcher(err)}
			break
		}
	}()
	defer Recover(res)
	if r := recover(); r != nil {
		panic(r)
	}
}
