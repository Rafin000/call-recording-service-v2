package task

import (
	"context"
	"log/slog"

	"github.com/Rafin000/call-recording-service-v2/internal/domain"
)

// Function to get the list of i_customer
func iCustomerList(userRepo domain.UserRepository, ctx context.Context) []string {
	var iCustomers []string
	users, err := domain.UserRepository.GetAllUsersWithICustomer(userRepo, ctx)
	if err != nil {
		slog.Error("Error fetching users with i_customer: ", err)
		return nil
	}

	for _, user := range users {
		if user.ICustomer != nil && *user.ICustomer != "" {
			iCustomers = append(iCustomers, *user.ICustomer)
		}
	}
	return iCustomers
}
