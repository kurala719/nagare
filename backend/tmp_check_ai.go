package main

import (
	"fmt"
	"os"

	"nagare/internal/database"
	"nagare/internal/repository/llm"
)

func main() {
	os.Setenv("NAGARE_CONFIG_PATH", "../configs/nagare_config.json")
	if err := database.InitDBFromConfig(); err != nil {
		fmt.Println("DB Init Error:", err)
		return
	}

	cfg := llm.Config{
		Type:   llm.ProviderOpenAI,
		APIKey: "test-key", // Usually we'd want the actual key from the DB, but let's query the DB first
	}

	fmt.Println("Done")
}
