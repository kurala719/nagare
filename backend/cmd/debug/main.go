package main

import (
	"fmt"
	"log"

	"nagare/internal/database"
	"nagare/internal/service"
)

func main() {
	err := service.InitConfigServ("configs/nagare_config.json")
	if err != nil {
		log.Fatalf("InitConfigServ error: %v", err)
	}

	err = database.InitDBFromConfig()
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	fmt.Println("DB connection successful!")
}
