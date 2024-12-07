package promise

type AsyncResult[T any] struct {
	Value T
	Error error
}

type Promise[T any] struct {
	ch chan AsyncResult[T]
}

func New[T any](fn func() (T, error)) *Promise[T] {
	ch := make(chan AsyncResult[T])
	go func() {
		defer close(ch)
		v, err := fn()
		ch <- AsyncResult[T]{Value: v, Error: err}
	}()
	return &Promise[T]{ch: ch}
}

func (p *Promise[T]) Await() (T, error) {
	result := <-p.ch
	return result.Value, result.Error
}
