package main

import (
	"fmt"
	"log"
	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/service"
)

func main() {
	if err := service.InitConfigServ("configs/nagare_config.json"); err != nil {
		log.Fatal(err)
	}
	if err := database.InitDBFromConfig(); err != nil {
		log.Fatal(err)
	}
	defer database.CloseDB(database.DB)

	fmt.Println("Running AutoMigrate for model.Host...")
	if err := database.DB.AutoMigrate(&model.Host{}); err != nil {
		fmt.Printf("AutoMigrate failed: %v\n", err)
	} else {
		fmt.Println("AutoMigrate successful.")
	}

	var columns []struct {
		Field string
		Type  string
	}
	if err := database.DB.Raw("DESCRIBE hosts").Scan(&columns).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Println("Columns in 'hosts' table after migration:")
	for _, c := range columns {
		fmt.Printf("- %s (%s)\n", c.Field, c.Type)
	}
}
