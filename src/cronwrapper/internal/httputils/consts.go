package httputils

import "time"

const (
	IdleConnTimeout = 10 * time.Second
	DialTimeout     = 2 * time.Second
	RequestTimeout  = 5 * time.Second
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 30 * time.Second
)
