package common

import (
	"errors"
)

var (
	// --- 4xx Client Errors ---
	ErrBadRequest       = errors.New("bad request")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrNotPermitted     = errors.New("not permitted")
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("resource conflict")
	ErrValidationFailed = errors.New("validation failed")
	ErrRateLimited      = errors.New("rate limit exceeded")
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidToken     = errors.New("invalid token")
	ErrUnsupportedMedia = errors.New("unsupported media type")
	ErrTooManyRequests  = errors.New("too many requests")

	// --- 5xx Server Errors ---
	ErrInternal           = errors.New("internal server error")
	ErrSomethingWentWrong = errors.New("something went wrong")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrTimeout            = errors.New("request timeout")
	ErrDatabase           = errors.New("database error")
	ErrCache              = errors.New("cache error")
	ErrUpstream           = errors.New("upstream service error")
	ErrSerialization      = errors.New("serialization error")
	ErrDeserialization    = errors.New("deserialization error")

	// --- Auth / Access ---
	ErrNoAuthHeader      = errors.New("authorization header missing")
	ErrInsufficientScope = errors.New("insufficient oauth scope")

	// --- File / Upload ---
	ErrFileTooLarge    = errors.New("file size too large")
	ErrInvalidFileType = errors.New("invalid file type")
	ErrUploadFailed    = errors.New("file upload failed")

	// --- Payments / Business Logic ---
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrDuplicateRequest  = errors.New("duplicate request")
	ErrProcessingFailed  = errors.New("processing failed")
)
