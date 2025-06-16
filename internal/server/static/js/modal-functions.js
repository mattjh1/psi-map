    const detailModal = document.getElementById('detailModal');
    const modalCloseButton = document.querySelector('#detailModal [data-modal-close]');
    
    function showModal() {
        detailModal.classList.remove('hidden');
        detailModal.classList.add('flex', 'fade-in');
        document.body.classList.add('overflow-hidden'); // Prevent background scrolling
    }
    
    function hideModal() {
        detailModal.classList.add('hidden');
        detailModal.classList.remove('flex', 'fade-in');
        document.body.classList.remove('overflow-hidden');
    }
    
    // Close modal on button click
    if (modalCloseButton) {
        modalCloseButton.addEventListener('click', hideModal);
    }
    
    // Close modal on outside click
    detailModal.addEventListener('click', (e) => {
        if (e.target === detailModal) {
            hideModal();
        }
    });

    // Tab switching logic
    function setupTabs() {
        const tabButtons = document.querySelectorAll('#detailTabs button');
        const tabPanes = document.querySelectorAll('#detailTabContent .tab-pane');
        
        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                // Remove active classes
                tabButtons.forEach(btn => {
                    btn.classList.remove('bg-primary-500', 'text-white');
                    btn.classList.add('bg-gray-700', 'text-gray-300');
                });
                tabPanes.forEach(pane => {
                    pane.classList.remove('block');
                    pane.classList.add('hidden');
                });
                
                // Add active classes to clicked tab
                button.classList.remove('bg-gray-700', 'text-gray-300');
                button.classList.add('bg-primary-500', 'text-white');
                const targetPane = document.querySelector(button.getAttribute('data-tab-target'));
                targetPane.classList.remove('hidden');
                targetPane.classList.add('block', 'fade-in');
            });
        });
    }

    // Simple markdown parser for basic formatting
    function parseMarkdown(text) {
        if (!text) return '';
        
        return text
            .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" class="text-primary-500 hover:underline">$1</a>')
            .replace(/\*\*([^*]+)\*\*/g, '<strong class="font-bold">$1</strong>')
            .replace(/\*([^*]+)\*/g, '<em class="italic">$1</em>')
            .replace(/`([^`]+)`/g, '<code class="bg-gray-800 px-1 rounded text-sm">$1</code>');
    }

    function showDetails(data) {
        // Ensure data has required properties with defaults
        const safeData = {
            url: data.url || '',
            final_url: data.final_url || data.url || '',
            strategy: data.strategy || 'N/A',
            user_agent: data.user_agent || 'Not available',
            elapsed: data.elapsed || 0,
            scores: data.scores || {},
            metrics: data.metrics || {},
            opportunities: data.opportunities || []
        };

        const modalBody = document.getElementById('modalBody');
        
        modalBody.innerHTML = `
            <!-- URL Header -->
            <div class="mb-4 p-4 glass-card rounded-lg">
                <h6 class="mb-2 text-lg font-semibold flex items-center">
                    <i class="fas fa-link mr-2"></i>URL Analysis
                </h6>
                <p class="mb-1"><strong>Original URL:</strong> <code class="bg-gray-800 px-1 rounded">${safeData.url}</code></p>
                <p class="mb-0"><strong>Final URL:</strong> <code class="bg-gray-800 px-1 rounded">${safeData.final_url}</code></p>
            </div>

            <!-- Tabs Navigation -->
            <ul class="flex border-b border-gray-700 mb-4" id="detailTabs" role="tablist">
                <li>
                    <button class="px-4 py-2 bg-primary-500 text-white rounded-t-lg" id="scores-tab" data-tab-target="#scores" role="tab">
                        <i class="fas fa-chart-bar mr-2"></i>Scores
                    </button>
                </li>
                <li>
                    <button class="px-4 py-2 bg-gray-700 text-gray-300 rounded-t-lg" id="metrics-tab" data-tab-target="#metrics" role="tab">
                        <i class="fas fa-stopwatch mr-2"></i>Metrics
                    </button>
                </li>
                <li>
                    <button class="px-4 py-2 bg-gray-700 text-gray-300 rounded-t-lg" id="opportunities-tab" data-tab-target="#opportunities" role="tab">
                        <i class="fas fa-lightbulb mr-2"></i>Opportunities
                        ${safeData.opportunities.length > 0 ? `<span class="ml-1 bg-yellow-500 text-black text-xs px-2 py-1 rounded-full">${safeData.opportunities.length}</span>` : ''}
                    </button>
                </li>
                <li>
                    <button class="px-4 py-2 bg-gray-700 text-gray-300 rounded-t-lg" id="technical-tab" data-tab-target="#technical" role="tab">
                        <i class="fas fa-cog mr-2"></i>Technical
                    </button>
                </li>
            </ul>

            <!-- Tab Content -->
            <div id="detailTabContent">
                <!-- Scores Tab -->
                <div class="tab-pane block fade-in" id="scores" role="tabpanel">
                    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
                        ${Object.entries(safeData.scores).map(([key, value]) => `
                            <div class="text-center p-4 glass-card rounded-lg">
                                <div class="score-circle ${getScoreClass(value)} score-circle-animation mx-auto mb-2">
                                    ${value}
                                </div>
                                <h6 class="mb-0 font-semibold">${formatLabel(key)}</h6>
                                <small class="text-gray-400">${getScoreLabel(value)}</small>
                            </div>
                        `).join('')}
                    </div>
                    
                    <div class="mt-4">
                        <h6 class="text-lg font-semibold flex items-center">
                            <i class="fas fa-info-circle mr-2"></i>Score Interpretation
                        </h6>
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <div class="flex items-center">
                                <div class="score-circle score-excellent mr-2 w-8 h-8 text-sm flex items-center justify-center">90+</div>
                                <span>Excellent (90-100)</span>
                            </div>
                            <div class="flex items-center">
                                <div class="score-circle score-good mr-2 w-8 h-8 text-sm flex items-center justify-center">50+</div>
                                <span>Good (50-89)</span>
                            </div>
                            <div class="flex items-center">
                                <div class="score-circle score-poor mr-2 w-8 h-8 text-sm flex items-center justify-center"><50</div>
                                <span>Needs Improvement (<50)</span>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Metrics Tab -->
                <div class="tab-pane hidden" id="metrics" role="tabpanel">
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <h6 class="text-lg font-semibold flex items-center">
                                <i class="fas fa-tachometer-alt mr-2"></i>Core Web Vitals
                            </h6>
                            <div class="space-y-3">
                                ${[
                                    { key: 'first_contentful_paint', label: 'First Contentful Paint (FCP)', threshold: 1800 },
                                    { key: 'largest_contentful_paint', label: 'Largest Contentful Paint (LCP)', threshold: 2500 },
                                    { key: 'first_input_delay', label: 'First Input Delay (FID)', threshold: 100 },
                                    { key: 'cumulative_layout_shift', label: 'Cumulative Layout Shift (CLS)', threshold: 0.1 }
                                ].map(metric => {
                                    const value = safeData.metrics[metric.key];
                                    if (value === undefined || value === null) {
                                        return `
                                            <div class="flex justify-between items-center py-3 border-b border-gray-700">
                                                <div>
                                                    <span class="font-bold">${metric.label}</span>
                                                    <br><small class="text-gray-400">Threshold: ${formatMetricThreshold(metric.threshold, metric.key)}</small>
                                                </div>
                                                <div class="text-right">
                                                    <span class="bg-gray-600 text-white text-xs px-2 py-1 rounded">N/A</span>
                                                </div>
                                            </div>
                                        `;
                                    }
                                    
                                    const isGood = metric.key === 'cumulative_layout_shift' ? 
                                        value <= metric.threshold : 
                                        value <= metric.threshold;
                                    return `
                                        <div class="flex justify-between items-center py-3 border-b border-gray-700">
                                            <div>
                                                <span class="font-bold">${metric.label}</span>
                                                <br><small class="text-gray-400">Threshold: ${formatMetricThreshold(metric.threshold, metric.key)}</small>
                                            </div>
                                            <div class="text-right">
                                                <span class="${isGood ? 'bg-green-500' : 'bg-yellow-500'} text-black text-xs px-2 py-1 rounded">${formatMetric(value, metric.key)}</span>
                                            </div>
                                        </div>
                                    `;
                                }).join('')}
                            </div>
                        </div>
                        
                        <div>
                            <h6 class="text-lg font-semibold flex items-center">
                                <i class="fas fa-chart-line mr-2"></i>Performance Metrics
                            </h6>
                            <div class="space-y-3">
                                ${Object.entries(safeData.metrics)
                                    .filter(([key]) => !['first_contentful_paint', 'largest_contentful_paint', 'first_input_delay', 'cumulative_layout_shift'].includes(key))
                                    .map(([key, value]) => `
                                        <div class="flex justify-between items-center py-3 border-b border-gray-700">
                                            <span class="font-bold">${formatLabel(key)}</span>
                                            <span class="text-gray-400">${formatMetric(value, key)}</span>
                                        </div>
                                    `).join('')}
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Opportunities Tab -->
                <div class="tab-pane hidden" id="opportunities" role="tabpanel">
                    <div class="flex justify-between items-center mb-3">
                        <h6 class="text-lg font-semibold flex items-center">
                            <i class="fas fa-lightbulb mr-2"></i>Performance Optimization Opportunities
                        </h6>
                        ${safeData.opportunities.length > 0 ? `
                            <div class="text-right">
                                <small class="text-gray-400">${safeData.opportunities.length} opportunities found</small>
                                <br><span class="bg-green-500 text-black text-xs px-2 py-1 rounded">
                                    ${safeData.opportunities.reduce((total, opp) => total + (opp.potentialSavings || 0), 0)}ms potential savings
                                </span>
                            </div>
                        ` : ''}
                    </div>
                    
                    <div class="max-h-[500px] overflow-y-auto">
                        ${safeData.opportunities.length > 0 ? safeData.opportunities.map((opp, index) => `
                            <div class="mb-3 p-4 glass-card rounded-lg impact-${(opp.impact || 'low').toLowerCase()}">
                                <div class="flex justify-between items-start mb-2">
                                    <div class="flex-grow">
                                        <div class="flex items-center mb-2">
                                            <h6 class="mb-0 mr-3 font-semibold">${opp.title || 'Optimization Opportunity'}</h6>
                                            <span class="${getImpactBadge(opp.impact)} text-xs px-2 py-1 rounded">${opp.impact || 'Low'}</span>
                                            ${opp.potentialSavings ? `<span class="bg-blue-500 text-white text-xs px-2 py-1 rounded ml-2">${opp.potentialSavings}ms saved</span>` : ''}
                                        </div>
                                    </div>
                                    <div class="ml-3">
                                        <i class="fas fa-${getImpactIcon(opp.impact)} text-${getImpactColor(opp.impact)}"></i>
                                    </div>
                                </div>
                                <div class="opportunity-description text-gray-300">
                                    ${parseMarkdown(opp.description || 'No description available')}
                                </div>
                                ${opp.id ? `<div class="mt-2"><small class="text-gray-400">ID: <code class="bg-gray-800 px-1 rounded">${opp.id}</code></small></div>` : ''}
                            </div>
                        `).join('') : `
                            <div class="text-center py-5">
                                <i class="fas fa-check-circle text-green-500 mb-3 text-4xl"></i>
                                <h5 class="font-semibold">No optimization opportunities found!</h5>
                                <p class="text-gray-400">Your site is performing well. Great job!</p>
                            </div>
                        `}
                    </div>
                </div>

                <!-- Technical Tab -->
                <div class="tab-pane hidden" id="technical" role="tabpanel">
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <h6 class="text-lg font-semibold flex items-center">
                                <i class="fas fa-server mr-2"></i>Request Information
                            </h6>
                            <div class="space-y-3">
                                <div class="flex justify-between py-2 border-b border-gray-700">
                                    <span>Strategy</span>
                                    <span class="bg-blue-500 text-white text-xs px-2 py-1 rounded">${safeData.strategy}</span>
                                </div>
                                <div class="flex justify-between py-2 border-b border-gray-700">
                                    <span>Elapsed Time</span>
                                    <span class="text-gray-400">${formatMetric(safeData.elapsed)}</span>
                                </div>
                                <div class="flex justify-between py-2 border-b border-gray-700">
                                    <span>DOM Size</span>
                                    <span class="text-gray-400">${safeData.metrics.dom_size || 'N/A'} ${safeData.metrics.dom_size ? 'elements' : ''}</span>
                                </div>
                                <div class="flex justify-between py-2">
                                    <span>Resources</span>
                                    <span class="text-gray-400">${safeData.metrics.resource_count || 'N/A'} ${safeData.metrics.resource_count ? 'requests' : ''}</span>
                                </div>
                            </div>
                        </div>
                        
                        <div>
                            <h6 class="text-lg font-semibold flex items-center">
                                <i class="fas fa-globe mr-2"></i>Transfer Information
                            </h6>
                            <div class="space-y-3">
                                <div class="flex justify-between py-2 border-b border-gray-700">
                                    <span>Transfer Size</span>
                                    <span class="text-gray-400">${formatBytes(safeData.metrics.transfer_size)}</span>
                                </div>
                                <div class="flex justify-between py-2 border-b border-gray-700">
                                    <span>Total Blocking Time</span>
                                    <span class="text-gray-400">${formatMetric(safeData.metrics.total_blocking_time)}</span>
                                </div>
                            </div>
                            
                            <h6 class="mt-4 text-lg font-semibold flex items-center">
                                <i class="fas fa-user-agent mr-2"></i>User Agent
                            </h6>
                            <div class="p-3 glass-card rounded-lg">
                                <small class="text-gray-400">${safeData.user_agent}</small>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        showModal();
        setupTabs(); // Initialize tab functionality
    }

    function showDetailsFromButton(button) {
        try {
            const rawData = button.getAttribute('data-page-data');
            const resultData = JSON.parse(rawData);
            const url = button.getAttribute('data-url');
            const strategy = button.getAttribute('data-strategy');
            
            const detailsData = {
                url: url,
                final_url: resultData.final_url,
                strategy: strategy,
                user_agent: resultData.user_agent,
                elapsed: resultData.elapsed,
                scores: resultData.scores,
                metrics: resultData.metrics,
                opportunities: resultData.opportunities || []
            };
            
            showDetails(detailsData);
        } catch (error) {
            console.error('Error parsing button data:', error);
            console.log('Button data:', button.getAttribute('data-page-data'));
        }
    }

    function getScoreClass(score) {
        if (score >= 90) return 'score-excellent bg-green-500 text-white rounded-full w-12 h-12 flex items-center justify-center text-lg font-bold';
        if (score >= 50) return 'score-good bg-yellow-500 text-black rounded-full w-12 h-12 flex items-center justify-center text-lg font-bold';
        return 'score-poor bg-red-500 text-white rounded-full w-12 h-12 flex items-center justify-center text-lg font-bold';
    }

    function getScoreLabel(score) {
        if (score >= 90) return 'Excellent';
        if (score >= 50) return 'Good';
        return 'Needs Improvement';
    }

    function getImpactBadge(impact) {
        if (!impact) return 'bg-gray-600 text-white';
        switch(impact.toLowerCase()) {
            case 'high': return 'bg-red-500 text-white';
            case 'medium': return 'bg-yellow-500 text-black';
            case 'low': return 'bg-green-500 text-black';
            default: return 'bg-gray-600 text-white';
        }
    }

    function getImpactIcon(impact) {
        if (!impact) return 'info-circle';
        switch(impact.toLowerCase()) {
            case 'high': return 'exclamation-triangle';
            case 'medium': return 'exclamation-circle';
            case 'low': return 'check-circle';
            default: return 'info-circle';
        }
    }

    function getImpactColor(impact) {
        if (!impact) return 'gray-400';
        switch(impact.toLowerCase()) {
            case 'high': return 'red-500';
            case 'medium': return 'yellow-500';
            case 'low': return 'green-500';
            default: return 'gray-400';
        }
    }

    function formatLabel(key) {
        return key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
    }

    function formatMetric(value, key = '') {
        if (value === null || value === undefined) return 'N/A';
        
        if (key === 'cumulative_layout_shift') {
            return value.toFixed(3);
        }
        
        if (typeof value === 'number') {
            if (value > 1000) {
                return (value / 1000).toFixed(2) + 's';
            }
            return Math.round(value) + 'ms';
        }
        return value.toString();
    }

    function formatMetricThreshold(value, key) {
        if (key === 'cumulative_layout_shift') {
            return value.toFixed(1);
        }
        if (value > 1000) {
            return (value / 1000).toFixed(1) + 's';
        }
        return value + 'ms';
    }

    function formatBytes(bytes) {
        if (bytes === 0 || bytes === null || bytes === undefined) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
