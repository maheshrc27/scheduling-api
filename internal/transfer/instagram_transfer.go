package transfer

import "time"

type InstagramToken struct {
	UserID         int       `json:"user_id"`
	AccessToken    string    `json:"access_token"`
	LongLivedToken string    `json:"long_lived_token"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type InstagramUserInfo struct {
	UserID         string `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture_url"`
}
