package http

// RegisterClientDoc godoc
// @Summary Register client user
// @Tags Identity
// @Accept json
// @Produce json
// @Param body body RegisterClientRequest true "Client registration payload"
// @Success 201 {object} IdentityResponse
// @Router /identity/users/client [post]
func RegisterClientDoc() {}

// RegisterUserDoc godoc
// @Summary Register standard user
// @Tags Identity
// @Accept json
// @Produce json
// @Param body body RegisterUserRequest true "User registration payload"
// @Success 201 {object} IdentityResponse
// @Router /identity/users [post]
func RegisterUserDoc() {}

// VerifyUserDoc godoc
// @Summary Verify user with code
// @Tags Identity
// @Accept json
// @Produce json
// @Param body body VerifyUserRequest true "Verification payload"
// @Success 204
// @Router /identity/verify [post]
func VerifyUserDoc() {}

// LoginDoc godoc
// @Summary Login user
// @Tags Identity
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login payload"
// @Success 200 {object} LoginResponse
// @Router /identity/login [post]
func LoginDoc() {}

// UpdateUserDoc godoc
// @Summary Update user profile (no role change)
// @Tags Identity
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body UpdateUserRequest true "User update payload"
// @Success 200 {object} IdentityResponse
// @Security BearerAuth
// @Router /identity/users/{id} [put]
func UpdateUserDoc() {}

// UpdateUserRoleDoc godoc
// @Summary Update user role (admin)
// @Tags Identity
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body UpdateUserRoleRequest true "Role payload"
// @Success 200 {object} IdentityResponse
// @Security BearerAuth
// @Router /identity/users/{id}/role [put]
func UpdateUserRoleDoc() {}

// BlockUserDoc godoc
// @Summary Block user (admin)
// @Tags Identity
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body BlockUserRequest false "Block reason"
// @Success 204
// @Security BearerAuth
// @Router /identity/users/{id}/block [post]
func BlockUserDoc() {}
