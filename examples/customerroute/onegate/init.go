package onegate

import "github.com/acoderup/nano/component"

var (
	// All services in master server
	Services = &component.Components{}

	bindService = newRegisterService()
)

func init() {
	Services.Register(bindService)
}
