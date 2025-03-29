package main

import (
	_ "github.com/joho/godotenv/autoload"

	"github.com/charmbracelet/log"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
)

var (
	DEBUG = getEnvExists("DEBUG")

	MUMBLE_CHANNEL = getEnv("MUMBLE_CHANNEL", "")
)

// func handleTextMessage(e *gumble.TextMessageEvent, msg string, browser *rod.Browser) {
// 	if e.Sender.Name != "Maki" {
// 		return
// 	}

// 	if strings.HasPrefix(msg, "test") {
// 		// if len(msg) <= 5 {
// 		// 	return
// 		// }

// 	}
// }

func handleUserConnected(e *gumble.UserChangeEvent, browser *rod.Browser) {
	wordArtImg := makeWordArtPng(browser, e.User.Name)

	html := imageForMumble(wordArtImg, &MumbleImageOptions{
		Transparent: true,
		MaxHeight:   100,
	})

	sendToAll(e.Client, html)
}

func main() {
	if DEBUG {
		log.SetLevel(log.DebugLevel)
	}

	browserLauncher := launcher.New().Headless(true).MustLaunch()
	browser := rod.New().ControlURL(browserLauncher).MustConnect()

	// TODO: pass username and server through env?

	gumbleutil.Main(gumbleutil.AutoBitrate, gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			log.Infof("connected as: %s", e.Client.Self.Name)

			var foundChannel *gumble.Channel

			for _, channel := range e.Client.Channels {
				if channel.Name == MUMBLE_CHANNEL {
					foundChannel = channel
					break
				}
			}

			if foundChannel == nil {
				return
			}

			e.Client.Self.Move(foundChannel)
			log.Infof("moved to: %s", foundChannel.Name)
		},

		// TextMessage: func(e *gumble.TextMessageEvent) {
		// 	if e.Sender == nil {
		// 		return
		// 	}

		// 	msg := strings.TrimSpace(e.Message)

		// 	log.Infof("%s: %s", e.Sender.Name, msg)

		// 	go handleTextMessage(e, msg, browser)
		// },

		UserChange: func(e *gumble.UserChangeEvent) {
			if e.Type.Has(gumble.UserChangeConnected) {
				log.Infof("%s %s", e.User.Name, "joined")
				go handleUserConnected(e, browser)
			}
		},
	})

}
