package presentationhttpresponse

type LoginResponse struct {
	User         UserResponse         `json:"user"`
	Role         RoleResponse         `json:"role"`
	Permissions  []PermissionResponse `json:"permissions"`
	AccessToken  string               `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJncmFjZS5ob3BwZXIifQ.espresso-shot-pulled"`
	RefreshToken string               `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJncmFjZS5ob3BwZXIifQ.decaf-is-not-an-option"`
}
