package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/netco-crawler/internal/crawler"
	"github.com/netco-crawler/internal/models"
)

const (
	htmlDir        = "./respone"
	documentsDir   = "./static/documents"
	baseNetcoURL   = "https://www.netcovn.com.vn"
	dataOutputFile = "./static/data.json"
)

func main() {
	// Kiểm tra thư mục HTML
	if _, err := os.Stat(htmlDir); os.IsNotExist(err) {
		log.Fatalf("Thư mục HTML không tồn tại: %s", htmlDir)
	}

	// Đảm bảo thư mục đích tồn tại
	if err := os.MkdirAll(documentsDir, 0755); err != nil {
		log.Fatalf("Không thể tạo thư mục đích: %v", err)
	}

	// Tạo crawler
	c := crawler.NewCrawler(htmlDir, documentsDir, baseNetcoURL)

	// Xử lý tệp HTML
	log.Println("Bắt đầu phân tích các tệp HTML...")
	if err := c.ProcessHTMLFiles(); err != nil {
		log.Fatalf("Lỗi khi xử lý tệp HTML: %v", err)
	}

	// Tải xuống tài liệu
	log.Println("Bắt đầu tải xuống tài liệu...")
	if err := c.DownloadDocuments(); err != nil {
		log.Fatalf("Lỗi khi tải xuống tài liệu: %v", err)
	}

	// Xuất dữ liệu sang JSON
	if err := saveDataToJSON(c.GetDocuments(), dataOutputFile); err != nil {
		log.Fatalf("Lỗi khi lưu dữ liệu vào JSON: %v", err)
	}

	// In thống kê
	printStats(c.GetDocuments())

	log.Println("Hoàn tất! Các tài liệu đã được lưu trong", documentsDir)
}

// saveDataToJSON lưu dữ liệu vào tệp JSON
func saveDataToJSON(docs map[string][]models.Document, outputPath string) error {
	// Đảm bảo thư mục đích tồn tại
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục đích: %w", err)
	}

	// Tạo tệp đích
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("không thể tạo tệp JSON: %w", err)
	}
	defer file.Close()

	// Encode dữ liệu thành JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(docs); err != nil {
		return fmt.Errorf("không thể encode dữ liệu thành JSON: %w", err)
	}

	log.Printf("Đã lưu dữ liệu vào %s", outputPath)
	return nil
}

// printStats in thống kê về tài liệu đã tải
func printStats(docs map[string][]models.Document) {
	var totalDocs int
	log.Println("Thống kê tài liệu:")
	log.Println("--------------------------------------------------")

	for category, categoryDocs := range docs {
		displayName, exists := models.CategoryFolderMapping[category]
		if !exists {
			displayName = category
		}

		log.Printf("- %s: %d tài liệu", displayName, len(categoryDocs))
		totalDocs += len(categoryDocs)
	}

	log.Println("--------------------------------------------------")
	log.Printf("Tổng cộng: %d tài liệu đã thu thập", totalDocs)
	log.Println("Lưu ý: Số lượng tài liệu này đã loại bỏ các bản trùng lặp")
	log.Println("Các tài liệu trùng lặp được xác định dựa trên tên, URL tải xuống, danh mục và đường dẫn tệp")
	log.Println("Hệ thống ưu tiên giữ lại tài liệu có đầy đủ thông tin nhất (ít trường rỗng nhất)")
	log.Println("--------------------------------------------------")
}
