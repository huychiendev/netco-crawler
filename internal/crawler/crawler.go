package crawler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/netco-crawler/internal/models"
	"github.com/netco-crawler/internal/utils"
)

// Các URL cơ sở cho các trang cần thu thập
var baseCategories = []string{
	"bao-cao-thuong-nien",
	"bao-cao-tai-chinh",
	"dieu-le-cong-ty",
	"quy-che-quan-tri-cong-ty",
	"cong-bao-thong-tin",
	"ban-cao-bach",
}

// Crawler thực hiện thu thập dữ liệu từ các trang web của Netco
type Crawler struct {
	htmlDir        string
	documentsDir   string
	baseURL        string
	categories     []string
	documents      map[string][]models.Document
	mu             sync.Mutex
	maxConcurrent  int
	duplicateMap   map[string]models.Document // Thay đổi từ map[string]bool thành map[string]models.Document để lưu trữ tài liệu
	duplicateCount int                        // Số lượng tài liệu trùng lặp
}

// NewCrawler tạo một crawler mới
func NewCrawler(htmlDir, documentsDir, baseURL string) *Crawler {
	// Đảm bảo thư mục tồn tại
	if err := utils.EnsureDirectoryExists(htmlDir); err != nil {
		log.Printf("Lỗi tạo thư mục HTML: %v, sẽ tiếp tục với thư mục hiện có", err)
	}

	if err := utils.EnsureDirectoryExists(documentsDir); err != nil {
		log.Printf("Lỗi tạo thư mục documents: %v, sẽ tiếp tục với thư mục hiện có", err)
	}

	return &Crawler{
		htmlDir:        htmlDir,
		documentsDir:   documentsDir,
		baseURL:        baseURL,
		categories:     []string{"bao-cao-thuong-nien", "bao-cao-tai-chinh", "dieu-le-cong-ty", "quy-che-quan-tri-cong-ty", "cong-bao-thong-tin", "ban-cao-bach"},
		documents:      make(map[string][]models.Document),
		maxConcurrent:  10,                               // Tăng số luồng tải xuống tối đa từ 5 lên 10
		duplicateMap:   make(map[string]models.Document), // Khởi tạo map phát hiện trùng lặp
		duplicateCount: 0,
	}
}

// ProcessHTMLFiles xử lý các tệp HTML đã cho để trích xuất thông tin tài liệu
func (c *Crawler) ProcessHTMLFiles() error {
	for _, category := range baseCategories {
		// Đảm bảo thư mục đích tồn tại
		categoryDir := filepath.Join(c.documentsDir, models.CategoryFolderMapping[category])
		if err := utils.EnsureDirectoryExists(categoryDir); err != nil {
			return fmt.Errorf("không thể tạo thư mục danh mục %s: %w", category, err)
		}

		// Đọc tệp HTML để xác định số trang tối đa
		htmlPath := filepath.Join(c.htmlDir, category+".html")
		file, err := os.Open(htmlPath)
		if err != nil {
			return fmt.Errorf("không thể mở tệp HTML %s: %w", htmlPath, err)
		}

		// Phân tích tài liệu HTML chỉ để lấy thông tin phân trang
		doc, err := goquery.NewDocumentFromReader(file)
		file.Close() // Đóng file sớm sau khi đọc xong
		if err != nil {
			return fmt.Errorf("không thể phân tích tệp HTML %s: %w", htmlPath, err)
		}

		// Kiểm tra phân trang
		var maxPage int
		doc.Find("a.ModulePager.LastPage").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}

			// Trích xuất số trang cuối cùng từ URL
			parts := strings.Split(href, "pagenumber=")
			if len(parts) < 2 {
				return
			}

			// Phân tích số trang
			fmt.Sscanf(parts[1], "%d", &maxPage)
		})

		if maxPage <= 0 {
			maxPage = 1
		}

		// Danh sách để lưu tất cả tài liệu từ danh mục này
		var allCategoryDocs []models.Document

		// Luôn bắt đầu từ trang 1 với tham số pagenumber=1
		for page := 1; page <= maxPage; page++ {
			// Xây dựng URL với tham số pagenumber cho tất cả các trang, kể cả trang 1
			pageURL := fmt.Sprintf("%s/%s?pagenumber=%d", c.baseURL, category, page)
			log.Printf("Đang xử lý trang %d/%d của danh mục %s", page, maxPage, category)

			// Tải trang HTML
			resp, err := http.Get(pageURL)
			if err != nil {
				log.Printf("Lỗi khi tải trang %d của danh mục %s: %v", page, category, err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("Trang %d của danh mục %s trả về mã trạng thái không thành công: %d", page, category, resp.StatusCode)
				resp.Body.Close()
				continue
			}

			// Phân tích tài liệu HTML
			pageDoc, err := goquery.NewDocumentFromReader(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("Không thể phân tích trang %d của danh mục %s: %v", page, category, err)
				continue
			}

			// Trích xuất tài liệu từ trang này
			var pageDocs []models.Document
			extractDocumentsFromHTML(pageDoc, category, &pageDocs)

			// Kiểm tra trang rỗng
			if len(pageDocs) == 0 {
				log.Printf("Trang %d của danh mục %s không có tài liệu nào, dừng phân trang", page, category)
				break
			}

			log.Printf("Tìm thấy %d tài liệu trong danh mục %s (trang %d/%d)", len(pageDocs), category, page, maxPage)

			// Thêm tài liệu từ trang này vào danh sách
			allCategoryDocs = append(allCategoryDocs, pageDocs...)
		}

		// Lưu tài liệu vào bản đồ
		c.mu.Lock()
		c.documents[category] = allCategoryDocs
		c.mu.Unlock()
	}

	return nil
}

// Tách logic trích xuất tài liệu từ HTML thành một hàm riêng
func extractDocumentsFromHTML(doc *goquery.Document, category string, docs *[]models.Document) {
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		var document models.Document

		// Tên và URL tải xuống
		nameCell := s.Find("td").First()
		linkElem := nameCell.Find("a").First()
		document.Name = strings.TrimSpace(linkElem.Text())
		if document.Name == "" {
			// Bỏ qua hàng không có tên tài liệu
			return
		}

		document.DownloadURL, _ = linkElem.Attr("href")

		// Kích thước
		document.Size = strings.TrimSpace(nameCell.Next().Text())

		// Số lượt tải
		document.Downloads = strings.TrimSpace(nameCell.Next().Next().Text())

		// Ngày sửa đổi
		document.Modified = strings.TrimSpace(nameCell.Next().Next().Next().Text())

		// Người tải lên
		document.UploadedBy = strings.TrimSpace(nameCell.Next().Next().Next().Next().Text())

		// Danh mục
		document.Category = category

		// Đường dẫn tệp cục bộ
		document.FilePath = filepath.Join(
			models.CategoryFolderMapping[category],
			document.FileName(),
		)

		*docs = append(*docs, document)
	})
}

// generateDocumentHash tạo chuỗi hash để xác định tài liệu trùng lặp
func (c *Crawler) generateDocumentHash(doc models.Document) string {
	// Tạo hash dựa trên 4 trường: name, download_url, category, file_path
	return fmt.Sprintf("%s_%s_%s_%s", doc.Name, doc.DownloadURL, doc.Category, doc.FilePath)
}

// isDuplicate kiểm tra xem tài liệu có phải là bản sao không
// Nếu là bản sao, quyết định giữ lại tài liệu nào dựa trên số lượng trường thông tin rỗng
func (c *Crawler) isDuplicate(doc models.Document) bool {
	hash := c.generateDocumentHash(doc)

	c.mu.Lock()
	defer c.mu.Unlock()

	existingDoc, exists := c.duplicateMap[hash]
	if exists {
		// Đã tìm thấy tài liệu trùng lặp, quyết định giữ lại tài liệu nào
		// Đếm số trường rỗng trong tài liệu hiện tại
		currentEmptyFields := countEmptyFields(doc)

		// Đếm số trường rỗng trong tài liệu đã tồn tại
		existingEmptyFields := countEmptyFields(existingDoc)

		if currentEmptyFields < existingEmptyFields {
			// Tài liệu hiện tại có ít trường rỗng hơn, thay thế tài liệu cũ
			c.duplicateMap[hash] = doc
			return false // Không coi là trùng lặp để giữ lại tài liệu hiện tại
		}

		// Tài liệu hiện tại có nhiều trường rỗng hơn hoặc bằng, coi là trùng lặp
		c.duplicateCount++
		return true
	}

	// Chưa có tài liệu trùng lặp, thêm vào map
	c.duplicateMap[hash] = doc
	return false
}

// countEmptyFields đếm số lượng trường thông tin rỗng trong tài liệu
func countEmptyFields(doc models.Document) int {
	emptyCount := 0

	if doc.Size == "" {
		emptyCount++
	}

	if doc.Downloads == "" {
		emptyCount++
	}

	if doc.Modified == "" {
		emptyCount++
	}

	if doc.UploadedBy == "" {
		emptyCount++
	}

	return emptyCount
}

// DownloadDocuments tải xuống tất cả các tài liệu
func (c *Crawler) DownloadDocuments() error {
	log.Println("Bắt đầu tải các tài liệu...")

	// Kiểm tra xem có tài liệu để tải không
	if len(c.documents) == 0 {
		return errors.New("không có tài liệu để tải xuống")
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.maxConcurrent)

	// Đếm tổng số tài liệu
	var totalDocs int
	for _, docs := range c.documents {
		totalDocs += len(docs)
	}

	if totalDocs == 0 {
		return errors.New("không có tài liệu để tải xuống")
	}

	log.Printf("Tổng cộng có %d tài liệu cần tải xuống", totalDocs)

	// Theo dõi tiến độ tải xuống
	var downloadedDocs int
	var downloadMutex sync.Mutex

	// Đặt lại biến đếm trùng lặp
	c.duplicateCount = 0
	// Đặt lại map phát hiện trùng lặp để bắt đầu mới
	c.duplicateMap = make(map[string]models.Document)

	// Tải tài liệu theo danh mục
	for category, docs := range c.documents {
		log.Printf("Đang tải %d tài liệu từ danh mục %s", len(docs), category)

		for i, doc := range docs {
			// Kiểm tra trùng lặp và quyết định giữ lại tài liệu nào
			isDup := c.isDuplicate(doc)
			if isDup {
				log.Printf("Phát hiện tài liệu trùng lặp, bỏ qua: %s", doc.Name)
				downloadMutex.Lock()
				downloadedDocs++
				progress := float64(downloadedDocs) / float64(totalDocs) * 100
				log.Printf("Tiến độ: %.1f%% (%d/%d)", progress, downloadedDocs, totalDocs)
				downloadMutex.Unlock()
				continue
			}

			wg.Add(1)
			semaphore <- struct{}{} // Lấy token

			go func(index int, document models.Document) {
				defer wg.Done()
				defer func() { <-semaphore }() // Trả lại token

				destPath := filepath.Join(c.documentsDir, document.FilePath)

				// Nếu tệp đã tồn tại, bỏ qua tải xuống
				if _, err := os.Stat(destPath); err == nil {
					log.Printf("Tệp đã tồn tại, bỏ qua tải xuống: %s", destPath)
					downloadMutex.Lock()
					downloadedDocs++
					progress := float64(downloadedDocs) / float64(totalDocs) * 100
					log.Printf("Tiến độ: %.1f%% (%d/%d)", progress, downloadedDocs, totalDocs)
					downloadMutex.Unlock()
					return
				}

				// Tải tệp
				log.Printf("Đang tải: %s", document.Name)
				if err := utils.DownloadFile(document.DownloadURL, destPath); err != nil {
					log.Printf("Lỗi khi tải tệp %s: %v", document.Name, err)
					return
				}

				downloadMutex.Lock()
				downloadedDocs++
				progress := float64(downloadedDocs) / float64(totalDocs) * 100
				log.Printf("Tiến độ: %.1f%% (%d/%d)", progress, downloadedDocs, totalDocs)
				downloadMutex.Unlock()

				log.Printf("Đã tải xong: %s", document.Name)
			}(i, doc)
		}
	}

	// Đợi tất cả tải xuống hoàn tất
	wg.Wait()
	log.Printf("Đã tải xuống tất cả tài liệu. Tổng số: %d", totalDocs)
	log.Printf("Phát hiện %d tài liệu trùng lặp trong cơ sở dữ liệu", c.duplicateCount)

	// Cập nhật lại danh sách tài liệu để loại bỏ các bản trùng lặp
	c.updateDocumentsFromDuplicateMap()

	return nil
}

// updateDocumentsFromDuplicateMap cập nhật lại danh sách tài liệu từ map trùng lặp
func (c *Crawler) updateDocumentsFromDuplicateMap() {
	// Tạo map mới để lưu trữ tài liệu đã lọc
	newDocuments := make(map[string][]models.Document)

	// Duyệt qua tất cả tài liệu trong map trùng lặp
	for _, doc := range c.duplicateMap {
		category := doc.Category
		newDocuments[category] = append(newDocuments[category], doc)
	}

	// Cập nhật lại danh sách tài liệu
	c.documents = newDocuments

	log.Printf("Đã cập nhật lại danh sách tài liệu sau khi loại bỏ %d bản trùng lặp", c.duplicateCount)
}

// GetDocuments trả về tất cả tài liệu đã trích xuất
func (c *Crawler) GetDocuments() map[string][]models.Document {
	return c.documents
}

// GetAllDocuments trả về danh sách phẳng của tất cả tài liệu
func (c *Crawler) GetAllDocuments() []models.Document {
	var allDocs []models.Document
	for _, docs := range c.documents {
		allDocs = append(allDocs, docs...)
	}
	return allDocs
}
