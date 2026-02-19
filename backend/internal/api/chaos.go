package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/service"
)

// TriggerAlertStormCtrl simulates a large number of alerts incoming
func TriggerAlertStormCtrl(c *gin.Context) {
	go func() {
		messages := []string{
			"CPU usage exceeds 95%% on host %d",
			"Memory leak detected in service %d",
			"Disk space critical (99%%) on /dev/sda%d",
			"Network latency spike detected on interface eth%d",
			"Connection timeout to database cluster %d",
			"SSH login failure from unknown IP to node %d",
			"Fan failure detected in chassis %d",
			"High temperature warning on core %d",
		}

		// Generate 200 alerts over a few seconds
		for i := 0; i < 200; i++ {
			msg := fmt.Sprintf(messages[rand.Intn(len(messages))], rand.Intn(100))
			severity := rand.Intn(3) // 0=info, 1=warning, 2=critical (based on alert.go)

			alert := model.Alert{
				Message:  "[CHAOS] " + msg,
				Severity: severity,
				Status:   0, // Open
				Comment:  "Simulated chaos alert",
			}

			// Add to DB directly to avoid AI analysis storm
			_ = repository.AddAlertDAO(&alert)
			
			// Sleep briefly to simulate bursty nature
			if i%20 == 0 {
				time.Sleep(200 * time.Millisecond)
			}
		}
		
		service.LogSystem("warn", "chaos alert storm simulation completed", map[string]interface{}{"count": 200}, nil, "")
	}()

	respondSuccessMessage(c, http.StatusOK, "Alert storm simulation started")
}
