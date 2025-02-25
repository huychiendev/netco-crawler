package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadFile tải xuống tệp từ URL và lưu vào đường dẫn đã chỉ định
func DownloadFile(url, destPath string) error {
	// Tạo thư mục đích nếu chưa tồn tại
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục đích: %w", err)
	}

	// Tạo tệp tạm
	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("không thể tạo tệp tạm: %w", err)
	}
	defer out.Close()

	// Lấy nội dung từ URL
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("không thể tải tệp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("phản hồi lỗi: %s", resp.Status)
	}

	// Sao chép nội dung vào tệp
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("lỗi khi sao chép nội dung: %w", err)
	}
	out.Close()

	// Đổi tên tệp tạm thành tệp đích
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("không thể đổi tên tệp tạm: %w", err)
	}

	return nil
}

// EnsureDirectoryExists đảm bảo thư mục đã cho tồn tại
func EnsureDirectoryExists(dir string) error {
	return os.MkdirAll(dir, 0755)
}
