package model

// SocialActionRequest defines the request body for social actions.
// SocialActionRequest là request để thực hiện một hành động trên mạng xã hội
type SocialActionRequest struct {
	Social string `json:"social" binding:"required"`
	Action string `json:"action" binding:"required"`
	IDUser string `json:"iduser" binding:"required"`
}

// SocialPlatform định nghĩa các nền tảng mạng xã hội được hỗ trợ
type SocialPlatform string

const (
	Facebook  SocialPlatform = "facebook"
	Instagram SocialPlatform = "instagram"
	Twitter   SocialPlatform = "twitter"
	TikTok    SocialPlatform = "tiktok"
	YouTube   SocialPlatform = "youtube"
	LinkedIn  SocialPlatform = "linkedin"
)

// CheckRequest là request để kiểm tra tài khoản mạng xã hội
type CheckRequest struct {
	Platform SocialPlatform `json:"platform" binding:"required"`
	Username string         `json:"username" binding:"required,min=1,max=100"`
}

// CheckResponse là response sau khi kiểm tra tài khoản
type CheckResponse struct {
	Platform   SocialPlatform `json:"platform"`
	Username   string         `json:"username"`
	Exists     bool           `json:"exists"`
	ProfileURL string         `json:"profile_url,omitempty"`
	Message    string         `json:"message,omitempty"`
}

// BatchCheckRequest để kiểm tra nhiều tài khoản cùng lúc
type BatchCheckRequest struct {
	Checks []CheckRequest `json:"checks" binding:"required,min=1,max=10"`
}

// BatchCheckResponse chứa kết quả kiểm tra nhiều tài khoản
type BatchCheckResponse struct {
	Results []CheckResponse `json:"results"`
	Total   int             `json:"total"`
	Success int             `json:"success"`
	Failed  int             `json:"failed"`
}
