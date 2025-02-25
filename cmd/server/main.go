package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/netco-crawler/internal/crawler"
	"github.com/netco-crawler/internal/models"
)

const (
	port           = 8080
	htmlDir        = "./respone"
	documentsDir   = "./static/documents"
	baseNetcoURL   = "https://www.netcovn.com.vn"
	dataOutputFile = "./static/data.json"
)

func main() {
	// Parse command line flags
	skipCrawl := flag.Bool("skip-crawl", false, "Skip crawling data and only start the web server")
	flag.Parse()

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

	// Nếu không skip crawl, thực hiện thu thập dữ liệu
	if !*skipCrawl {
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

		log.Println("Thu thập dữ liệu hoàn tất. Bắt đầu khởi động web server...")
	} else {
		log.Println("Bỏ qua thu thập dữ liệu, chỉ khởi động web server...")
	}

	// Thiết lập web server
	r := gin.Default()

	// Phục vụ tệp tĩnh
	r.Static("/documents", documentsDir)
	r.Static("/assets", "./static")

	// Thêm hàm trợ giúp cho template (phải đặt TRƯỚC khi load template)
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	})

	// Sau đó mới load templates
	r.LoadHTMLGlob("templates/*")

	// Phục vụ trang chủ - hiển thị tất cả các danh mục
	r.GET("/", func(c *gin.Context) {
		// Đọc dữ liệu từ tệp JSON
		docs, categoryMap := loadDocumentsFromJSON(dataOutputFile)

		// Tính toán thêm thống kê
		totalDocs := 0
		for _, categoryDocs := range docs {
			totalDocs += len(categoryDocs)
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":       "Tài liệu Netco",
			"Documents":   docs,
			"Categories":  categoryMap,
			"TotalDocs":   totalDocs,
			"TotalCats":   len(categoryMap),
			"LastUpdated": time.Now().Format("15:04 02/01/2006"),
		})
	})

	// Phục vụ tài liệu theo danh mục
	r.GET("/category/:name", func(c *gin.Context) {
		categoryName := c.Param("name")

		// Đọc dữ liệu từ tệp JSON
		docs, categoryMap := loadDocumentsFromJSON(dataOutputFile)

		var categoryDocs []models.Document
		var categoryTitle string

		if displayName, ok := categoryMap[categoryName]; ok {
			categoryTitle = displayName
			categoryDocs = docs[categoryName]
		}

		c.HTML(http.StatusOK, "category.html", gin.H{
			"Title":       categoryTitle,
			"Documents":   categoryDocs,
			"Categories":  categoryMap,
			"CategoryKey": categoryName,
		})
	})

	// API point để lấy dữ liệu JSON
	r.GET("/api/documents", func(c *gin.Context) {
		// Đọc dữ liệu từ tệp JSON
		docs, _ := loadDocumentsFromJSON(dataOutputFile)
		c.JSON(http.StatusOK, docs)
	})

	log.Printf("Khởi động server tại http://localhost:%d", port)
	r.Run(fmt.Sprintf(":%d", port))
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

// loadDocumentsFromJSON tải dữ liệu từ tệp JSON
func loadDocumentsFromJSON(inputPath string) (map[string][]models.Document, map[string]string) {
	file, err := os.Open(inputPath)
	if err != nil {
		log.Printf("Không thể mở tệp JSON: %v", err)
		return make(map[string][]models.Document), make(map[string]string)
	}
	defer file.Close()

	var docs map[string][]models.Document
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&docs); err != nil {
		log.Printf("Không thể decode JSON: %v", err)
		return make(map[string][]models.Document), make(map[string]string)
	}

	// Sử dụng CategoryFolderMapping để tạo danh sách danh mục cho frontend
	categoryMap := make(map[string]string)
	for cat := range docs {
		if displayName, ok := models.CategoryFolderMapping[cat]; ok {
			categoryMap[cat] = displayName
		} else {
			categoryMap[cat] = cat // Sử dụng key như là display name nếu không tìm thấy
		}
	}

	return docs, categoryMap
}
