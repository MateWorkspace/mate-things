package presentationhttprequest

type ProfilePatchRequest struct {
	Name     *string `json:"name"`
	Bio      *string `json:"bio"`
	Username *string `json:"username"`
}

type ProfilePasswordPatchRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
