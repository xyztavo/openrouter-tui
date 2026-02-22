package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	openrouter "github.com/revrost/go-openrouter"
)

func getToken() (token string, err error) {
	configDir, _ := os.UserConfigDir()
	appDir := filepath.Join(configDir, "chatui")
	envPath := filepath.Join(appDir, ".env")
	os.MkdirAll(appDir, 0755)

	data, err := os.ReadFile(envPath)
	if err != nil {
		os.WriteFile(envPath, []byte(""), 0644)
	}
	tokenFromEnv := strings.TrimSpace(string(data))
	if tokenFromEnv != "" {
		return tokenFromEnv, nil
	}

	var tokenFromUser string
	fmt.Print("Digite seu token: ")
	fmt.Scanln(&tokenFromUser)
	err = os.WriteFile(envPath, []byte(tokenFromUser), 0644)
	if err != nil {
		return tokenFromUser, err
	}
	fmt.Printf("token from user: %s\n", tokenFromUser)
	return tokenFromUser, nil
}
func getMessage(messages []openrouter.ChatCompletionMessage) string {
	reader := bufio.NewScanner(os.Stdin)
	if len(messages) == 0 {
		fmt.Print(" > ")
	} else {
		fmt.Print("\n > ")
	}

	if reader.Scan() {
		return strings.TrimSpace(reader.Text())
	}
	return ""
}

func main() {
	token, err := getToken()
	if err != nil {
		log.Fatal(err.Error())
	}
	client := openrouter.NewClient(token)

	systemMsg := openrouter.SystemMessage("you are a chill and cool terminal chatbot.")
	ctx := context.Background()
	var messages []openrouter.ChatCompletionMessage
	var renderedOutputs []string

	for {
		var message string
		message = getMessage(messages)
		messages = append(messages, openrouter.UserMessage(message))
		fullMessages := []openrouter.ChatCompletionMessage{systemMsg}
		fullMessages = append(fullMessages, messages...)
		stream, err := client.CreateChatCompletionStream(
			ctx, openrouter.ChatCompletionRequest{
				Model:    "google/gemini-2.5-flash-lite",
				Messages: fullMessages,
				Stream:   true,
			},
		)
		if err != nil {
			log.Fatal(err.Error())
		}

		var fullAssistantResponse strings.Builder
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err.Error())
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				fullAssistantResponse.WriteString(content)
			}
		}
		assistantText := fullAssistantResponse.String()

		if len(renderedOutputs) > 0 {
			fmt.Print("\n--------\n\n")
		}
		rendered, err := glamour.Render(assistantText, "light")
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Print(rendered)
		renderedOutputs = append(renderedOutputs, rendered)

		fmt.Println()
		messages = append(messages, openrouter.AssistantMessage(assistantText))
	}
}
