package main

import (
	"github.com/acoderup/core/logger"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/acoderup/nano"
	"github.com/acoderup/nano/cluster"
	"github.com/acoderup/nano/examples/cluster/chat"
	"github.com/acoderup/nano/examples/cluster/gate"
	"github.com/acoderup/nano/examples/cluster/master"
	"github.com/acoderup/nano/serialize/json"
	"github.com/acoderup/nano/session"
	"github.com/pingcap/errors"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "NanoClusterDemo"
	app.Author = "Lonng"
	app.Email = "heng@acoderup.org"
	app.Description = "Nano cluster demo"
	app.Commands = []cli.Command{
		{
			Name: "master",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "listen,l",
					Usage: "Master service listen address",
					Value: "127.0.0.1:34567",
				},
			},
			Action: runMaster,
		},
		{
			Name: "gate",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "master",
					Usage: "master server address",
					Value: "127.0.0.1:34567",
				},
				cli.StringFlag{
					Name:  "listen,l",
					Usage: "Gate service listen address",
					Value: "",
				},
				cli.StringFlag{
					Name:  "gate-address",
					Usage: "Client connect address",
					Value: "",
				},
			},
			Action: runGate,
		},
		{
			Name: "chat",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "master",
					Usage: "master server address",
					Value: "127.0.0.1:34567",
				},
				cli.StringFlag{
					Name:  "listen,l",
					Usage: "Chat service listen address",
					Value: "",
				},
			},
			Action: runChat,
		},
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Startup server error %+v", err)
	}
}

func srcPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func runMaster(args *cli.Context) error {
	listen := args.String("listen")
	if listen == "" {
		return errors.Errorf("master listen address cannot empty")
	}

	webDir := filepath.Join(srcPath(), "master", "web")
	logger.Logger.Tracef("Nano master server web content directory [%v]", webDir)
	logger.Logger.Tracef("Nano master listen address [%v]", listen)
	logger.Logger.Tracef("Open http://127.0.0.1:12345/web/ in browser")

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir(webDir))))
	go func() {
		if err := http.ListenAndServe(":12345", nil); err != nil {
			panic(err)
		}
	}()

	// Register session closed callback
	session.Lifetime.OnClosed(master.OnSessionClosed)

	// Startup Nano server with the specified listen address
	nano.Listen(listen,
		nano.WithMaster(),
		nano.WithComponents(master.Services),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithDebugMode(),
		nano.WithUnregisterCallback(func(m cluster.Member) {
			logger.Logger.Tracef("Todo alarm unregister: [%v]", m.String())
		}),
	)

	return nil
}

func runGate(args *cli.Context) error {
	listen := args.String("listen")
	if listen == "" {
		return errors.Errorf("gate listen address cannot empty")
	}

	masterAddr := args.String("master")
	if masterAddr == "" {
		return errors.Errorf("master address cannot empty")
	}

	gateAddr := args.String("gate-address")
	if gateAddr == "" {
		return errors.Errorf("gate address cannot empty")
	}

	logger.Logger.Tracef("Current server listen address [%v]", listen)
	logger.Logger.Tracef("Current gate server address [%v]", gateAddr)
	logger.Logger.Tracef("Remote master server address [%v]", masterAddr)

	// Startup Nano server with the specified listen address
	nano.Listen(listen,
		nano.WithAdvertiseAddr(masterAddr),
		nano.WithClientAddr(gateAddr),
		nano.WithComponents(gate.Services),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithIsWebsocket(true),
		nano.WithWSPath("/nano"),
		nano.WithCheckOriginFunc(func(_ *http.Request) bool { return true }),
		nano.WithDebugMode(),
		nano.WithNodeId(2), // if you deploy multi gate, option set nodeId, default nodeId = os.Getpid()
	)
	return nil
}

func runChat(args *cli.Context) error {
	listen := args.String("listen")
	if listen == "" {
		return errors.Errorf("chat listen address cannot empty")
	}

	masterAddr := args.String("master")
	if masterAddr == "" {
		return errors.Errorf("master address cannot empty")
	}

	logger.Logger.Tracef("Current chat server listen address [%v]", listen)
	logger.Logger.Tracef("Remote master server address [%v]", masterAddr)

	// Register session closed callback
	session.Lifetime.OnClosed(chat.OnSessionClosed)

	// Startup Nano server with the specified listen address
	nano.Listen(listen,
		nano.WithAdvertiseAddr(masterAddr),
		nano.WithComponents(chat.Services),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithDebugMode(),
	)

	return nil
}
