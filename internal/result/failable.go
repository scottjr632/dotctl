package result

// Failable is a result that can be used to check if an error occurred
// and does not hold any value
type Failable struct {
	err error
}

func NewFailable(err error) Failable {
	return Failable{err}
}

func (f Failable) IsOk() bool {
	return f.err == nil
}

func (f Failable) IsErr() bool {
	return f.err != nil
}

func (f Failable) Err() error {
	return f.err
}

// Must unraps the result and panics if there is an error
// can be caughts with a defer result.Recover(&err)
func (f Failable) Must() {
	if f.err != nil {
		panic(resultError{f.err})
	}
}
