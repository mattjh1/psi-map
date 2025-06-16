document.addEventListener('DOMContentLoaded', () => {
    updateCounters();
    initializePagination();
});

// Updates the Total, Successful, and Failed counters
function updateCounters() {
    const rows = document.querySelectorAll('#resultsTable tbody tr');
    let total = rows.length;
    let successful = 0;
    let failed = 0;

    rows.forEach(row => {
        const errorCell = row.querySelector('td[colspan="6"]');
        if (errorCell) {
            failed++;
        } else {
            successful++;
        }
    });

    document.getElementById('totalCount').textContent = total;
    document.getElementById('successCount').textContent = successful;
    document.getElementById('failedCount').textContent = failed;
}

// Sorts the table by the specified column
function sortTable(column) {
    const table = document.getElementById('resultsTable');
    const tbody = table.querySelector('tbody');
    const rows = Array.from(tbody.querySelectorAll('tr'));
    let isAscending = table.dataset.sortDirection !== 'asc';
    table.dataset.sortDirection = isAscending ? 'asc' : 'desc';

    rows.sort((rowA, rowB) => {
        let valueA, valueB;
        switch (column) {
            case 'url':
                valueA = rowA.dataset.url.toLowerCase();
                valueB = rowB.dataset.url.toLowerCase();
                break;
            case 'strategy':
                valueA = rowA.dataset.strategy.toLowerCase();
                valueB = rowB.dataset.strategy.toLowerCase();
                break;
            case 'performance':
            case 'accessibility':
            case 'best-practices':
            case 'seo':
                valueA = parseFloat(rowA.querySelector(`td:nth-child(${getColumnIndex(column)})`).textContent) || 0;
                valueB = parseFloat(rowB.querySelector(`td:nth-child(${getColumnIndex(column)})`).textContent) || 0;
                break;
            case 'loadtime':
                valueA = parseDuration(rowA.querySelector(`td:nth-child(${getColumnIndex(column)}) span`).textContent);
                valueB = parseDuration(rowB.querySelector(`td:nth-child(${getColumnIndex(column)}) span`).textContent);
                break;
            default:
                return 0;
        }

        if (valueA < valueB) return isAscending ? -1 : 1;
        if (valueA > valueB) return isAscending ? 1 : -1;
        return 0;
    });

    tbody.innerHTML = '';
    rows.forEach(row => tbody.appendChild(row));
    updatePagination();
}

// Helper to get column index based on column name
function getColumnIndex(column) {
    const headers = ['url', 'strategy', 'performance', 'accessibility', 'best-practices', 'seo', 'loadtime', 'core-web-vitals', 'actions'];
    return headers.indexOf(column) + 1;
}

// Parses duration string (e.g., "1.5s") to seconds
function parseDuration(str) {
    return parseFloat(str.replace('s', '')) || 0;
}

// Exports table data as CSV
function exportResults() {
    const rows = document.querySelectorAll('#resultsTable tr');
    let csv = [];
    
    // Headers
    const headers = Array.from(rows[0].querySelectorAll('th')).map(th => th.textContent.trim());
    csv.push(headers.join(','));

    // Rows
    Array.from(rows).slice(1).forEach(row => {
        const cells = Array.from(row.querySelectorAll('td')).map(cell => {
            let text = cell.textContent.trim();
            // Handle special cases
            if (cell.querySelector('.inline-flex')) {
                text = cell.querySelector('.inline-flex').textContent.trim();
            }
            // Escape commas and quotes
            return `"${text.replace(/"/g, '""')}"`;
        });
        csv.push(cells.join(','));
    });

    // Download CSV
    const csvContent = csv.join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'results.csv';
    a.click();
    window.URL.revokeObjectURL(url);
}

// Refreshes results (placeholder for AJAX or page reload)
function refreshResults() {
    document.getElementById('resultsTable').closest('.glass-card').querySelector('#loadingState').classList.remove('hidden');
    document.getElementById('resultsTable').closest('.overflow-x-auto').classList.add('hidden');
    
    // Simulate refresh (replace with actual API call if needed)
    setTimeout(() => {
        window.location.reload(); // Or fetch new data via AJAX
    }, 1000);
}

// Clears all filters and shows all results
function clearAllFilters() {
    const rows = document.querySelectorAll('#resultsTable tbody tr');
    rows.forEach(row => row.classList.remove('hidden'));
    document.getElementById('emptyState').classList.add('hidden');
    document.getElementById('resultsTable').closest('.overflow-x-auto').classList.remove('hidden');
    updateCounters();
    updatePagination();
}

// Initializes pagination
function initializePagination() {
    const rowsPerPage = 10;
    const rows = document.querySelectorAll('#resultsTable tbody tr');
    const totalPages = Math.ceil(rows.length / rowsPerPage);
    
    updatePaginationDisplay(1, rowsPerPage, rows.length);
    renderPageNumbers(totalPages, 1);
    showPage(1, rowsPerPage);
}

// Updates pagination display
function updatePaginationDisplay(currentPage, rowsPerPage, totalRows) {
    const start = (currentPage - 1) * rowsPerPage + 1;
    const end = Math.min(currentPage * rowsPerPage, totalRows);
    
    document.getElementById('pageStart').textContent = start;
    document.getElementById('pageEnd').textContent = end;
    document.getElementById('pageTotal').textContent = totalRows;
    document.getElementById('prevPage').disabled = currentPage === 1;
    document.getElementById('nextPage').disabled = end >= totalRows;
}

// Renders page number buttons
function renderPageNumbers(totalPages, currentPage) {
    const pageNumbers = document.getElementById('pageNumbers');
    pageNumbers.innerHTML = '';
    
    for (let i = 1; i <= totalPages; i++) {
        const button = document.createElement('button');
        button.className = `px-3 py-2 rounded-lg text-sm ${i === currentPage ? 'bg-blue-500/20 text-blue-300' : 'glass-effect text-white/70 hover:text-white hover:bg-white/10'} transition-all duration-200`;
        button.textContent = i;
        button.onclick = () => changePageTo(i);
        pageNumbers.appendChild(button);
    }
}

// Changes to a specific page
function changePageTo(page) {
    const rowsPerPage = 10;
    showPage(page, rowsPerPage);
    updatePaginationDisplay(page, rowsPerPage, document.querySelectorAll('#resultsTable tbody tr').length);
    renderPageNumbers(Math.ceil(document.querySelectorAll('#resultsTable tbody tr').length / rowsPerPage), page);
}

// Changes page relative to current
function changePage(delta) {
    const currentPage = parseInt(document.querySelector('#pageNumbers button.bg-blue-500\\/20')?.textContent || '1');
    const newPage = currentPage + delta;
    if (newPage > 0 && newPage <= Math.ceil(document.querySelectorAll('#resultsTable tbody tr').length / 10)) {
        changePageTo(newPage);
    }
}

// Shows specific page
function showPage(page, rowsPerPage) {
    const rows = document.querySelectorAll('#resultsTable tbody tr');
    const start = (page - 1) * rowsPerPage;
    const end = start + rowsPerPage;
    
    rows.forEach((row, index) => {
        row.classList.toggle('hidden', index < start || index >= end);
    });
    
    document.getElementById('emptyState').classList.toggle('hidden', rows.length > 0);
    document.getElementById('resultsTable').closest('.overflow-x-auto').classList.toggle('hidden', rows.length === 0);
}
