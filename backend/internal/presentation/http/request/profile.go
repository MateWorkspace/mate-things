package presentationhttprequest

type ProfilePatchRequest struct {
	Name     *string `json:"name" example:"Grace Hopper"`
	Bio      *string `json:"bio" example:"Keeps the espresso machines humming and the firmware fresh."`
	Username *string `json:"username" example:"grace.hopper"`
}

type ProfilePasswordPatchRequest struct {
	CurrentPassword string `json:"current_password" example:"BrewMeUp!42"`
	NewPassword     string `json:"new_password" example:"EvenMoreEspresso!7"`
}
