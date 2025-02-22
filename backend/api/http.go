package api

type HttpError struct {
	Description string
	Code        int
}

func (e *HttpError) Error() string {
	return e.Description
}
