package main

import (
	"fmt"
	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/service"
)

func main() {
	// Initialize config for DB connection
	service.InitConfigServ("configs/nagare_config.json")
	database.InitDBFromConfig()

	fmt.Println("DB connection successful!")

	// Inspect HostHistory columns
	rows, err := database.DB.Raw("DESCRIBE host_histories").Rows()
	if err == nil {
		fmt.Println("COLUMNS_START")
		var field, typ, null, key, def, extra string
		for rows.Next() {
			rows.Scan(&field, &typ, &null, &key, &def, &extra)
			fmt.Printf("COLUMN: %s\n", field)
		}
		rows.Close()
		fmt.Println("COLUMNS_END")
	} else {
		fmt.Printf("Failed to describe host_histories: %v\n", err)
	}

	var count int64
	database.DB.Model(&model.Item{}).Count(&count)
	fmt.Printf("TOTAL ITEMS IN DB: %d\n", count)

	var hosts []model.Host
	database.DB.Find(&hosts)
	for _, h := range hosts {
		var c int64
		database.DB.Model(&model.Item{}).Where("hid = ?", h.ID).Count(&c)
		fmt.Printf("Host ID %d: %s (Monitor: %d, Items: %d)\n", h.ID, h.Name, h.MonitorID, c)

		// Check history
		var history []model.HostHistory
		database.DB.Where("host_id = ?", h.ID).Order("sampled_at desc").Limit(20).Find(&history)
		for _, hist := range history {
			fmt.Printf("   -> History ID %d: At %s: Tot=%d, Act=%d\n", hist.ID, hist.SampledAt.Format("15:04:05"), hist.ItemTotal, hist.ItemActive)
		}
	}
}
