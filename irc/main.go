package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"regexp"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/charmbracelet/log"
	emoji "github.com/tmdvs/Go-Emoji-Utils"
)

var (
	IRC_ADDRESS  = GetEnv("IRC_ADDRESS", "127.0.0.1:6697")
	IRC_CHANNEL  = GetEnv("IRC_CHANNEL", "#mikogo")
	IRC_USERNAME = GetEnv("IRC_USERNAME", "mikogo")

	PRIVMSG_REGEXP = regexp.MustCompile(`^:(.+?)!~.+? PRIVMSG (#.+?) :(.+?)\r\n$`)
)

func handleMessage(conn net.Conn, username string, channel string, message string) {
	if channel != IRC_CHANNEL {
		return
	}

	log.Infof("> %s: %s", username, message)

	messageLower := strings.ToLower(message)

	if strings.HasPrefix(messageLower, "!") {
		if strings.HasPrefix(messageLower, "!testlong") {
			WriteToChannel(conn, channel, "ok trying to send a long message..")

			WriteToChannel(conn, channel, "yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn.")
		}

	} else if strings.Contains(strings.ToLower(messageLower), IRC_USERNAME) {
		gemmaResultChan := make(chan string)
		go Gemma3(username, message, gemmaResultChan)
		gemmaResult := <-gemmaResultChan

		if gemmaResult == "" {
			return
		}

		gemmaResult = strings.ToLower(emoji.RemoveAll(gemmaResult))

		log.Infof("gemma3: %s", gemmaResult)

		WriteToChannel(conn, channel, username+": "+gemmaResult)
	}
}

const (
	ConnStateConnecting = iota
	ConnStateConnected
	ConnStateDisconnected
)

var connState = ConnStateConnecting

func loop(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			connState = ConnStateDisconnected
			if err == io.EOF {
				log.Info("server closed connection")
				break
			} else {
				log.Info("error reading:", err.Error())
				break
			}
		}

		// fmt.Print(message)

		if connState == ConnStateConnecting {
			if strings.Contains(message, "001") || strings.Contains(message, "Welcome") {
				log.Info("connected to server!")
				connState = ConnStateConnected
			}
		}

		matches := PRIVMSG_REGEXP.FindStringSubmatch(message)
		if len(matches) == 0 {
			continue
		}

		username := matches[1]
		channel := matches[2]
		userMessage := matches[3]

		go handleMessage(conn, username, channel, userMessage)
	}
}

func main() {
	log.Infof("connecting to: %s", IRC_ADDRESS)

	conn, err := tls.Dial("tcp", IRC_ADDRESS, &tls.Config{
		InsecureSkipVerify: false,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	tcpConn, _ := conn.NetConn().(*net.TCPConn)
	tcpConn.SetKeepAlive(true)

	WritePrintf(conn, "NICK %s\r\n", IRC_USERNAME)
	WritePrintf(conn, "USER %s %s %s :Real Name\r\n",
		IRC_USERNAME, IRC_USERNAME, IRC_USERNAME)

	WritePrintf(conn, "JOIN %s\r\n", IRC_CHANNEL)

	go func() {
		for {
			// go routine ends early anyway when loop returns
			if connState == ConnStateDisconnected {
				break
			}
			WritePrintf(conn, "PING reee\r\n")
			time.Sleep(time.Second * 60)
		}
	}()

	loop(conn)

}
