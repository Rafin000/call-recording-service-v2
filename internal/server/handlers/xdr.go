package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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

	// Calculate date range exactly as in Python version
	today := time.Now().UTC().Add(6 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	// Format dates to match Python exactly
	fromDate := today.Format("2006-01-02") + " 00:00:00"
	toDate := tomorrow.Format("2006-01-02") + " 23:59:59"

	// Create auth_info JSON
	authInfo := map[string]string{
		"session_id": sessionID,
	}
	authInfoJSON, err := json.Marshal(authInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to marshal auth_info"})
		return
	}

	// Create params JSON with integer i_customer
	params := map[string]interface{}{
		"billing_model":  1,
		"call_recording": 1,
		"from_date":      fromDate,
		"i_customer":     iCustomer, // Now passing as integer
		"to_date":        toDate,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to marshal params"})
		return
	}

	// Create form values and encode them
	formValues := url.Values{}
	formValues.Set("auth_info", string(authInfoJSON))
	formValues.Set("params", string(paramsJSON))
	formData := formValues.Encode()

	slog.Debug("XDR request prepared",
		"auth_info", string(authInfoJSON),
		"params", string(paramsJSON))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", xdrURL, strings.NewReader(formData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create HTTP request"})
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		body, _ := io.ReadAll(resp.Body)
		slog.Debug("Error response from server", "status", resp.Status, "body", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("Failed to get XDRs: %s", resp.Status)})
		return
	}

	// Read and parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to read response body"})
		return
	}
	slog.Debug("Response received", "body", string(body))

	var xdrResponse map[string]interface{}
	err = json.Unmarshal(body, &xdrResponse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse XDR response"})
		return
	}

	c.JSON(http.StatusOK, xdrResponse)
}

func (h *XDRHandler) GetCallRecording(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Minute)
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
	// Get current time in UTC+6
	currentTime := time.Now().UTC().Add(6 * time.Hour)
	currentTimeStr := currentTime.Format("2006-01-02 15:04:05")
	slog.Debug("Received GET request for XDRDumps", "time", currentTimeStr)

	// Get i_customer from context and handle type assertion
	iCustomerAny, exists := c.Get("i_customer")
	if !exists {
		slog.Debug("Error: i_customer is required", "time", currentTimeStr)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "i_customer is required"})
		return
	}

	// Convert iCustomer to int
	var iCustomer int
	switch v := iCustomerAny.(type) {
	case int:
		iCustomer = v
	case float64:
		iCustomer = int(v)
	case string:
		var err error
		iCustomer, err = strconv.Atoi(v)
		if err != nil {
			slog.Debug("Error: invalid i_customer format", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid i_customer format"})
			return
		}
	default:
		slog.Debug("Error: invalid i_customer type", "type", fmt.Sprintf("%T", iCustomerAny))
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid i_customer type"})
		return
	}

	slog.Debug("i_customer retrieved", "i_customer", iCustomer)

	// Get query parameters with defaults
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid page number"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid page size"})
		return
	}

	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	slog.Debug("Parameters received",
		"page", page,
		"pageSize", pageSize,
		"fromDate", fromDateStr,
		"toDate", toDateStr)

	// Define default dates
	defaultFromDate := currentTime.AddDate(0, 0, -30)
	defaultToDate := currentTime.AddDate(0, 0, 1)

	// Parse dates
	fromDate, err := h.parseDateTime(fromDateStr, true, defaultFromDate)
	if err != nil {
		slog.Debug("Error parsing from_date", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid date format. Use YYYY-MM-DD or YYYY-MM-DD HH:MM[:SS].",
		})
		return
	}

	toDate, err := h.parseDateTime(toDateStr, false, defaultToDate)
	if err != nil {
		slog.Debug("Error parsing to_date", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid date format. Use YYYY-MM-DD or YYYY-MM-DD HH:MM[:SS].",
		})
		return
	}

	// Check if fromDate is after toDate
	if fromDate.After(toDate) {
		slog.Debug("Error: from_date cannot be after to_date")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "from_date cannot be after to_date",
		})
		return
	}

	// Convert to Unix timestamps
	fromDateUnix := fromDate.Unix()
	toDateUnix := toDate.Unix()

	slog.Debug("Parsed dates",
		"fromDateUnix", fromDateUnix,
		"toDateUnix", toDateUnix)

	// Call the service function to get XDR list with the properly typed iCustomer
	response, err := h.xdrRepo.GetXDRList(c.Request.Context(), iCustomer, fromDateUnix, toDateUnix, page, pageSize)
	if err != nil {
		slog.Error("Failed to get XDR list", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// parseDateTime attempts to parse a datetime string using multiple formats
func (h *XDRHandler) parseDateTime(dateStr string, isStartDate bool, defaultDate time.Time) (time.Time, error) {
	if dateStr == "" {
		return defaultDate, nil
	}

	// Define formats to try
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	var parsedTime time.Time
	var err error
	var lastErr error

	for _, format := range formats {
		parsedTime, err = time.Parse(format, dateStr)
		if err == nil {
			// If only date was provided, set appropriate time
			if format == "2006-01-02" {
				if isStartDate {
					parsedTime = time.Date(
						parsedTime.Year(),
						parsedTime.Month(),
						parsedTime.Day(),
						0, 0, 0, 0,
						time.UTC,
					)
				} else {
					parsedTime = time.Date(
						parsedTime.Year(),
						parsedTime.Month(),
						parsedTime.Day(),
						23, 59, 59, 999999000,
						time.UTC,
					)
				}
			} else {
				// For other formats, ensure UTC timezone
				parsedTime = parsedTime.UTC()
			}
			return parsedTime, nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("time data '%s' does not match any supported format: %v", dateStr, lastErr)
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
