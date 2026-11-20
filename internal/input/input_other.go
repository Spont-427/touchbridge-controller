//go:build !windows && !linux

package input

func New() (Backend, error) { return nil, ErrUnsupported }
