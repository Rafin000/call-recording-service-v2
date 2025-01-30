package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/portaone"
)

func BackupTask(
	userRepo domain.UserRepository,
	XDRRepo domain.XDRRepository,
	ctx context.Context,
	cfg common.AppSettings,
	portaOneClient portaone.PortaOneClient,
) {
	customers := iCustomerList(userRepo, ctx)

	fmt.Println("*********************ENTERED BACKUP TASK**************")

	currentTime := time.Now().UTC()
	previousTime := currentTime.Add(-24 * time.Hour)

	fmt.Println("*********************Current time:", currentTime)
	fmt.Println("*********************Previous time (24 hours ago):", previousTime)

	previousTimeAdjusted := previousTime.Add(12 * time.Hour)

	startTimeStr := previousTimeAdjusted.Format("2006-01-02 00:00:00")
	endTimeAdjusted := previousTimeAdjusted.Add(24 * time.Hour).Add(-time.Second)
	endTimeStr := endTimeAdjusted.Format("2006-01-02 15:04:05")

	fmt.Println("Start Time:", startTimeStr)
	fmt.Println("End Time:", endTimeStr)
	dateString := previousTimeAdjusted.Format("2006-01-02")

	for _, customer := range customers {
		xdrList := GetXDRList(customer, startTimeStr, endTimeStr, portaOneClient, ctx)
		DownloadRecordings(xdrList, customer, dateString, cfg, portaOneClient, ctx, XDRRepo)
	}
}
