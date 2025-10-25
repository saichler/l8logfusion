// Global state
let currentLogContent = [];
let currentPage = 0;
const linesPerPage = 100;
let selectedFile = null;
let totalFiles = 0;

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    initializePagination();
    loadTreeData();
    updateHeaderStats();
});

// Load tree data - using example data directly
function loadTreeData() {
    const exampleData = {
        "isDirectory": true,
        "files": [
            {
                "isDirectory": true,
                "name": "192.168.86.220",
                "path": "/data/logdb",
                "files": [
                    {
                        "isDirectory": true,
                        "name": "logs",
                        "path": "/data/logdb/192.168.86.220",
                        "files": [
                            {
                                "name": "log.log",
                                "path": "/data/logdb/192.168.86.220/logs"
                            }
                        ]
                    }
                ]
            },
            {
                "isDirectory": true,
                "name": "logs",
                "path": "/data/logdb",
                "files": [
                    {
                        "name": "log.log",
                        "path": "/data/logdb/logs"
                    }
                ]
            }
        ]
    };
    renderTree(exampleData);
}

// Render the tree view
function renderTree(data) {
    const treeView = document.getElementById('tree-view');
    treeView.innerHTML = '';

    totalFiles = countFiles(data);
    updateHeaderStats();

    if (data.files && data.files.length > 0) {
        data.files.forEach(item => {
            const itemElement = createTreeItem(item);
            treeView.appendChild(itemElement);
        });
    } else {
        treeView.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">No files available</div>';
    }
}

// Count total files recursively
function countFiles(node) {
    let count = 0;
    if (node.files && node.files.length > 0) {
        node.files.forEach(item => {
            if (item.isDirectory) {
                count += countFiles(item);
            } else {
                count++;
            }
        });
    }
    return count;
}

// Update header statistics
function updateHeaderStats() {
    // File count display removed from header
}

// Create a tree item element
function createTreeItem(item, level = 0) {
    const itemDiv = document.createElement('div');
    itemDiv.className = 'tree-item';

    const nodeDiv = document.createElement('div');
    nodeDiv.className = item.isDirectory ? 'tree-node directory' : 'tree-node file';

    const icon = document.createElement('span');
    icon.className = item.isDirectory ? 'tree-icon folder-icon' : 'tree-icon file-icon';

    const name = document.createElement('span');
    name.className = 'tree-name';
    name.textContent = item.name;

    nodeDiv.appendChild(icon);
    nodeDiv.appendChild(name);
    itemDiv.appendChild(nodeDiv);

    if (item.isDirectory) {
        // Add click handler for directories to toggle expand/collapse
        nodeDiv.addEventListener('click', (e) => {
            e.stopPropagation();
            icon.classList.toggle('expanded');
            const children = nodeDiv.nextElementSibling;
            if (children && children.classList.contains('tree-children')) {
                children.classList.toggle('collapsed');
            }
        });

        // Create children container
        if (item.files && item.files.length > 0) {
            const childrenDiv = document.createElement('div');
            childrenDiv.className = 'tree-children';

            item.files.forEach(child => {
                const childItem = createTreeItem(child, level + 1);
                childrenDiv.appendChild(childItem);
            });

            itemDiv.appendChild(childrenDiv);
        }
    } else {
        // Add click handler for files to load log content
        nodeDiv.addEventListener('click', (e) => {
            e.stopPropagation();
            selectFile(nodeDiv, item);
        });
    }

    return itemDiv;
}

// Select a file and load its content
function selectFile(nodeElement, fileItem) {
    // Remove selection from all nodes
    document.querySelectorAll('.tree-node').forEach(node => {
        node.classList.remove('selected');
    });

    // Add selection to clicked node
    nodeElement.classList.add('selected');

    selectedFile = fileItem;

    // Update file path display
    const fullPath = fileItem.path + '/' + fileItem.name;
    document.getElementById('file-path').textContent = fullPath;

    // Load log content
    loadLogContent(fullPath);
}

// Load log content - using sample data directly
function loadLogContent(filePath) {
    const sampleLogs = [];
    for (let i = 1; i <= 250; i++) {
        sampleLogs.push(`[2025-10-25 ${String(Math.floor(i / 60)).padStart(2, '0')}:${String(i % 60).padStart(2, '0')}:00] INFO - Log entry number ${i} - Sample log message for testing purposes`);
    }
    currentLogContent = sampleLogs;
    currentPage = 0;
    displayCurrentPage();
}

// Display the current page of logs
function displayCurrentPage() {
    const start = currentPage * linesPerPage;
    const end = Math.min(start + linesPerPage, currentLogContent.length);

    const logDisplay = document.getElementById('log-display');
    const pageInfo = document.getElementById('page-info');
    const firstButton = document.getElementById('first-page');
    const prevButton = document.getElementById('prev-page');
    const nextButton = document.getElementById('next-page');
    const lastButton = document.getElementById('last-page');

    if (currentLogContent.length === 0) {
        logDisplay.textContent = 'No log content available';
        pageInfo.textContent = 'Showing lines 0-0';
        firstButton.disabled = true;
        prevButton.disabled = true;
        nextButton.disabled = true;
        lastButton.disabled = true;
        return;
    }

    // Display logs with line numbers
    const logsToDisplay = currentLogContent.slice(start, end);
    const numberedLogs = logsToDisplay.map((line, idx) => {
        const lineNumber = start + idx + 1;
        return `${String(lineNumber).padStart(6, ' ')} | ${line}`;
    }).join('\n');

    logDisplay.textContent = numberedLogs;

    // Update pagination info
    const totalPages = Math.ceil(currentLogContent.length / linesPerPage);
    pageInfo.textContent = `Showing lines ${start + 1}-${end} of ${currentLogContent.length} (Page ${currentPage + 1}/${totalPages})`;

    // Update button states
    const isFirstPage = currentPage === 0;
    const isLastPage = end >= currentLogContent.length;

    firstButton.disabled = isFirstPage;
    prevButton.disabled = isFirstPage;
    nextButton.disabled = isLastPage;
    lastButton.disabled = isLastPage;
}

// Initialize pagination controls
function initializePagination() {
    const firstButton = document.getElementById('first-page');
    const prevButton = document.getElementById('prev-page');
    const nextButton = document.getElementById('next-page');
    const lastButton = document.getElementById('last-page');

    firstButton.addEventListener('click', () => {
        if (currentPage > 0) {
            currentPage = 0;
            displayCurrentPage();
        }
    });

    prevButton.addEventListener('click', () => {
        if (currentPage > 0) {
            currentPage--;
            displayCurrentPage();
        }
    });

    nextButton.addEventListener('click', () => {
        const maxPage = Math.ceil(currentLogContent.length / linesPerPage) - 1;
        if (currentPage < maxPage) {
            currentPage++;
            displayCurrentPage();
        }
    });

    lastButton.addEventListener('click', () => {
        const maxPage = Math.ceil(currentLogContent.length / linesPerPage) - 1;
        if (currentPage < maxPage) {
            currentPage = maxPage;
            displayCurrentPage();
        }
    });
}
