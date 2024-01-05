package gerror

import "errors"

var ErrKeyNotFound = errors.New("entry not found in cache")

var (
	ErrEvictionPolicyNotFound = errors.New("the given eviction policy name provided was not found")
	ErrNotAGcacheCommand      = errors.New("command should be an Array frame")
	ErrNotGcacheCmd           = errors.New("this frame is not a gcache command")
	ErrInvalidPingCommand     = errors.New("ping command is malformed")

	ErrInvalidCmdName = errors.New("command not found")
)
