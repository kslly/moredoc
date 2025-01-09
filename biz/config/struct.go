package config

import "time"

// 系统配置
type Settings struct {
	System   *ConfigSystem   `json:"system,omitempty"`
	Footer   *ConfigFooter   `json:"footer,omitempty"`
	Security *ConfigSecurity `json:"security,omitempty"`
	Display  *ConfigDisplay  `json:"display,omitempty"`
	Language []*Language     `json:"language,omitempty"`
}

// 系统状态
type Stats struct {
	Os              string `json:"os,omitempty"`
	Version         string `json:"version,omitempty"`
	Hash            string `json:"hash,omitempty"`
	BuildAt         string `json:"build_at,omitempty"`
	UserCount       int64  `json:"user_count,omitempty"`
	DocumentCount   int64  `json:"document_count,omitempty"`
	CategoryCount   int64  `json:"category_count,omitempty"`
	ArticleCount    int64  `json:"article_count,omitempty"`
	CommentCount    int64  `json:"comment_count,omitempty"`
	BannerCount     int64  `json:"banner_count,omitempty"`
	FriendlinkCount int64  `json:"friendlink_count,omitempty"`
	ReportCount     int64  `json:"report_count,omitempty"`
}

// 系统配置项
type ConfigSystem struct {
	RecommendWords     []string `json:"recommend_words,omitempty"`
	Domain             string   `json:"domain,omitempty"`
	Title              string   `json:"title,omitempty"`
	Keywords           string   `json:"keywords,omitempty"`
	Description        string   `json:"description,omitempty"`
	Logo               string   `json:"logo,omitempty"`
	Favicon            string   `json:"favicon,omitempty"`
	Icp                string   `json:"icp,omitempty"`
	Analytics          string   `json:"analytics,omitempty"`
	Sitename           string   `json:"sitename,omitempty"`
	CopyrightStartYear string   `json:"copyright_start_year,omitempty"`
	RegisterBackground string   `json:"register_background,omitempty"`
	LoginBackground    string   `json:"login_background,omitempty"`
	Version            string   `json:"version,omitempty"`
	CreditName         string   `json:"credit_name,omitempty"`
	SecIcp             string   `json:"sec_icp,omitempty"`
}

// 底链配置项，为跳转的链接地址
type ConfigFooter struct {
	About     string `json:"about,omitempty"`
	Contact   string `json:"contact,omitempty"`
	Agreement string `json:"agreement,omitempty"`
	Copyright string `json:"copyright,omitempty"`
	Feedback  string `json:"feedback,omitempty"`
}

// 安全配置
type ConfigSecurity struct {
	DocumentAllowedExt        []string `json:"document_allowed_ext,omitempty"`
	CloseStatement            string   `json:"close_statement,omitempty"`
	MaxDocumentSize           int32    `json:"max_document_size,omitempty"`
	IsClose                   bool     `json:"is_close,omitempty"`
	EnableRegister            bool     `json:"enable_register,omitempty"`
	EnableCaptchaLogin        bool     `json:"enable_captcha_login,omitempty"`
	EnableCaptchaRegister     bool     `json:"enable_captcha_register,omitempty"`
	EnableCaptchaComment      bool     `json:"enable_captcha_comment,omitempty"`
	EnableCaptchaFindPassword bool     `json:"enable_captcha_find_password,omitempty"`
	EnableCaptchaUpload       bool     `json:"enable_captcha_upload,omitempty"`
	LoginRequired             bool     `json:"login_required,omitempty"`
	EnableVerifyRegisterEmail bool     `json:"enable_verify_register_email,omitempty"`
}

type ConfigDisplay struct {
	CopyrightStatement          string `json:"copyright_statement,omitempty"`
	WechatTip                   string `json:"wechat_tip,omitempty"`
	WechatQrcode                string `json:"wechat_qrcode,omitempty"`
	ContactTip                  string `json:"contact_tip,omitempty"`
	ContactLink                 string `json:"contact_link,omitempty"`
	IndexDocumentStyle          string `json:"index_document_style,omitempty"`
	PagesPerRead                int32  `json:"pages_per_read,omitempty"`
	ShowRegisterUserCount       bool   `json:"show_register_user_count,omitempty"`
	ShowIndexCategories         bool   `json:"show_index_categories,omitempty"`
	ShowDocumentDescriptions    bool   `json:"show_document_descriptions,omitempty"`
	HideKeywordsOnLists         bool   `json:"hide_keywords_on_lists,omitempty"`
	ShowDocumentCount           bool   `json:"show_document_count,omitempty"`
	ShowDocumentViewCount       bool   `json:"show_document_view_count,omitempty"`
	ShowDocumentDownloadCount   bool   `json:"show_document_download_count,omitempty"`
	ShowDocumentFavoriteCount   bool   `json:"show_document_favorite_count,omitempty"`
	HideCategoryWithoutDocument bool   `json:"hide_category_without_document,omitempty"`
}

type Language struct {
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Language  string     `json:"language,omitempty"`
	Code      string     `json:"code,omitempty"`
	Id        int64      `json:"id,omitempty"`
	Total     int32      `json:"total,omitempty"`
	Sort      int32      `json:"sort,omitempty"`
	Enable    bool       `json:"enable,omitempty"`
}
