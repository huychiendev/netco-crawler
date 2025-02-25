// JavaScript chính cho Netco Crawler

document.addEventListener('DOMContentLoaded', function() {
    // Xử lý tìm kiếm trong bảng
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('keyup', function() {
            const searchValue = this.value.toLowerCase();
            const tableRows = document.querySelectorAll('table tbody tr');
            
            tableRows.forEach(row => {
                const text = row.textContent.toLowerCase();
                if (text.includes(searchValue)) {
                    row.style.display = '';
                } else {
                    row.style.display = 'none';
                }
            });
        });
    }
    
    // Xử lý sắp xếp bảng
    const tableHeaders = document.querySelectorAll('th[data-sort]');
    tableHeaders.forEach(header => {
        header.addEventListener('click', function() {
            const column = this.dataset.sort;
            const table = this.closest('table');
            const tbody = table.querySelector('tbody');
            const rows = Array.from(tbody.querySelectorAll('tr'));
            
            // Chuyển đổi hướng sắp xếp
            const direction = this.classList.contains('sort-asc') ? 'desc' : 'asc';
            
            // Loại bỏ các lớp sắp xếp từ tất cả các tiêu đề
            tableHeaders.forEach(th => {
                th.classList.remove('sort-asc', 'sort-desc');
            });
            
            // Thêm lớp sắp xếp vào tiêu đề hiện tại
            this.classList.add(`sort-${direction}`);
            
            // Sắp xếp các hàng
            rows.sort((a, b) => {
                const aValue = a.querySelector(`td:nth-child(${column})`).textContent.trim();
                const bValue = b.querySelector(`td:nth-child(${column})`).textContent.trim();
                
                return direction === 'asc' ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
            });
            
            // Xóa các hàng hiện tại
            rows.forEach(row => {
                tbody.removeChild(row);
            });
            
            // Thêm các hàng đã sắp xếp
            rows.forEach(row => {
                tbody.appendChild(row);
            });
        });
    });
    
    // Kiểm tra thông báo thành công
    const successMessage = document.getElementById('successMessage');
    if (successMessage) {
        setTimeout(() => {
            successMessage.style.opacity = '0';
            setTimeout(() => {
                successMessage.style.display = 'none';
            }, 500);
        }, 3000);
    }
}); 