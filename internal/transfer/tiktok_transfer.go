package transfer

type TikTokResponse struct {
	Data  TiktokUserData `json:"data"`
	Error TiktokError    `json:"error"`
}

type TikTokUploadResponse struct {
	Data  TiktokPublishData `json:"data"`
	Error TiktokError       `json:"error"`
}

type TiktokPublishData struct {
	PublishID string `json:"publish_id"`
}

type TiktokUserData struct {
	User TiktokUser `json:"user"`
}

type TiktokUser struct {
	OpenID      string `json:"open_id"`
	AvatarURL   string `json:"avatar_url"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
}

type TiktokError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	LogID   string `json:"log_id"`
}

type VideoPostInfo struct {
	Title                 string `json:"title"`
	PrivacyLevel          string `json:"privacy_level"`
	DisableDuet           bool   `json:"disable_duet"`
	DisableComment        bool   `json:"disable_comment"`
	DisableStitch         bool   `json:"disable_stitch"`
	VideoCoverTimestampMs int    `json:"video_cover_timestamp_ms"`
	BrandContentToggle    bool   `json:"brand_content_toggle"`
	Brand_Organic_Toggle  bool   `json:"brand_organic_toggle"`
	IsAIGC                bool   `json:"is_aigc"`
}

type PhotoPostInfo struct {
	Title                string `json:"title"`
	Description          string `json:"description"`
	PrivacyLevel         string `json:"privacy_level"`
	DisableComment       bool   `json:"disable_comment"`
	AutoAddMusic         bool   `json:"auto_add_music"`
	BrandContentToggle   bool   `json:"brand_content_toggle"`
	Brand_Organic_Toggle bool   `json:"brand_organic_toggle"`
}

type VideoSourceInfo struct {
	Source   string `json:"source"`
	VideoURL string `json:"video_url"`
}

type PhotoSourceInfo struct {
	Source          string   `json:"source"`
	PhotoCoverIndex int      `json:"photo_cover_index"`
	PhotoImages     []string `json:"photo_images"`
}

type VideoUploadRequest struct {
	PostInfo   VideoPostInfo   `json:"post_info"`
	SourceInfo VideoSourceInfo `json:"source_info"`
}

type PhotUploadRequest struct {
	PostInfo   PhotoPostInfo   `json:"post_info"`
	SourceInfo PhotoSourceInfo `json:"source_info"`
	PostMode   string          `json:"post_mode"`
	MediaType  string          `json:"media_type"`
}

type TiktokCreatorInfo struct {
	CreatorAvatarURL        string   `json:"creator_avatar_url"`
	CreatorUsername         string   `json:"creator_username"`
	CreatorNickname         string   `json:"creator_nickname"`
	PrivacyLevelOptions     []string `json:"privacy_level_options"`
	CommentDisabled         bool     `json:"comment_disabled"`
	DuetDisabled            bool     `json:"duet_disabled"`
	StitchDisabled          bool     `json:"stitch_disabled"`
	MaxVideoPostDurationSec int32    `json:"max_video_post_duration_sec"`
}

type TiktokTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	OpenID           string `json:"open_id"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
}

type TiktokRevokeData struct {
	ErrorCode   int64  `json:"error_code"`
	Description string `json:"description"`
}
