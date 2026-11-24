package omnigraph

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrBadRequest           = errors.New("bad request")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrNotFound             = errors.New("not found")
	ErrMethodNotAllowed     = errors.New("method not allowed")
	ErrRequestTimeout       = errors.New("request timeout")
	ErrConflict             = errors.New("conflict")
	ErrGone                 = errors.New("gone")
	ErrPreconditionFailed   = errors.New("precondition failed")
	ErrPayloadTooLarge      = errors.New("payload too large")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	ErrRangeNotSatisfiable  = errors.New("range not satisfiable")
	ErrFailedDependency     = errors.New("failed dependency")
	ErrTooManyRequests      = errors.New("too many requests")
	ErrInternalServer       = errors.New("internal server error")
	ErrServiceUnavailable   = errors.New("service unavailable")
	ErrNetwork              = errors.New("network error")
	ErrConfiguration        = errors.New("configuration error")
	ErrAPIContract          = errors.New("api contract mismatch")
	ErrGraphUnavailable     = errors.New("graph unavailable")
)

type Error struct {
	Status            int
	Code              string
	Message           string
	RequestID         string
	Method            string
	URL               string
	RequestDispatched bool
	OutcomeUnknown    bool
	Disposition       *DispositionOutput
	RetryAfter        string
	Body              *ErrorOutput
	Err               error
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("omnigraph: %s (%d %s): %s", e.Method, e.Status, e.Code, e.Message)
	}
	if e.Status > 0 {
		return fmt.Sprintf("omnigraph: %s (%d): %s", e.Method, e.Status, e.Message)
	}
	return fmt.Sprintf("omnigraph: %s: %s", e.Method, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func sentinelForStatus(status int, code string) error {
	if code == "graph_unavailable" || code == "graph_startup_failed" {
		return ErrGraphUnavailable
	}
	switch status {
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusMethodNotAllowed:
		return ErrMethodNotAllowed
	case http.StatusRequestTimeout:
		return ErrRequestTimeout
	case http.StatusConflict:
		return ErrConflict
	case http.StatusGone:
		return ErrGone
	case http.StatusPreconditionFailed:
		return ErrPreconditionFailed
	case http.StatusRequestEntityTooLarge:
		return ErrPayloadTooLarge
	case http.StatusUnsupportedMediaType:
		return ErrUnsupportedMediaType
	case http.StatusRequestedRangeNotSatisfiable:
		return ErrRangeNotSatisfiable
	case http.StatusFailedDependency:
		return ErrFailedDependency
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	default:
		if status >= 500 {
			return ErrInternalServer
		}
		return nil
	}
}
