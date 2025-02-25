package models

import (
	"path/filepath"
	"strings"
)

// Document đại diện cho một tài liệu từ trang web Netco
type Document struct {
	Name        string `json:"name"`
	Size        string `json:"size"`
	Downloads   string `json:"downloads"`
	Modified    string `json:"modified"`
	UploadedBy  string `json:"uploaded_by"`
	DownloadURL string `json:"download_url"`
	Category    string `json:"category"`
	FilePath    string `json:"file_path"` // đường dẫn cục bộ sau khi tải về
}

// CategoryFromURL trả về tên danh mục từ đường dẫn URL
func CategoryFromURL(url string) string {
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "" || i == len(parts)-1 {
			continue
		}
		if strings.Contains(part, "?") {
			return strings.Split(part, "?")[0]
		}
	}
	return filepath.Base(url)
}

// CategoryFolderMapping ánh xạ từ danh mục URL sang tên thư mục tiếng Việt mô tả
var CategoryFolderMapping = map[string]string{
	"bao-cao-thuong-nien":      "Báo cáo thường niên",
	"bao-cao-tai-chinh":        "Báo cáo tài chính",
	"dieu-le-cong-ty":          "Điều lệ công ty",
	"quy-che-quan-tri-cong-ty": "Quy chế quản trị công ty",
	"cong-bao-thong-tin":       "Công bố thông tin",
	"ban-cao-bach":             "Bản cáo bạch",
}

// FileName trả về tên tệp duy nhất dựa trên tên tài liệu
func (d *Document) FileName() string {
	// Loại bỏ các ký tự không hợp lệ cho tên tệp
	name := strings.ReplaceAll(d.Name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")

	return name
}
