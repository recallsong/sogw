package core

import "errors"

var (
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrApiValidateFailed  = errors.New("fail to validate api")
	ErrRouteNotFound      = errors.New("route not found")
	ErrMethodNotAllow     = errors.New("method not allow")
	ErrHostNotAllow       = errors.New("host not allow")
	ErrAuthFailed         = errors.New("unauthorized")
)
