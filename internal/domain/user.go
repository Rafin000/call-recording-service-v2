package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type User struct {
// 	ID        int64   `json:"id"`
// 	Name      string  `json:"name"`
// 	Email     string  `json:"email"`
// 	Password  string  `json:"password"`
// 	Role      string  `json:"role"`
// 	ICustomer *string `json:"i_customer"`
// 	IsActive  bool    `json:"is_active"`
// 	CreatedAt string  `json:"created_at"`
// 	UpdatedAt string  `json:"updated_at"`
// }

// type Login struct {
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }

// type PaginatedUsers struct {
// 	Users       []User `json:"users"`
// 	TotalCount  int64  `json:"total_count"`
// 	TotalPages  int64    `json:"total_pages"`
// 	CurrentPage int    `json:"current_page"`
// }

// type UpdateUser struct {
// 	Name      string  `json:"name" binding:"required"`
// 	Email     string  `json:"email" binding:"required,email"`
// 	Role      *string `json:"role,omitempty"`       // Optional field, use pointer to indicate nullability
// 	ICustomer *string `json:"i_customer,omitempty"` // Optional field, use pointer to indicate nullability
// 	IsActive  *bool   `json:"is_active,omitempty"`  // Optional field, defaults to true
// }

// type ChangePassword struct {
// 	Password string `json:"password" binding:"required"`
// }

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"password"`
	Role      string             `bson:"role" json:"role"`
	ICustomer *string            `bson:"i_customer,omitempty" json:"i_customer,omitempty"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type UpdateUser struct {
	Name      string  `json:"name" bson:"name" binding:"required"`
	Email     string  `json:"email" bson:"email" binding:"required,email"`
	Role      *string `json:"role,omitempty" bson:"role,omitempty"`
	ICustomer *string `json:"i_customer,omitempty" bson:"i_customer,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty" bson:"is_active,omitempty"`
}

type ChangePassword struct {
	Password string `json:"password" bson:"password" binding:"required"`
}

type Login struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}

type PaginatedUsers struct {
	Users       []User `json:"users" bson:"users"`
	TotalCount  int64  `json:"total_count" bson:"total_count"`
	TotalPages  int64  `json:"total_pages" bson:"total_pages"`
	CurrentPage int    `json:"current_page" bson:"current_page"`
}
