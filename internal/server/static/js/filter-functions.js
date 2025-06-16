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
            filterTable(strategy);

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
                const activeTab = document.querySelector('.strategy-filter-btn.active');
                const activeStrategy = activeTab ? activeTab.getAttribute('data-strategy') : 'all';
                
                filterTable(activeStrategy, searchTerm);
                clearButton.classList.toggle('hidden', searchTerm === '');
                
                // Update search results text
                updateSearchResultsText(searchTerm);
            }, 300);
        });
        
        clearButton.addEventListener('click', clearSearch);
    }
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

function filterTable(strategy, searchTerm = '') {
    const rows = document.querySelectorAll('#resultsTable tbody tr');
    let visibleCount = 0;

    rows.forEach(row => {
        const rowStrategy = row.getAttribute('data-strategy');
        
        // Get URL from data attribute OR from the first cell's text content
        let rowUrl = row.getAttribute('data-url');
        if (!rowUrl) {
            // Fallback: get URL from first cell if data-url doesn't exist
            const firstCell = row.querySelector('td');
            rowUrl = firstCell ? firstCell.textContent : '';
        }
        rowUrl = rowUrl.toLowerCase();
        
        // Also search in all visible text content of the row
        const rowText = row.textContent.toLowerCase();

        const strategyMatch = strategy === 'all' || rowStrategy === strategy;
        
        // Search in URL and all row text content
        const searchMatch = searchTerm === '' || 
                           rowUrl.includes(searchTerm) || 
                           rowText.includes(searchTerm);

        if (strategyMatch && searchMatch) {
            row.classList.remove('hidden');
            row.classList.add('animate-fade-in');
            visibleCount++;
        } else {
            row.classList.add('hidden');
            row.classList.remove('animate-fade-in');
        }
    });

    // Update results counter if it exists
    updateResultsCounter(visibleCount);

    // Update filter counters
    updateFilterCounters();

    // Show/hide empty state
    toggleEmptyState(visibleCount === 0);
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

    document.getElementById('count-all').textContent = (mobileCount + desktopCount).toString();
    document.getElementById('count-mobile').textContent = mobileCount.toString();
    document.getElementById('count-desktop').textContent = desktopCount.toString();
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
        searchInput.focus(); // Keep focus on input after clearing
        
        const activeTab = document.querySelector('.strategy-filter-btn.active');
        const activeStrategy = activeTab ? activeTab.getAttribute('data-strategy') : 'all';
        filterTable(activeStrategy);
        
        // Hide clear button
        if (clearButton) {
            clearButton.classList.add('hidden');
        }
        
        // Reset search results text
        updateSearchResultsText('');
    }
}

document.addEventListener('DOMContentLoaded', function() {
    initializeFilters();
    initializeSearch();
    updateFilterCounters();
    initializeCtrlK();
    filterTable('all'); // Set initial filter to show all results
});
