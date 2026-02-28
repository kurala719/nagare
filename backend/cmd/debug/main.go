package main

import (
	"fmt"
	"log"

	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/service"
)

func main() {
	err := service.InitConfigServ("configs/nagare_config.json")
	if err != nil {
		log.Fatalf("InitConfigServ error: %v", err)
	}

	err = database.InitDBFromConfig()
	if err != nil {
		log.Fatalf("InitDB error: %v", err)
	}

	var items []model.Item
	database.DB.Limit(20).Find(&items)
	for _, it := range items {
		fmt.Printf("Item ID: %d, HID: %d, Name: %s\n", it.ID, it.HID, it.Name)
	}
}
