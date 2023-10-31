package controllers

import (
	"context"
	"fmt"
	"net/http"
	"ppdb_sekolah_go/constans"

	"github.com/labstack/echo/v4"
	openai "github.com/sashabaranov/go-openai"
)

func AIController(c echo.Context) error {
	query := c.QueryParam("tanya")
	client := openai.NewClient("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")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return err
	}

	// Add the global variable to the response map
	responseMap := map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success get all users",
		constans.DATA:    resp.Choices[0].Message.Content,
	}

	// Return the JSON response
	return c.JSON(http.StatusOK, responseMap)
}
