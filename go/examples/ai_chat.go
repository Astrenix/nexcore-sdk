// NexCore Go SDK — Astrenix AI 对话(OpenAI 兼容协议)
//
// 运行:
//   cd sdk/go && go run examples/ai_chat.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	nexcore "github.com/nexcore-platform/nexcore-sdk-go"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	c := nexcore.NewClient(nexcore.Config{
		BaseURL:  envOr("NEXCORE_BASE_URL", "https://your-domain.com"),
		AIAPIKey: envOr("NEXCORE_AI_KEY", "sk-nc-xxx"),
	})

	raw, err := c.AI.Chat(
		[]nexcore.Message{
			{Role: "system", Content: "你是一个简洁的助手,回答不超过 2 句。"},
			{Role: "user", Content: "介绍一下 NexCore"},
		},
		"claude-opus-4-7",
		map[string]any{"temperature": 0.7, "max_tokens": 512},
	)
	if err != nil {
		if ne := nexcore.AsError(err); ne != nil {
			log.Fatalf("❌ Error #%d: %s (trace=%s)", ne.Code, ne.Message, ne.RequestID)
		}
		log.Fatal(err)
	}

	var reply struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &reply); err != nil {
		log.Fatal(err)
	}

	if len(reply.Choices) > 0 {
		fmt.Printf("🤖 Claude:\n%s\n\n", reply.Choices[0].Message.Content)
	}
	fmt.Printf("Usage: %d → %d tokens\n", reply.Usage.PromptTokens, reply.Usage.CompletionTokens)

	// 列出可用模型
	rawModels, err := c.AI.Models()
	if err == nil {
		var models struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(rawModels, &models)
		fmt.Println("\n可用模型:")
		for _, m := range models.Data {
			fmt.Printf("  - %s\n", m.ID)
		}
	}
}
