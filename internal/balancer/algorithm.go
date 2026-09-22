package balancer

import "lb/internal/backend"

type Algorithm interface {
	Next() (*backend.Backend, error)
}
