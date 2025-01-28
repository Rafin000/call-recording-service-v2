package domain

import (
	"context"
	"database/sql"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user User) (int64, error)
	UpdateUser(ctx context.Context, userID int64, data map[string]interface{}) error
	GetAllUsers(ctx context.Context, currentPage int, pageSize int) (PaginatedUsers, error)
	GetAllUsersWithICustomer(ctx context.Context) ([]User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := "SELECT id, name, email, password, role, i_customer, is_active, created_at, updated_at FROM users WHERE email = ? AND is_active = true"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.ICustomer, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user User) (int64, error) {
	query := "INSERT INTO users (name, email, password, role, i_customer, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := r.db.Exec(query, user.Name, user.Email, user.Password, user.Role, user.ICustomer, user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return 0, err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, userID int64, data map[string]interface{}) error {
	query := "UPDATE users SET name = ?, email = ?, role = ?, i_customer = ?, updated_at = ? WHERE id = ?"
	_, err := r.db.Exec(query, data["name"], data["email"], data["role"], data["i_customer"], data["updated_at"], userID)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetAllUsers(ctx context.Context, currentPage int, pageSize int) (PaginatedUsers, error) {
	offset := (currentPage - 1) * pageSize
	var users []User

	// Count total active users
	var totalCount int64
	countQuery := "SELECT COUNT(*) FROM users WHERE is_active = true"
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return PaginatedUsers{}, err
	}

	// Calculate total pages
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	// Fetch paginated users
	query := `
		SELECT id, name, email, role, i_customer
		FROM users
		WHERE is_active = true
		ORDER BY id
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return PaginatedUsers{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.ICustomer)
		if err != nil {
			return PaginatedUsers{}, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return PaginatedUsers{}, err
	}

	return PaginatedUsers{
		Users:       users,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: currentPage,
	}, nil
}

func (r *userRepository) GetAllUsersWithICustomer(ctx context.Context) ([]User, error) {
	var users []User

	query := `
		SELECT id, name, email, role, i_customer
		FROM users
		WHERE is_active = true AND i_customer IS NOT NULL
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.ICustomer)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
