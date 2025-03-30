package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"slices"
	"strings"

	_ "github.com/joho/godotenv/autoload"

	"github.com/charmbracelet/log"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
)

var (
	DEBUG = getEnvExists("DEBUG")

	CHROMIUM_PATH = getEnv("CHROMIUM_PATH", "") // if empty, will download

	MUMBLE_USERNAME = getEnv("MUMBLE_USERNAME", "mikogo")
	MUMBLE_SERVER   = getEnv("MUMBLE_SERVER", "")
	MUMBLE_CHANNEL  = getEnv("MUMBLE_CHANNEL", "")

	JOIN_IGNORE_USERS = []string{"tesutogo"}
)

func handleTextMessage(e *gumble.TextMessageEvent, msg string, browser *rod.Browser) {
	if msg == "test" {
		wordArtImg, err := makeWordArtPng(browser, e.Sender.Name)
		if err != nil {
			e.Sender.Channel.Send(err.Error(), false)
		}

		html, err := imageForMumble(wordArtImg, &MumbleImageOptions{
			Transparent: true,
			MaxHeight:   100,
		})

		if err != nil {
			e.Sender.Channel.Send(err.Error(), false)
		}

		e.Sender.Channel.Send(html, false)

		// sendToAll(e.Client, html)
	}
}

func handleUserConnected(e *gumble.UserChangeEvent, browser *rod.Browser) {
	wordArtImg, err := makeWordArtPng(browser, e.User.Name)
	if err != nil {
		return
	}

	html, err := imageForMumble(wordArtImg, &MumbleImageOptions{
		Transparent: true,
		MaxHeight:   100,
	})

	if err != nil {
		return
	}

	sendToAll(e.Client, html)
}

func main() {
	if DEBUG {
		log.SetLevel(log.DebugLevel)
	}

	if MUMBLE_SERVER == "" {
		log.Fatal("please specify MUMBLE_SERVER")
	}

	log.Info("initializing browser...")

	browserLauncher, err := launcher.New().
		Headless(true).
		Bin(CHROMIUM_PATH).
		Launch()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	browser := rod.New().ControlURL(browserLauncher)

	err = browser.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Info("connecting to mumble...")

	keepAlive := make(chan bool)

	config := gumble.NewConfig()

	config.Username = MUMBLE_USERNAME
	config.Attach(gumbleutil.AutoBitrate)

	config.Attach(gumbleutil.Listener{
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

			e.Client.Self.SetSelfDeafened(true)
		},

		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}

			msg := strings.TrimSpace(e.Message)

			log.Infof("%s: %s", e.Sender.Name, msg)

			go handleTextMessage(e, msg, browser)
		},

		UserChange: func(e *gumble.UserChangeEvent) {
			if e.Type.Has(gumble.UserChangeConnected) {
				log.Infof("%s %s", e.User.Name, "joined")

				if slices.Contains(JOIN_IGNORE_USERS, strings.ToLower(e.User.Name)) {
					return
				}

				go handleUserConnected(e, browser)
			}
		},

		Disconnect: func(e *gumble.DisconnectEvent) {
			keepAlive <- true
		},
	})

	_, err = gumble.DialWithDialer(new(net.Dialer), MUMBLE_SERVER, config, &tls.Config{})
	if err != nil {
		log.Fatal(err)
	}

	<-keepAlive

}
