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
		// fmt.Println(line)

		if strings.TrimSpace(line) == "" {
			continue
		}

		privmsgFormat := "PRIVMSG %s :%s\r\n"

		overhead := len(fmt.Sprintf(privmsgFormat, channel, ""))
		// should be 512 bytes but there's more overhead so just -64 extra
		maxLength := (512 - 64) - overhead

		splitLines := SplitStringBySpace(line, maxLength)
		// splitLinesLen := len(splitLines)

		for _, splitLine := range splitLines {
			// if splitLinesLen > 1 {
			// 	splitLine = fmt.Sprintf("[%d/%d] %s", i+1, splitLinesLen, splitLine)
			// }
			WriteSprintf(conn, privmsgFormat, channel, splitLine)
		}
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

func SplitString(line string, x int) []string {
	if x <= 0 {
		return []string{line}
	}

	var result []string

	for i := 0; i < len(line); i += x {
		end := min(i+x, len(line))
		result = append(result, line[i:end])
	}

	return result
}

// split string every x but split early by spaces
func SplitStringBySpace(line string, x int) []string {
	if x <= 0 {
		return []string{line}
	}

	var result []string

	lastSpace := 0
	lastSplit := 0

	for i := 1; i < len(line); i += 1 {
		if line[i] == ' ' {
			lastSpace = i
		}

		if (i-lastSplit)%x == 0 {
			result = append(result, line[lastSplit:lastSpace])

			i = lastSpace + 1
			lastSplit = i
		}
	}

	result = append(result, line[lastSplit:])

	return result
}
