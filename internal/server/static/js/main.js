console.log('=== SCRIPT START ===');
console.log('JavaScript is loading...');
console.log('Chart.js available immediately:', typeof Chart !== 'undefined');

// Fetch results and summary data from API
async function loadData() {
    try {
        const response = await fetch('/api/report-data');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        return {
            resultsData: data.results,
            summaryData: data.summary
        };
    } catch (error) {
        console.error('Error fetching data:', error);
        return { resultsData: [], summaryData: {} };
    }
}

console.log('Functions after script inclusion:', {
    initializeCharts: typeof initializeCharts,
    initializeDistributionChart: typeof initializeDistributionChart,
    initializeAverageScoresChart: typeof initializeAverageScoresChart
});

// Initialize everything
async function initializeAll() {
    console.log('=== INITIALIZE ALL CALLED ===');
    console.log('Chart.js available:', typeof Chart !== 'undefined');

    if (typeof Chart === 'undefined') {
        console.error('Chart.js is still not available!');
        return;
    }
    
    const { resultsData, summaryData } = await loadData();
    console.log('Data loaded:', { resultsData, summaryData });
    
    if (typeof initializeCharts === 'function') {
        console.log('Calling initializeCharts...');
        initializeCharts(resultsData, summaryData);
    } else {
        console.error('initializeCharts function not found');
    }
    
    if (typeof initializeFilters === 'function') {
        initializeFilters();
    }
    
    if (typeof initializeSearch === 'function') {
        initializeSearch();
    }
}

// Initialize immediately if Chart.js is available, otherwise wait for DOM
if (typeof Chart !== 'undefined') {
    console.log('Chart.js is available, initializing immediately...');
    initializeAll();
} else {
    console.log('Chart.js not available yet, waiting for DOM...');
    document.addEventListener('DOMContentLoaded', function() {
        console.log('DOM loaded, Chart.js available:', typeof Chart !== 'undefined');
        if (typeof Chart !== 'undefined') {
            initializeAll();
        } else {
            console.error('Chart.js still not available after DOM load');
        }
    });
}

console.log('=== SCRIPT END ===');
