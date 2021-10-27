package download

// DoneEvent ...
type DoneEvent struct{}

// ErrorEvent ...
type ErrorEvent struct {
	Payload error
}
