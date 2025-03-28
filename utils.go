package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func WriteSprintf(conn net.Conn, format string, a ...any) (int, error) {
	return conn.Write(fmt.Appendf([]byte{}, format, a...))
}

func WriteToChannel(conn net.Conn, channel string, message string) {
	lines := strings.SplitSeq(message, "\n")

	for line := range lines {
		fmt.Println(line)

		if strings.TrimSpace(line) == "" {
			continue
		}

		// TODO: seperate 512 bytes by \n

		WriteSprintf(conn, "PRIVMSG %s :%s\r\n", channel, line)
	}
}

func GetEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	} else {
		return fallback
	}
}
