package models

// WSEvent chuẩn hóa cấu trúc gói tin gửi nhận qua WebSocket
type WSEvent struct {
	Event   string      `json:"event"`   // "chat:send", "chat:receive", "chat:ack", "ping", "pong"
	Payload interface{} `json:"payload"` // Nội dung dữ liệu tương ứng
}
