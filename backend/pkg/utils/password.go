package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword băm mật khẩu bằng thuật toán Bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash so sánh mật khẩu người dùng nhập với chuỗi băm trong DB
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
