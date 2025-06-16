function initializeCharts(resultsData, summaryData) {
    console.log('initializeCharts called with data:', { resultsData: resultsData.length, summaryData });
    
    if (typeof Chart === 'undefined') {
        console.warn('Chart.js not loaded yet, retrying...');
        setTimeout(() => initializeCharts(resultsData, summaryData), 100);
        return;
    }
    
    console.log('Chart.js is available');
    
    // Hide loading states
    const loadingElements = document.querySelectorAll('[id$="-loading"]');
    loadingElements.forEach(el => el.style.display = 'none');
    
    if (!summaryData || Object.keys(summaryData).length === 0) {
        console.warn('No summaryData available, skipping chart initialization');
        return;
    }
    
    // Debug logging
    console.log('About to initialize charts with summaryData:', summaryData);
    
    try {
        initializeDistributionChart(summaryData);
        console.log('Distribution chart initialized');
    } catch (error) {
        console.error('Error initializing distribution chart:', error);
        console.error('Error details:', error.message, error.stack);
    }
    
    try {
        initializeAverageScoresChart(summaryData);
        console.log('Average scores chart initialized');
    } catch (error) {
        console.error('Error initializing average scores chart:', error);
        console.error('Error details:', error.message, error.stack);
    }
}

function initializeDistributionChart(data) {
    console.log('initializeDistributionChart called');
    const ctx = document.getElementById('scoreDistributionChart');
    console.log('Canvas element found:', !!ctx);
    
    if (!ctx) {
        console.error('scoreDistributionChart canvas not found');
        return;
    }
    
    // Set explicit canvas dimensions for crisp rendering
    const container = ctx.parentElement;
    const containerWidth = container.clientWidth - 32; // Account for padding
    const containerHeight = 300; // Fixed height
    
    ctx.width = containerWidth * window.devicePixelRatio;
    ctx.height = containerHeight * window.devicePixelRatio;
    ctx.style.width = containerWidth + 'px';
    ctx.style.height = containerHeight + 'px';
    
    console.log('About to access ScoreDistribution...');
    console.log('ScoreDistribution object:', data.ScoreDistribution);
    
    if (!data.ScoreDistribution) {
        console.warn('ScoreDistribution not available in summaryData');
        return;
    }
    
    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: ['Performance', 'Accessibility', 'Best Practices', 'SEO'],
            datasets: [{
                label: 'Excellent (90-100)',
                data: [
                    data.ScoreDistribution["performance"]?.[0] || 0,
                    data.ScoreDistribution["accessibility"]?.[0] || 0,
                    data.ScoreDistribution["best_practices"]?.[0] || 0,
                    data.ScoreDistribution["seo"]?.[0] || 0
                ],
                backgroundColor: 'rgba(16, 185, 129, 0.8)',
                borderColor: 'rgba(16, 185, 129, 1)',
                borderWidth: 1,
                borderRadius: 4
            }, {
                label: 'Needs Improvement (50-89)',
                data: [
                    data.ScoreDistribution["performance"]?.[1] || 0,
                    data.ScoreDistribution["accessibility"]?.[1] || 0,
                    data.ScoreDistribution["best_practices"]?.[1] || 0,
                    data.ScoreDistribution["seo"]?.[1] || 0
                ],
                backgroundColor: 'rgba(251, 191, 36, 0.8)',
                borderColor: 'rgba(251, 191, 36, 1)',
                borderWidth: 1,
                borderRadius: 4
            }, {
                label: 'Poor (0-49)',
                data: [
                    data.ScoreDistribution["performance"]?.[2] || 0,
                    data.ScoreDistribution["accessibility"]?.[2] || 0,
                    data.ScoreDistribution["best_practices"]?.[2] || 0,
                    data.ScoreDistribution["seo"]?.[2] || 0
                ],
                backgroundColor: 'rgba(239, 68, 68, 0.8)',
                borderColor: 'rgba(239, 68, 68, 1)',
                borderWidth: 1,
                borderRadius: 4
            }]
        },
        options: getBarChartOptions()
    });
}

function initializeAverageScoresChart(data) {
    console.log('initializeAverageScoresChart called');
    console.log('About to access AverageScores...');
    console.log('AverageScores object:', data.AverageScores);
    
    const ctx = document.getElementById('averageScoresChart');
    
    if (!ctx) {
        console.error('averageScoresChart canvas not found');
        return;
    }
    
    // Set explicit canvas dimensions for crisp rendering
    const container = ctx.parentElement;
    const containerWidth = container.clientWidth - 32;
    const containerHeight = 300;
    
    ctx.width = containerWidth * window.devicePixelRatio;
    ctx.height = containerHeight * window.devicePixelRatio;
    ctx.style.width = containerWidth + 'px';
    ctx.style.height = containerHeight + 'px';
    
    if (!data.AverageScores) {
        console.warn('AverageScores not available in summaryData');
        return;
    }
    
    new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: ['Performance', 'Accessibility', 'Best Practices', 'SEO'],
            datasets: [{
                data: [
                    data.AverageScores["performance"] || 0,
                    data.AverageScores["accessibility"] || 0,
                    data.AverageScores["best_practices"] || 0,
                    data.AverageScores["seo"] || 0
                ],
                backgroundColor: [
                    'rgba(59, 130, 246, 0.8)',
                    'rgba(139, 92, 246, 0.8)', 
                    'rgba(6, 182, 212, 0.8)',
                    'rgba(16, 185, 129, 0.8)'
                ],
                borderColor: [
                    'rgba(59, 130, 246, 1)',
                    'rgba(139, 92, 246, 1)',
                    'rgba(6, 182, 212, 1)', 
                    'rgba(16, 185, 129, 1)'
                ],
                borderWidth: 2,
                hoverOffset: 8
            }]
        },
        options: getDoughnutChartOptions()
    });
}

function getBarChartOptions() {
    return {
        responsive: true,
        maintainAspectRatio: false,
        devicePixelRatio: window.devicePixelRatio,
        interaction: {
            intersect: false,
            mode: 'index'
        },
        layout: {
            padding: {
                top: 20,
                bottom: 20,
                left: 20,
                right: 20
            }
        },
        scales: {
            x: {
                stacked: true,
                ticks: { 
                    color: 'rgba(255, 255, 255, 0.8)',
                    font: {
                        size: 12,
                        family: 'Inter, system-ui, sans-serif'
                    }
                },
                grid: { 
                    color: 'rgba(255, 255, 255, 0.1)',
                    drawBorder: false
                },
                border: {
                    display: false
                }
            },
            y: {
                stacked: true,
                ticks: { 
                    color: 'rgba(255, 255, 255, 0.8)',
                    font: {
                        size: 12,
                        family: 'Inter, system-ui, sans-serif'
                    }
                },
                grid: { 
                    color: 'rgba(255, 255, 255, 0.1)',
                    drawBorder: false
                },
                border: {
                    display: false
                }
            }
        },
        plugins: {
            legend: {
                position: 'top',
                align: 'center',
                maxHeight: 60,
                labels: { 
                    color: 'rgba(255, 255, 255, 0.9)',
                    font: {
                        size: 12,
                        family: 'Inter, system-ui, sans-serif',
                        weight: '500'
                    },
                    padding: 12,
                    boxWidth: 12,
                    boxHeight: 12,
                    usePointStyle: true,
                    pointStyle: 'rectRounded',
                    generateLabels: function(chart) {
                        const datasets = chart.data.datasets;
                        return datasets.map((dataset, i) => ({
                            text: dataset.label,
                            fillStyle: dataset.backgroundColor,
                            strokeStyle: dataset.borderColor,
                            lineWidth: dataset.borderWidth,
                            hidden: !chart.isDatasetVisible(i),
                            index: i
                        }));
                    }
                }
            },
            tooltip: {
                backgroundColor: 'rgba(17, 24, 39, 0.95)',
                titleColor: 'rgba(255, 255, 255, 0.9)',
                bodyColor: 'rgba(255, 255, 255, 0.8)',
                borderColor: 'rgba(75, 85, 99, 0.5)',
                borderWidth: 1,
                cornerRadius: 8,
                displayColors: true,
                titleFont: {
                    size: 14,
                    family: 'Inter, system-ui, sans-serif',
                    weight: '600'
                },
                bodyFont: {
                    size: 13,
                    family: 'Inter, system-ui, sans-serif'
                }
            }
        }
    };
}

function getDoughnutChartOptions() {
    return {
        responsive: true,
        maintainAspectRatio: false,
        devicePixelRatio: window.devicePixelRatio,
        layout: {
            padding: {
                top: 20,
                bottom: 20,
                left: 20,
                right: 20
            }
        },
        plugins: {
            legend: {
                position: 'bottom',
                align: 'center',
                labels: { 
                    color: 'rgba(255, 255, 255, 0.9)',
                    font: {
                        size: 13,
                        family: 'Inter, system-ui, sans-serif',
                        weight: '500'
                    },
                    padding: 15,
                    usePointStyle: true,
                    pointStyle: 'circle'
                }
            },
            tooltip: {
                backgroundColor: 'rgba(17, 24, 39, 0.95)',
                titleColor: 'rgba(255, 255, 255, 0.9)',
                bodyColor: 'rgba(255, 255, 255, 0.8)',
                borderColor: 'rgba(75, 85, 99, 0.5)',
                borderWidth: 1,
                cornerRadius: 8,
                displayColors: true,
                titleFont: {
                    size: 14,
                    family: 'Inter, system-ui, sans-serif',
                    weight: '600'
                },
                bodyFont: {
                    size: 13,
                    family: 'Inter, system-ui, sans-serif'
                },
                callbacks: {
                    label: function(context) {
                        return context.label + ': ' + Math.round(context.parsed) + '/100';
                    }
                }
            }
        },
        cutout: '60%',
        elements: {
            arc: {
                borderJoinStyle: 'round'
            }
        }
    };
}

// Add resize handler for responsive charts
function handleChartResize() {
    const charts = Chart.instances;
    Object.values(charts).forEach(chart => {
        if (chart && chart.canvas) {
            const container = chart.canvas.parentElement;
            const containerWidth = container.clientWidth - 32;
            const containerHeight = 300;
            
            chart.canvas.width = containerWidth * window.devicePixelRatio;
            chart.canvas.height = containerHeight * window.devicePixelRatio;
            chart.canvas.style.width = containerWidth + 'px';
            chart.canvas.style.height = containerHeight + 'px';
            
            chart.resize();
        }
    });
}

// Initialize resize handler
window.addEventListener('resize', debounce(handleChartResize, 250));

// Debounce function for performance
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}
