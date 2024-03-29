package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func (p *Plugin) getSummaryForPosts(posts map[string]*model.Post) string {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", p.configuration.GeminiAPIKey)

	chatPrompt := "Summarize the below conversation: "

	for _, post := range posts {
		chatPrompt += fmt.Sprintf("%s: %s\n", post.UserId, post.Message)
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": chatPrompt,
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return ""
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return ""
	}
	defer resp.Body.Close()

	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return ""
	}
	candidates := responseData["candidates"].([]interface{})
	if len(candidates) > 0 {
		content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
		parts := content["parts"].([]interface{})
		if len(parts) > 0 {
			text := parts[0].(map[string]interface{})["text"].(string)
			return text
		} else {
			fmt.Println("No 'text' found in 'parts' array")
		}
	} else {
		fmt.Println("No 'candidates' found in response")
	}
	byteArray, _ := io.ReadAll(resp.Body)
	return bytes.NewBuffer(byteArray).String()
}

func (p *Plugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	if post.UserId == p.botID {
		return
	}

	if !strings.Contains(post.Message, fmt.Sprintf("@%s", "summary-ai")) {
		return
	}

	rootPost := post.RootId

	postThread, err := p.API.GetPostThread(rootPost)
	if err != nil {
		return
	}

	msg := p.getSummaryForPosts(postThread.Posts)

	_, _ = p.API.CreatePost(&model.Post{
		UserId:    p.botID,
		ChannelId: post.ChannelId,
		RootId:    rootPost,
		Message:   msg,
	})
}
