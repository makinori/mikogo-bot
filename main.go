package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"

	_ "github.com/joho/godotenv/autoload"

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

	log.Printf("%s: %s", username, message)

	messageLower := strings.ToLower(message)

	if strings.HasPrefix(messageLower, "!") {
		if strings.HasPrefix(messageLower, "!testlong") {
			WriteToChannel(conn, channel, "ok trying to send a long message..")

			WriteToChannel(conn, channel, "yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn. yo yo, what's up gamer? how are you? hell yeah lets go, this is working out great. let's gettinnn.")
		}

	} else if strings.Contains(strings.ToLower(messageLower), IRC_USERNAME) {
		gemmaResultChan := make(chan string)
		go WeebGemma3(username, message, gemmaResultChan)
		gemmaResult := <-gemmaResultChan

		if gemmaResult == "" {
			return
		}

		WriteToChannel(
			conn, channel,
			username+": "+strings.ToLower(emoji.RemoveAll(gemmaResult)),
		)
	}
}

func loop(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Server closed connection")
				break
			} else {
				log.Panicln("Error reading:", err.Error())
				break
			}
		}

		fmt.Print(message)

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
	conn, err := tls.Dial("tcp", IRC_ADDRESS, &tls.Config{
		InsecureSkipVerify: false,
	})

	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	WriteSprintf(conn, "NICK %s\r\n", IRC_USERNAME)
	WriteSprintf(conn, "USER %s %s %s :Real Name\r\n",
		IRC_USERNAME, IRC_USERNAME, IRC_USERNAME)

	WriteSprintf(conn, "JOIN %s\r\n", IRC_CHANNEL)

	loop(conn)

}
