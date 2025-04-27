package balancer

type Balancer interface {
	GetNextBackend() *Backend
}
