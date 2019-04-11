package errors

func New(code int, err error) *Error {
	return &Error{code, err.Error()}
}

type Error struct {
	Code int
	Text string
}

const (
	BindDataError   = 1410
	ValidationError = 1420
	PaginationError = 1430
	ServiceError    = 1510
	LimitationError = 1520
)

func (err *Error) Error() string {
	return err.Text
}
