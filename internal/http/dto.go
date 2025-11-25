package http

// Identity DTOs keep transport-only fields outside the domain layer.

type RegisterClientRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" binding:"omitempty"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type VerifyUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

type BlockUserRequest struct {
	Reason string `json:"reason" binding:"omitempty"`
}

type IdentityResponse struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Role       string `json:"role"`
	Status     string `json:"status"`
	IsVerified bool   `json:"is_verified"`
}

// Catalog DTOs

type CategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"omitempty"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty"`
}

// Product DTOs

type ProductResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int64  `json:"stock"`
}

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"omitempty"`
	Price       int64  `json:"price" binding:"required,min=0"`
	Stock       int64  `json:"stock" binding:"required,min=0"`
}

type UpdateProductRequest struct {
	Name        string `json:"name" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty"`
	Price       int64  `json:"price" binding:"omitempty,min=0"`
	Stock       int64  `json:"stock" binding:"omitempty,min=0"`
}

type ProductHistoryResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Price     int64  `json:"price"`
	Stock     int64  `json:"stock"`
	ChangedAt string `json:"changed_at"`
}
