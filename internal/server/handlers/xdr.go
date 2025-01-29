package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/gin-gonic/gin"
)

type XDRHandler struct {
	xdrRepo domain.XDRRepository
}

func NewXDRHandler(xdrRepo domain.XDRRepository) *XDRHandler {
	return &XDRHandler{
		xdrRepo: xdrRepo,
	}
}

func (h *XDRHandler) GetXDR(c *gin.Context) {
	sessionID, err := signInToPortaOne()
	if err != nil || sessionID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to sign in to PortaOne"})
		return
	}

	var req domain.XDRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON format"})
		return
	}

	// Check if i_customer is provided in the request
	if req.ICustomer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "i_customer is required"})
		return
	}

	xdrURL := "https://pbwebsrv.intercloud.com.bd/rest/Customer/get_customer_xdrs"
	today := time.Now().UTC().Add(6 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	xdrData := map[string]interface{}{
		"auth_info": map[string]string{"session_id": sessionID},
		"params": map[string]interface{}{
			"billing_model":  1,
			"call_recording": 1,
			"from_date":      today.Format("2006-01-02 15:04:05"),
			"i_customer":     req.ICustomer,
			"to_date":        tomorrow.Format("2006-01-02 15:04:05"),
		},
	}

	xdrDataBytes, _ := json.Marshal(xdrData)

	resp, err := http.Post(xdrURL, "application/x-www-form-urlencoded", bytes.NewBuffer(xdrDataBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	defer resp.Body.Close()

	var xdrResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&xdrResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse XDR response"})
		return
	}

	c.JSON(http.StatusOK, xdrResponse)
}

func (h *XDRHandler) GetCallRecording(c *gin.Context) {
	sessionID, err := signInToPortaOne()
	if err != nil || sessionID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to sign in to PortaOne"})
		return
	}

	iXdr := c.Param("i_xdr")
	recordingURL := "https://pbwebsrv.intercloud.com.bd/rest/CDR/get_call_recording"

	recordingData := map[string]interface{}{
		"auth_info": map[string]string{"session_id": sessionID},
		"params": map[string]string{
			"i_xdr": iXdr,
		},
	}

	recordingDataBytes, _ := json.Marshal(recordingData)

	resp, err := http.Post(recordingURL, "application/x-www-form-urlencoded", bytes.NewBuffer(recordingDataBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	defer resp.Body.Close()

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to read audio data"})
		return
	}

	// Write the recording to a temporary file
	inputFilePath := fmt.Sprintf("/tmp/%s.wav", iXdr)
	err = os.WriteFile(inputFilePath, audioData, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save recording"})
		return
	}

	// Convert the WAV file to MP3 using ffmpeg
	outputFilePath := fmt.Sprintf("/tmp/%s.mp3", iXdr)
	cmd := exec.Command("ffmpeg", "-i", inputFilePath, "-acodec", "mp3", "-ar", "24000", "-ac", "2", "-ab", "128k", outputFilePath)
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Audio conversion failed"})
		return
	}

	// Read the converted MP3 file
	mp3Data, err := os.ReadFile(outputFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to read MP3 data"})
		return
	}

	// Clean up temporary files
	defer os.Remove(inputFilePath)
	defer os.Remove(outputFilePath)

	// Send the MP3 file as a response
	c.Header("Content-Type", "audio/mp3")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.mp3", iXdr))
	c.Data(http.StatusOK, "audio/mp3", mp3Data)
}

// getXDRDumps handles fetching XDR dumps within a given date range
func (h *XDRHandler) GetXDRDumps(c *gin.Context) {
	var req domain.XDRDumpsRequest
	// Bind query parameters to the struct
	if err := c.ShouldBindQuery(&req); err != nil {
		// Handle missing or invalid parameters
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Missing required parameters"})
		return
	}

	// Parse and validate date range (from_date, to_date)
	fromDate, err := time.Parse("2006-01-02 15:04:05", req.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid from_date format"})
		return
	}

	toDate, err := time.Parse("2006-01-02 15:04:05", req.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid to_date format"})
		return
	}

	// Ensure fromDate is not after toDate
	if fromDate.After(toDate) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "from_date cannot be after to_date"})
		return
	}

	// Convert to Unix timestamps
	fromDateUnix := fromDate.Unix()
	toDateUnix := toDate.Unix()

	// Get customer ID from somewhere (e.g., request context or headers)
	// Here, I'm assuming a hardcoded value for the example
	iCustomer := 123 // Replace with dynamic value

	// Define pagination parameters (default to page 1 and pageSize 10)
	page := 1
	pageSize := 10

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Fetch the XDR data using the repository method
	xdrData, err := h.xdrRepo.GetXDRList(ctx, iCustomer, fromDateUnix, toDateUnix, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error fetching XDR data"})
		return
	}

	// Return the fetched XDR data
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"xdr_data": xdrData,
	})
}

// getXDRByI_XDR handles fetching XDR data for a specific i_xdr
func (h *XDRHandler) GetXDRByI_XDR(c *gin.Context) {
	iXdrStr := c.Param("i_xdr")

	// Convert iXdr from string to int using strconv.Atoi
	iXdr, err := strconv.Atoi(iXdrStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid i_xdr format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Fetch the XDR data using the repository method
	xdrData, err := h.xdrRepo.GetXDRByIXDR(ctx, iXdr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error fetching XDR data"})
		return
	}

	if xdrData == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "XDR not found"})
		return
	}

	// Return the fetched XDR data
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"xdr_data": xdrData,
	})
}
