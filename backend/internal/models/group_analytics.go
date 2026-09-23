package models

// DailyStat thống kê số tin nhắn theo ngày
type DailyStat struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"` // Số lượng tin nhắn
}

// MemberLeaderboardItem thống kê thành viên tích cực nhất
type MemberLeaderboardItem struct {
	UserID       string `json:"user_id"`
	DisplayName  string `json:"display_name"`
	Username     string `json:"username"`
	AvatarURL    string `json:"avatar_url"`
	Role         string `json:"role"` // "owner", "admin", "moderator", "member"
	MessageCount int    `json:"message_count"`
}

// ChannelActivityItem thống kê hoạt động theo từng kênh
type ChannelActivityItem struct {
	ChannelID    string  `json:"channel_id"`
	ChannelName  string  `json:"channel_name"`
	ChannelType  string  `json:"channel_type"`
	MessageCount int     `json:"message_count"`
	Percentage   float64 `json:"percentage"`
}

// MessageTypeItem thống kê theo loại nội dung tin nhắn
type MessageTypeItem struct {
	Type       string  `json:"type"` // "text", "image", "file", "voice", "poll", "event"
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// GroupAnalyticsResponse tổng hợp toàn bộ dữ liệu phân tích nhóm
type GroupAnalyticsResponse struct {
	GroupID          string                  `json:"group_id"`
	GroupName        string                  `json:"group_name"`
	TotalMessages    int                     `json:"total_messages"`
	TotalMembers     int                     `json:"total_members"`
	ActiveMembers7d  int                     `json:"active_members_7d"`
	DailyActivity    []DailyStat             `json:"daily_activity"`
	TopMembers       []MemberLeaderboardItem `json:"top_members"`
	ChannelStats     []ChannelActivityItem   `json:"channel_stats"`
	MessageTypeStats []MessageTypeItem       `json:"message_type_stats"`
	PeakHour         int                     `json:"peak_hour"` // 0-23
}
