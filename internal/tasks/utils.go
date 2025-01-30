package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/portaone"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Function to get the list of i_customer
func iCustomerList(userRepo domain.UserRepository, ctx context.Context) []string {
	var iCustomers []string
	users, err := userRepo.GetAllUsersWithICustomer(ctx)
	if err != nil {
		slog.Error("Error fetching users with i_customer", "error", err)
		return nil
	}

	for _, user := range users {
		if user.ICustomer != nil && *user.ICustomer != "" {
			iCustomers = append(iCustomers, *user.ICustomer)
		}
	}
	return iCustomers
}

// GetXDRList fetches the XDR list for a given customer within a time range.
func GetXDRList(iCustomer string, startTime string, endTime string, portaOneClient portaone.PortaOneClient, ctx context.Context) []map[string]interface{} {
	slog.Info("Fetching XDR list", "iCustomer", iCustomer, "startTime", startTime, "endTime", endTime)

	// Convert iCustomer to an integer
	iCustomerInt, err := strconv.Atoi(iCustomer)
	if err != nil {
		slog.Error("Error converting iCustomer to integer", "error", err)
		return nil
	}

	// Get session ID
	sessionID, _ := portaOneClient.GetSessionID(ctx)

	// Prepare the request data
	authInfo := map[string]string{"session_id": sessionID}
	params := map[string]interface{}{
		"billing_model":  1,
		"call_recording": 1,
		"from_date":      startTime,
		"i_customer":     iCustomerInt,
		"to_date":        endTime,
	}

	data := map[string]string{
		"auth_info": mustJSON(authInfo),
		"params":    mustJSON(params),
	}

	// Encode data as form-urlencoded
	reqBody := EncodeFormData(data)

	// Make the HTTP request
	xdrURL := "https://pbwebsrv.intercloud.com.bd/rest/Customer/get_customer_xdrs"
	req, err := http.NewRequest("POST", xdrURL, bytes.NewBufferString(reqBody))
	if err != nil {
		slog.Error("Error creating request", "error", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error making request", "error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Error response from server", "status", resp.Status)
		return nil
	}

	// Parse response JSON
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Error decoding JSON response", "error", err)
		return nil
	}

	// Extract XDR list
	if xdrList, ok := result["xdr_list"].([]map[string]interface{}); ok {
		return xdrList
	}

	return nil
}

// mustJSON marshals a value to JSON and returns it as a string.
func mustJSON(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("Error marshaling JSON", "error", err)
		return "{}"
	}
	return string(jsonData)
}

// EncodeFormData encodes a map[string]string into a form-urlencoded string.
func EncodeFormData(data map[string]string) string {
	form := ""
	for key, value := range data {
		form += fmt.Sprintf("%s=%s&", key, value)
	}
	return form[:len(form)-1] // Remove trailing '&'
}

func DownloadRecordings(xdrList []map[string]interface{}, iCustomer string, dateString string, cfg common.AppSettings, portaOneClient portaone.PortaOneClient, ctx context.Context, xdrRepo domain.XDRRepository) {
	recordingURL := "https://pbwebsrv.intercloud.com.bd/rest/CDR/get_call_recording"

	// Create recordings directory if it doesn't exist
	saveDirectory := filepath.Join(".", "recordings")
	if err := os.MkdirAll(saveDirectory, 0755); err != nil {
		slog.Error("Failed to create recordings directory", "error", err)
		return
	}

	s3Client := createS3Client(cfg)

	for _, xdr := range xdrList {
		sessionID, err := portaOneClient.GetSessionID(ctx)
		if err != nil {
			slog.Error("Failed to sign in to PortaOne", "error", err)
			continue
		}

		// Prepare recording request data
		recordingData := map[string]string{
			"auth_info": fmt.Sprintf(`{"session_id": "%s"}`, sessionID),
			"params":    fmt.Sprintf(`{"i_xdr": %v}`, xdr["i_xdr"]),
		}

		// Encode data as form-urlencoded
		reqBody := EncodeFormData(recordingData)

		// Make the HTTP request to fetch the recording
		req, err := http.NewRequest("POST", recordingURL, bytes.NewBufferString(reqBody))
		if err != nil {
			slog.Error("Error creating request", "error", err)
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Error making request", "error", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			slog.Error("Error response from server", "status", resp.Status)
			continue
		}

		// Read the recording into memory
		var audioBuffer bytes.Buffer
		if _, err := io.Copy(&audioBuffer, resp.Body); err != nil {
			slog.Error("Failed to read recording response", "error", err)
			continue
		}

		// Save the recording locally
		iXDR := fmt.Sprintf("%v", xdr["i_xdr"])
		filename := fmt.Sprintf("recording_%s.wav", iXDR)
		filepath := filepath.Join(saveDirectory, filename)

		if err := os.WriteFile(filepath, audioBuffer.Bytes(), 0644); err != nil {
			slog.Error("Failed to write recording to file", "error", err)
			continue
		}

		// Upload to S3
		s3Key := fmt.Sprintf("%s/%s/%s", iCustomer, dateString, filename)
		if success := uploadToS3(ctx, s3Client, filepath, cfg.S3_BUCKET_NAME, s3Key); success {
			// Remove local file after successful upload
			if err := os.Remove(filepath); err != nil {
				slog.Error("Failed to remove local file", "error", err, "filepath", filepath)
			} else {
				slog.Info("Removed local file", "filepath", filepath)
			}
		} else {
			slog.Error("Failed to upload recording to S3", "filepath", filepath)
		}
	}
}

func uploadToS3(ctx context.Context, client *s3.Client, filePath, bucketName, s3Key string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("Failed to open file for S3 upload", "error", err)
		return false
	}
	defer file.Close()

	// Create S3 upload input
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	}

	// Upload to S3
	_, err = client.PutObject(ctx, input)
	if err != nil {
		slog.Error("Failed to upload to S3", "error", err, "bucket", bucketName, "key", s3Key)
		return false
	}

	slog.Info("Uploaded file to S3", "filepath", filePath, "bucket", bucketName, "key", s3Key)
	return true
}

func createS3Client(cfg common.AppSettings) *s3.Client {
	// Custom endpoint resolver for S3
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: cfg.S3_ENDPOINT_URL,
		}, nil
	})

	// AWS configuration setup
	awsCfg := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AWS_ACCESS_KEY,
			cfg.AWS_SECRET_ACCESS_KEY,
			"",
		),
		EndpointResolverWithOptions: customResolver,
	}

	return s3.NewFromConfig(awsCfg)
}
