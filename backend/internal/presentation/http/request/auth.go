package presentationhttprequest

type AuthLoginRequest struct {
	Username string `json:"username" example:"grace.hopper"`
	Password string `json:"password" example:"BrewMeUp!42"`
}

type AuthRefreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJncmFjZS5ob3BwZXIifQ.brewbrewbrew"`
}
