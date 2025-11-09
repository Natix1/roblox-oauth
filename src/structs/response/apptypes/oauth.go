package apptypes

type TokenReply struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type UserInfoReply struct {
	UserId         string `json:"sub"`
	DisplayName    string `json:"nickname"`
	Username       string `json:"preferred_username"`
	CreatedAtEpoch int64  `json:"created_at"`
	ProfileUri     string `json:"profile"`
}
