package presentationhttpresponse

type LoginResponse struct {
	User         UserResponse         `json:"user"`
	Role         RoleResponse         `json:"role"`
	Permissions  []PermissionResponse `json:"permissions"`
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token,omitempty"`
}
