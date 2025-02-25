# Netco Crawler

Công cụ thu thập dữ liệu từ trang web cũ của Netco và hiển thị dưới dạng trang web dễ sử dụng.

## Tính năng

- Thu thập thông tin tài liệu từ 6 trang tài liệu của Netco
- Hỗ trợ phân trang tự động, thu thập từ nhiều trang cho mỗi danh mục
- Tải xuống các tài liệu và lưu trữ cục bộ
- Hiển thị tài liệu dưới dạng trang web đẹp mắt với Tailwind CSS
- Phân loại tài liệu theo danh mục
- Tìm kiếm và lọc tài liệu
- Sắp xếp tài liệu theo nhiều tiêu chí
- Tải tài liệu trực tiếp từ trang web

## Yêu cầu cài đặt

- Go 1.18 trở lên
- Kết nối Internet để tải tài liệu

## Cài đặt

1. Clone dự án:

```
git clone https://github.com/yourusername/netco-crawler.git
cd netco-crawler
```

2. Cài đặt dependencies:

```
go mod download
```

## Sử dụng

### Chạy một lệnh duy nhất (Khuyến nghị)

Để thu thập dữ liệu và khởi động web server trong một lệnh duy nhất, sử dụng:

```
go run cmd/server/main.go
```

hoặc sử dụng script:

```
# Windows
run.bat

# Linux/Mac
./run.sh
```

và chọn tùy chọn 3.

### Thu thập dữ liệu (Chỉ crawler)

Để chỉ thu thập dữ liệu từ trang web Netco và tải xuống tài liệu:

```
go run cmd/crawler/main.go
```

### Khởi động web server (Không thu thập dữ liệu)

Để khởi động web server mà không thu thập dữ liệu:

```
go run cmd/server/main.go --skip-crawl
```

Sau đó, mở trình duyệt và truy cập http://localhost:8080 để xem tất cả tài liệu đã thu thập.

## Cấu trúc dự án

```
netco-crawler/
  ├── cmd/
  │   ├── crawler/      # Ứng dụng thu thập dữ liệu
  │   └── server/       # Web server
  ├── internal/
  │   ├── crawler/      # Logic thu thập dữ liệu
  │   ├── models/       # Định nghĩa dữ liệu
  │   └── utils/        # Tiện ích
  ├── respone/          # Tệp HTML mẫu
  ├── static/
  │   ├── css/          # CSS
  │   ├── js/           # JavaScript
  │   └── documents/    # Tài liệu đã tải xuống
  ├── templates/        # Template HTML
  └── go.mod            # Quản lý dependencies
```

## Danh mục tài liệu

Crawler thu thập dữ liệu từ 6 danh mục sau:

- Báo cáo thường niên (bao-cao-thuong-nien)
- Báo cáo tài chính (bao-cao-tai-chinh)
- Điều lệ công ty (dieu-le-cong-ty)
- Quy chế quản trị công ty (quy-che-quan-tri-cong-ty)
- Công bố thông tin (cong-bao-thong-tin)
- Bản cáo bạch (ban-cao-bach)

## Phân trang

Crawler tự động phát hiện và thu thập dữ liệu từ tất cả các trang của mỗi danh mục, sử dụng cơ chế phân trang của trang web (?pagenumber={number}).

## Giấy phép

Dự án này được phát hành theo giấy phép MIT. 