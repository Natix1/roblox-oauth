package appTypes

// FetchResponse - > response from another API
// Reply -> this API's response

type TokenFetchResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type UserInfoFetchResponse struct {
	UserId         string `json:"sub"`
	DisplayName    string `json:"nickname"`
	Username       string `json:"preferred_username"`
	CreatedAtEpoch int64  `json:"created_at"`
	ProfileUri     string `json:"profile"`
}

type SessionReply struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	UserId      int    `json:"user_id"`
}

type AuthUriReply struct {
	Url string `json:"url"`
}
