package pacemaker

func NewNotFoundErr(msg string) error {
	return &NotFoundObject{msg}
}

type NotFoundObject struct {
	msg string
}

func (err *NotFoundObject) Error() string {
	return err.msg
}

func NewCibError(msg string) error {
	return &CibError{msg}
}

type CibError struct {
	msg string
}

func (e *CibError) Error() string {
	return e.msg
}

func NewAlreadyExistedErr(msg string) error {
	return &ConnectionErr{msg}
}

type AlreadyExistedErr struct {
	msg string
}

func (err *AlreadyExistedErr) Error() string {
	return err.msg
}

func NewConnectionErr(msg string) error {
	return &ConnectionErr{msg}
}

type ConnectionErr struct {
	msg string
}

func (err *ConnectionErr) Error() string {
	return err.msg
}

func NewNotSupportedOpErr(msg string) error {
	return &NotSupportedOpErr{msg}
}

type NotSupportedOpErr struct {
	msg string
}

func (err *NotSupportedOpErr) Error() string {
	return err.msg
}
