// filters-functions.js - Enhanced version with advanced filters
function initializeFilters() {
    const tabs = document.querySelectorAll('.strategy-filter-btn');
    tabs.forEach(tab => {
        tab.addEventListener('click', function(e) {
            e.preventDefault();
            // Update active tab styles
            tabs.forEach(t => {
                t.classList.remove('active', 'bg-blue-500/20', 'text-blue-300', 'border-blue-500/30');
                t.classList.add('bg-white/5', 'text-white/70', 'border-white/20');
            });
            // Add active styles to clicked tab
            this.classList.add('active', 'bg-blue-500/20', 'text-blue-300', 'border-blue-500/30');
            this.classList.remove('bg-white/5', 'text-white/70', 'border-white/20');
            // Filter table rows
            const strategy = this.getAttribute('data-strategy');
            applyAllFilters();
            // Update summary cards if they exist
            updateSummaryCards();
        });
    });
}

function initializeSearch() {
    const searchInput = document.getElementById('searchInput');
    const clearButton = document.getElementById('clearSearch');
    
    if (searchInput && clearButton) {
        // Add debouncing for better performance
        let debounceTimer;
        
        searchInput.addEventListener('input', function() {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                const searchTerm = this.value.toLowerCase().trim();
                applyAllFilters();
                clearButton.classList.toggle('hidden', searchTerm === '');
                
                // Update search results text
                updateSearchResultsText(searchTerm);
            }, 300);
        });
        
        clearButton.addEventListener('click', clearSearch);
    }
}

function initializeAdvancedFilters() {
    const performanceFilter = document.getElementById('performanceFilter');
    const statusFilter = document.getElementById('statusFilter');
    const sortFilter = document.getElementById('sortFilter');
    
    // Add event listeners for advanced filters
    [performanceFilter, statusFilter, sortFilter].forEach(filter => {
        if (filter) {
            filter.addEventListener('change', function() {
                applyAllFilters();
                // Add visual feedback
                filter.classList.add('animate-pulse');
                setTimeout(() => filter.classList.remove('animate-pulse'), 300);
            });
        }
    });
}

function initializeCtrlK() {
    document.addEventListener('keydown', function(e) {
        // Ctrl+K or Cmd+K to focus search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            const searchInput = document.getElementById('searchInput');
            if (searchInput) {
                searchInput.focus();
                searchInput.select(); // Select all text for easy replacement
            }
        }
        
        // Escape key to clear search when input is focused
        if (e.key === 'Escape') {
            const searchInput = document.getElementById('searchInput');
            if (searchInput && document.activeElement === searchInput) {
                clearSearch();
            }
        }
    });
}

// NEW: Toggle advanced filters panel
function toggleAdvancedFilters() {
    const advancedFilters = document.getElementById('advancedFilters');
    const toggleIcon = document.getElementById('advancedToggleIcon');
    
    if (advancedFilters && toggleIcon) {
        const isHidden = advancedFilters.classList.contains('hidden');
        
        if (isHidden) {
            // Show advanced filters
            advancedFilters.classList.remove('hidden');
            advancedFilters.classList.add('animate-slide-down');
            toggleIcon.classList.add('rotate-180');
        } else {
            // Hide advanced filters
            advancedFilters.classList.add('animate-slide-up');
            toggleIcon.classList.remove('rotate-180');
            
            // Hide after animation
            setTimeout(() => {
                advancedFilters.classList.add('hidden');
                advancedFilters.classList.remove('animate-slide-up', 'animate-slide-down');
            }, 200);
        }
    }
}

// NEW: Apply all filters (strategy, search, and advanced filters)
function applyAllFilters() {
    const rows = document.querySelectorAll('#resultsTable tbody tr');
    const searchInput = document.getElementById('searchInput');
    const activeTab = document.querySelector('.strategy-filter-btn.active');
    
    // Get current filter values
    const strategy = activeTab ? activeTab.getAttribute('data-strategy') : 'all';
    const searchTerm = searchInput ? searchInput.value.toLowerCase().trim() : '';
    const performanceFilter = document.getElementById('performanceFilter')?.value || 'all';
    const statusFilter = document.getElementById('statusFilter')?.value || 'all';
    
    let visibleCount = 0;
    let filteredResults = [];
    
    rows.forEach(row => {
        // Strategy filter
        const rowStrategy = row.getAttribute('data-strategy');
        const strategyMatch = strategy === 'all' || rowStrategy === strategy;
        
        // Search filter
        let rowUrl = row.getAttribute('data-url') || '';
        if (!rowUrl) {
            const firstCell = row.querySelector('td');
            rowUrl = firstCell ? firstCell.textContent : '';
        }
        const rowText = row.textContent.toLowerCase();
        const searchMatch = searchTerm === '' || 
                           rowUrl.toLowerCase().includes(searchTerm) || 
                           rowText.includes(searchTerm);
        
        // Performance filter
        const performanceMatch = checkPerformanceFilter(row, performanceFilter);
        
        // Status filter
        const statusMatch = checkStatusFilter(row, statusFilter);
        
        // Apply all filters
        if (strategyMatch && searchMatch && performanceMatch && statusMatch) {
            row.classList.remove('hidden');
            row.classList.add('animate-fade-in');
            visibleCount++;
            filteredResults.push(row);
        } else {
            row.classList.add('hidden');
            row.classList.remove('animate-fade-in');
        }
    });
    
    // Apply sorting
    applySorting(filteredResults);
    
    // Update counters and UI
    updateResultsCounter(visibleCount);
    updateFilterCounters();
    toggleEmptyState(visibleCount === 0);
}

// NEW: Check performance filter
function checkPerformanceFilter(row, performanceFilter) {
    if (performanceFilter === 'all') return true;
    
    // Try to find performance score in the row
    // This assumes you have performance data in data attributes or specific cells
    const performanceScore = getPerformanceScore(row);
    
    if (performanceScore === null) return true; // If no score found, show row
    
    switch (performanceFilter) {
        case 'excellent':
            return performanceScore >= 90;
        case 'good':
            return performanceScore >= 50 && performanceScore < 90;
        case 'poor':
            return performanceScore < 50;
        default:
            return true;
    }
}

// NEW: Check status filter
function checkStatusFilter(row, statusFilter) {
    if (statusFilter === 'all') return true;
    
    // Try to find status in the row
    const status = getRowStatus(row);
    
    switch (statusFilter) {
        case 'success':
            return status === 'success' || status === 'successful';
        case 'error':
            return status === 'error' || status === 'failed';
        default:
            return true;
    }
}

// NEW: Get performance score from row
function getPerformanceScore(row) {
    // Method 1: Check data attribute
    const scoreAttr = row.getAttribute('data-performance');
    if (scoreAttr) return parseInt(scoreAttr);
    
    // Method 2: Look for score in specific cell (adjust selector as needed)
    const scoreCell = row.querySelector('[data-performance-score]');
    if (scoreCell) return parseInt(scoreCell.textContent);
    
    // Method 3: Look for score pattern in text content
    const scoreMatch = row.textContent.match(/(\d+)%?/);
    if (scoreMatch) return parseInt(scoreMatch[1]);
    
    return null; // No score found
}

// NEW: Get row status
function getRowStatus(row) {
    // Method 1: Check data attribute
    const statusAttr = row.getAttribute('data-status');
    if (statusAttr) return statusAttr.toLowerCase();
    
    // Method 2: Look for status indicators in classes
    if (row.classList.contains('success') || row.querySelector('.success')) return 'success';
    if (row.classList.contains('error') || row.querySelector('.error')) return 'error';
    
    // Method 3: Look for status text patterns
    const rowText = row.textContent.toLowerCase();
    if (rowText.includes('success') || rowText.includes('✓')) return 'success';
    if (rowText.includes('error') || rowText.includes('failed') || rowText.includes('✗')) return 'error';
    
    return 'unknown';
}

// NEW: Apply sorting to visible rows
function applySorting(visibleRows) {
    const sortFilter = document.getElementById('sortFilter');
    if (!sortFilter || visibleRows.length === 0) return;
    
    const sortBy = sortFilter.value;
    const tbody = document.querySelector('#resultsTable tbody');
    if (!tbody) return;
    
    // Sort the visible rows
    const sortedRows = visibleRows.sort((a, b) => {
        switch (sortBy) {
            case 'url':
                const urlA = getRowUrl(a).toLowerCase();
                const urlB = getRowUrl(b).toLowerCase();
                return urlA.localeCompare(urlB);
            
            case 'performance':
                const perfA = getPerformanceScore(a) || 0;
                const perfB = getPerformanceScore(b) || 0;
                return perfB - perfA; // Descending order
            
            case 'accessibility':
                const accA = getAccessibilityScore(a) || 0;
                const accB = getAccessibilityScore(b) || 0;
                return accB - accA; // Descending order
            
            case 'loadtime':
                const loadA = getLoadTime(a) || 0;
                const loadB = getLoadTime(b) || 0;
                return loadA - loadB; // Ascending order (faster first)
            
            default:
                return 0;
        }
    });
    
    // Re-append sorted rows to maintain order
    sortedRows.forEach(row => {
        tbody.appendChild(row);
    });
}

// NEW: Helper functions for sorting
function getRowUrl(row) {
    return row.getAttribute('data-url') || row.querySelector('td')?.textContent || '';
}

function getAccessibilityScore(row) {
    const accAttr = row.getAttribute('data-accessibility');
    if (accAttr) return parseInt(accAttr);
    
    const accCell = row.querySelector('[data-accessibility-score]');
    if (accCell) return parseInt(accCell.textContent);
    
    return null;
}

function getLoadTime(row) {
    const loadAttr = row.getAttribute('data-loadtime');
    if (loadAttr) return parseFloat(loadAttr);
    
    const loadCell = row.querySelector('[data-load-time]');
    if (loadCell) return parseFloat(loadCell.textContent);
    
    // Look for time pattern (e.g., "2.5s", "1500ms")
    const timeMatch = row.textContent.match(/(\d+\.?\d*)(s|ms)/);
    if (timeMatch) {
        const value = parseFloat(timeMatch[1]);
        const unit = timeMatch[2];
        return unit === 'ms' ? value : value * 1000; // Convert to ms
    }
    
    return null;
}

// Updated existing functions
function filterTable(strategy, searchTerm = '') {
    // This function is now replaced by applyAllFilters()
    // Keeping for backward compatibility
    applyAllFilters();
}

function updateFilterCounters() {
    const rows = document.querySelectorAll('#resultsTable tbody tr');
    let mobileCount = 0;
    let desktopCount = 0;
    
    rows.forEach(row => {
        const strategy = row.getAttribute('data-strategy');
        if (strategy === 'mobile') mobileCount++;
        else if (strategy === 'desktop') desktopCount++;
    });
    
    const countAll = document.getElementById('count-all');
    const countMobile = document.getElementById('count-mobile');
    const countDesktop = document.getElementById('count-desktop');
    
    if (countAll) countAll.textContent = (mobileCount + desktopCount).toString();
    if (countMobile) countMobile.textContent = mobileCount.toString();
    if (countDesktop) countDesktop.textContent = desktopCount.toString();
}

function updateResultsCounter(count) {
    const counter = document.getElementById('resultsCounter');
    if (counter) {
        counter.textContent = `${count} result${count !== 1 ? 's' : ''}`;
        counter.classList.add('animate-bounce-subtle');
        setTimeout(() => counter.classList.remove('animate-bounce-subtle'), 600);
    }
}

function toggleEmptyState(show) {
    const emptyState = document.getElementById('emptyState');
    const resultsTable = document.getElementById('resultsTable');
    
    if (emptyState && resultsTable) {
        if (show) {
            emptyState.classList.remove('hidden');
            resultsTable.classList.add('hidden');
        } else {
            emptyState.classList.add('hidden');
            resultsTable.classList.remove('hidden');
        }
    }
}

function updateSummaryCards() {
    const cards = document.querySelectorAll('[data-summary-card]');
    cards.forEach(card => {
        card.classList.add('animate-slide-up');
        setTimeout(() => card.classList.remove('animate-slide-up'), 300);
    });
}

function updateSearchResultsText(searchTerm) {
    const searchResults = document.getElementById('searchResults');
    if (searchResults) {
        const visibleRows = document.querySelectorAll('#resultsTable tbody tr:not(.hidden)');
        const totalRows = document.querySelectorAll('#resultsTable tbody tr').length;
        
        if (searchTerm) {
            searchResults.textContent = `Showing ${visibleRows.length} of ${totalRows} results for "${searchTerm}"`;
        } else {
            searchResults.textContent = `Showing all ${totalRows} results`;
        }
    }
}

function clearSearch() {
    const searchInput = document.getElementById('searchInput');
    const clearButton = document.getElementById('clearSearch');
    
    if (searchInput) {
        searchInput.value = '';
        searchInput.focus();
        
        applyAllFilters();
        
        if (clearButton) {
            clearButton.classList.add('hidden');
        }
        
        updateSearchResultsText('');
    }
}

// NEW: Reset all filters to default
function resetAllFilters() {
    // Reset strategy filter
    const allTab = document.querySelector('.strategy-filter-btn[data-strategy="all"]');
    if (allTab) {
        allTab.click();
    }
    
    // Reset search
    clearSearch();
    
    // Reset advanced filters
    const performanceFilter = document.getElementById('performanceFilter');
    const statusFilter = document.getElementById('statusFilter');
    const sortFilter = document.getElementById('sortFilter');
    
    if (performanceFilter) performanceFilter.value = 'all';
    if (statusFilter) statusFilter.value = 'all';
    if (sortFilter) sortFilter.value = 'url';
    
    // Apply filters
    applyAllFilters();
}

// Initialize everything when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    initializeFilters();
    initializeSearch();
    initializeAdvancedFilters();
    initializeCtrlK();
    updateFilterCounters();
    applyAllFilters(); // Set initial filter to show all results
    
    // Add global reset button functionality if it exists
    const resetButton = document.getElementById('resetFilters');
    if (resetButton) {
        resetButton.addEventListener('click', resetAllFilters);
    }
});

// Export functions for external use
window.filterFunctions = {
    applyAllFilters,
    resetAllFilters,
    toggleAdvancedFilters
};
