package data

// Parsing errors.
const (
	ErrBadFixedDuration strErr = "must be RFC 3339 duration in range week to fraction"
)

type strErr string

func (err strErr) Error() string { return string(err) }
