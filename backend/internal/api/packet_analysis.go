package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

func UploadPacketCtrl(c *gin.Context) {
	// Multipart form
	form, _ := c.MultipartForm()
	files := form.File["file"]
	
	name := c.PostForm("name")
	providerIDStr := c.PostForm("provider_id")
	modelName := c.PostForm("model")
	rawContent := c.PostForm("raw_content")

	providerID, _ := strconv.Atoi(providerIDStr)
	
	userID := uint(1) // Default or from context
	if val, ok := c.Get("user_id"); ok {
		if uid, ok := val.(uint); ok {
			userID = uid
		}
	}

	uploadDir := "public/uploads/packets"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		_ = os.MkdirAll(uploadDir, os.ModePerm)
	}

	var fileName string
	if len(files) > 0 {
		file := files[0]
		fileName = fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		dst := filepath.Join(uploadDir, fileName)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			respondError(c, fmt.Errorf("failed to save file: %w", err))
			return
		}
	}

	req := service.PacketAnalysisReq{
		Name:       name,
		ProviderID: uint(providerID),
		Model:      modelName,
		RawContent: rawContent,
	}

	pa, err := service.AddPacketAnalysisServ(req, userID, fileName)
	if err != nil {
		respondError(c, err)
		return
	}

	// Automatically start analysis
	_ = service.StartPacketAnalysisServ(pa.ID)

	respondSuccess(c, http.StatusCreated, pa)
}

func ListPacketAnalysesCtrl(c *gin.Context) {
	userID := uint(0) // Admin can see all, or filter by user
	// For now just show all
	pas, err := service.GetAllPacketAnalysesServ(userID)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, pas)
}

func DeletePacketAnalysisCtrl(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.DeletePacketAnalysisServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "deleted")
}

func StartPacketAnalysisCtrl(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.StartPacketAnalysisServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "analysis started")
}
