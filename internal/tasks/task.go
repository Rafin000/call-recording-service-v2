package task

import (
	"context"
	"fmt"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/gin-gonic/gin"
)

func BackupTask(userRepo domain.UserRepository, c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), common.Timeouts.User.Write)
	defer cancel()

	customers := iCustomerList(userRepo, ctx)

	currentTime := time.Now().UTC().Add(6 * time.Hour)
	fmt.Println("Current time:", currentTime)
	previousTime := currentTime.Add(-24 * time.Hour)

	startTime := previousTime.Format("2006-01-02 00:00:00")
	endTime := previousTime.Format("2006-01-02 23:59:59")
	dateString := previousTime.Format("2006-01-02")

	for _, customer := range customers {
		xdrList := getXdrList(customer, startTime, endTime)
		downloadRecordings(xdrList, customer, dateString)
	}
}
