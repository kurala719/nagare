import sys
with open('backend/internal/adapter/handler/alarm.go', 'a', encoding='utf-8') as f:
    f.write('''
// SetupAlarmMediaCtrl handles POST /alarms/:id/setup-media
func SetupAlarmMediaCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alarm ID")
		return
	}

	err = service.SetupAlarmMediaServ(c.Request.Context(), uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"message": "Alarm media setup successfully"})
}
''')
