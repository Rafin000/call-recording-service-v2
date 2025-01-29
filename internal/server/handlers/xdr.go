package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/portaone"
	"github.com/gin-gonic/gin"
)

type XDRHandler struct {
	xdrRepo        domain.XDRRepository
	portaoneClient portaone.PortaOneClient
}

func NewXDRHandler(xdrRepo domain.XDRRepository, portaoneClient portaone.PortaOneClient) *XDRHandler {
	return &XDRHandler{
		xdrRepo:        xdrRepo,
		portaoneClient: portaoneClient,
	}
}

func (h *XDRHandler) GetXDR(c *gin.Context) {
	// Get i_customer from the Gin context
	iCustomer, exists := c.Get("i_customer")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "i_customer is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Minute)
	defer cancel()

	slog.Debug("Starting GetXDR request")

	// First POST request - Login to PortaOne
	sessionID, err := h.portaoneClient.GetSessionID(ctx)
	if err != nil {
		slog.Debug("Failed to get session ID from PortaOne", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to sign in to PortaOne"})
		return
	}
	slog.Debug("Session ID retrieved", "session_id", sessionID)

	xdrURL := "https://pbwebsrv.intercloud.com.bd/rest/Customer/get_customer_xdrs"

	today := time.Now().UTC().Add(6 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	xdrHeaders := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// Format the data as URL-encoded form data
	xdrData := map[string]interface{}{
		"auth_info": json.RawMessage(fmt.Sprintf(`{"session_id": "%s"}`, sessionID)),
		"params": json.RawMessage(fmt.Sprintf(`{
			"billing_model": 1,
			"call_recording": 1,
			"from_date": "%s",
			"i_customer": "%s",
			"to_date": "%s"
		}`, today.Format("2006-01-02 15:04:05"), iCustomer, tomorrow.Format("2006-01-02 15:04:05"))),
	}
	slog.Debug("XDR data formatted", "xdr_data", xdrData)

	// Prepare form data
	formData := "auth_info=" + string(xdrData["auth_info"].(json.RawMessage)) +
		"&params=" + string(xdrData["params"].(json.RawMessage))

	req, err := http.NewRequest("POST", xdrURL, strings.NewReader(formData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create HTTP request"})
		return
	}

	// Add headers
	for key, value := range xdrHeaders {
		req.Header.Set(key, value)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("Request failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("Failed to get XDRs: %s", resp.Status)})
		return
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to read response body"})
		return
	}
	slog.Debug("Response body read", "body", string(body))

	// Parse the JSON response
	var xdrResponse map[string]interface{}
	err = json.Unmarshal(body, &xdrResponse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse XDR response"})
		return
	}

	// Send the XDR response
	c.JSON(http.StatusOK, xdrResponse)
}

func (h *XDRHandler) GetCallRecording(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Get session ID from PortaOne client
	sessionID, err := h.portaoneClient.GetSessionID(ctx)
	if err != nil || sessionID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to sign in to PortaOne"})
		return
	}

	// Get i_xdr from URL params
	iXdr := c.Param("i_xdr")
	if iXdr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "i_xdr is required"})
		return
	}

	recordingURL := "https://pbwebsrv.intercloud.com.bd/rest/CDR/get_call_recording"
	recordingData := map[string]interface{}{
		"auth_info": map[string]string{"session_id": sessionID},
		"params": map[string]string{
			"i_xdr": iXdr,
		},
	}

	// Prepare the data as JSON
	recordingDataBytes, err := json.Marshal(recordingData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to marshal request data"})
		return
	}

	// Make the POST request
	resp, err := http.Post(recordingURL, "application/json", bytes.NewBuffer(recordingDataBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"status": "error", "message": "Failed to get call recording", "response_code": resp.StatusCode})
		return
	}

	// Parse the response JSON
	var recordingResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&recordingResponse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse call recording response"})
		return
	}

	// Send the response back to the client
	c.JSON(http.StatusOK, recordingResponse)
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
	iCustomer, exists := c.Get("i_customer")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "i_customer is required"})
		return
	}

	// Type assertion to convert iCustomer to int
	customerID, ok := iCustomer.(int)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid i_customer type"})
		return
	}

	// Define pagination parameters (default to page 1 and pageSize 10)
	page := 1
	pageSize := 10

	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	// Fetch the XDR data using the repository method
	xdrData, err := h.xdrRepo.GetXDRList(ctx, customerID, fromDateUnix, toDateUnix, page, pageSize)
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
