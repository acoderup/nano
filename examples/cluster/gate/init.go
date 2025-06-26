package gate

import "github.com/acoderup/nano/component"

var (
	// All services in master server
	Services = &component.Components{}

	bindService = newBindService()
)

func init() {
	Services.Register(bindService)
}
