package transport

import "context"

type Server interface {
	Start(context.Context) error
	Close(context.Context) error
}
