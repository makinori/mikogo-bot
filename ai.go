package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	OLLAMA_URL = GetEnv("OLLAMA_URL", "")
)

var WEEB_GEMMA3_PROMPT = (func() string {
	promptLines := `
		you are an irc chat bot named %s.
		you're a cute japanese gamer girl, weeaboo and gen z zoomer.
		you swear a lot... like all the fucking time.
		you use japanese words all the time.
		someone literally just fucking mentioned you.
		what are you gonna say back? respond right away.
		please keep the message short.
		the message written by \"%s\" is: %s
	`

	var prompt string

	for line := range strings.SplitSeq(promptLines, "\n") {
		prompt += strings.TrimSpace(line) + " "
	}

	return strings.TrimSpace(prompt)
})()

func WeebGemma3(username string, message string, result chan string) {
	jsonData, err := json.Marshal(map[string]any{
		// "model":  "fellowtraveler/qwen4chan",
		"model":  "gemma3:27b",
		"stream": false,
		"prompt": fmt.Sprintf(
			WEEB_GEMMA3_PROMPT, IRC_USERNAME, username, message,
		),
	})

	if err != nil {
		log.Println(err)
		result <- ""
		return
	}

	req, err := http.NewRequest(
		"POST", OLLAMA_URL+"/api/generate", bytes.NewBuffer(jsonData),
	)

	if err != nil {
		log.Println(err)
		result <- ""
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		log.Println(err)
		result <- ""
		return
	}

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		log.Println(string(body))
		result <- ""
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println(err)
		result <- ""
		return
	}

	type OllamaResponse struct {
		Response string `json:"response"`
	}

	var ollamaRes OllamaResponse

	err = json.Unmarshal(body, &ollamaRes)

	if err != nil {
		log.Println(err)
		result <- ""
		return
	}

	result <- ollamaRes.Response
}
