package imports

import (
	"github.com/acoderup/core"
	_ "github.com/acoderup/core/basic"
	_ "github.com/acoderup/core/cmdline"
	"github.com/acoderup/core/logger"
	_ "github.com/acoderup/core/logger"
	"github.com/acoderup/core/module"
	_ "github.com/acoderup/core/signal"
	_ "github.com/acoderup/core/task"
	_ "github.com/acoderup/core/timer"
	"github.com/acoderup/nano"
	"github.com/acoderup/nano/component"
)

type Base struct {
	component.Base
}

func (s *Base) Init() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Errorf("Recovered from panic: %v\n", err)
			}
		}()
		defer core.ClosePackages()
		core.LoadPackagesAuto()

		waiter := module.Start()
		waiter.Wait("main")
	}()
}
func init() {
	nano.ServicesComponents.Register(&Base{})
}
