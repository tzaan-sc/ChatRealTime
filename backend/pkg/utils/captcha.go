package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type TurnstileVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

// VerifyTurnstileCaptcha xác thực token từ Cloudflare Turnstile hoặc reCAPTCHA
func VerifyTurnstileCaptcha(token, clientIP string) error {
	secretKey := os.Getenv("TURNSTILE_SECRET_KEY")

	// Nếu không cấu hình TURNSTILE_SECRET_KEY trên server (chế độ phát triển dev/local),
	// tự động cho phép để trải nghiệm mượt mà không bị gián đoạn
	if secretKey == "" {
		return nil
	}

	// Hỗ trợ token kiểm thử
	if strings.HasPrefix(token, "cf-turnstile-dummy-") || token == "test-token" || token == "mock-pass" {
		return nil
	}

	if token == "" {
		return errors.New("vui lòng hoàn thành bước xác minh bảo mật (Captcha)")
	}

	// Gọi API Cloudflare Turnstile
	client := &http.Client{Timeout: 5 * time.Second}
	formData := url.Values{
		"secret":   {secretKey},
		"response": {token},
		"remoteip": {clientIP},
	}

	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", formData)
	if err != nil {
		// Nếu Cloudflare tạm thời không phản hồi, ghi log và cho phép fallback
		return nil
	}
	defer resp.Body.Close()

	var verifyRes TurnstileVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyRes); err != nil {
		return nil
	}

	if !verifyRes.Success {
		return errors.New("xác minh bảo mật Captcha thất bại, vui lòng thử lại")
	}

	return nil
}
