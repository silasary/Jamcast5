package main

import (
	"io"
	"log"
	"os"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"github.com/getlantern/systray"

	"gitlab.com/redpointgames/jamcast/client/shutdown"
	"gitlab.com/redpointgames/jamcast/client/window/logs"
	"gitlab.com/redpointgames/jamcast/image"
	"gitlab.com/redpointgames/jamcast/network"
)

var clientApp fyne.App
var systrayReady chan bool

const enableSystray = true

func main() {
	if len(os.Args) == 3 && os.Args[1] == "--self-update" {
		finishSelfUpdate(os.Args[2])
		return
	}

	// split logs between stdout and our internal logging
	log.SetOutput(io.MultiWriter(logs.GetInMemoryLogBuffer(), os.Stdout))

	shutdown.SetupShutdownGlobalHandler()

	log.Println("currently version ", version)

	log.Println("performing self update check")

	doSelfUpdate()

	log.Println("registering with ZeroTier network")

	_ = network.Start()

	log.Println("starting JamCast")

	clientApp = app.New()
	clientApp.SetIcon(fyne.NewStaticResource("icon.png", image.IconPNG))

	shutdown.RegisterShutdownHandler(func() {
		clientApp.Quit()
	})

	if enableSystray {
		systrayReady = make(chan bool)
		go func() {
			systray.Run(func() {
				systray.SetIcon(image.Icon)
				systray.SetTitle("JamCast")
				systray.SetTooltip("JamCast - Not signed in")

				shutdown.RegisterShutdownHandler(func() {
					systray.Quit()
				})

				systrayReady <- true
			}, func() {})
		}()
	}

	go func() {
		stageIntro()
	}()

	go func() {
		for {
			doSelfUpdate()
			time.Sleep(time.Minute * 5)
		}
	}()

	for {
		clientApp.Run()

		clientApp = app.New()
		clientApp.SetIcon(fyne.NewStaticResource("icon.png", image.IconPNG))
	}

	if enableSystray {
		systray.Quit()
	}
}
