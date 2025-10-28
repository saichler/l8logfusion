// Global state
let currentLogContent = '';
let currentPage = 0;
const bytesPerPage = 5120; // 5KB per page
let selectedFile = null;
let totalFiles = 0;
let totalBytes = 0;

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    initializePagination();
    loadTreeData();
    updateHeaderStats();
});

// Load tree data from REST API
async function loadTreeData() {
    try {
        const response = await fetch('/probler/87/logs');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        renderTree(data);
    } catch (error) {
        console.error('Error loading tree data:', error);
        const treeView = document.getElementById('tree-view');
        treeView.innerHTML = '<div style="color: var(--text-muted); padding: 10px;">Error loading log files. Please try again later.</div>';
    }
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

// Load log content from file
async function loadLogContent(filePath) {
    try {
        // Reset to first page
        currentPage = 0;
        // Load the first page
        await loadLogPage(filePath, currentPage);
    } catch (error) {
        console.error('Error loading log content:', error);
        currentLogContent = 'Error loading log content. Please try again later.';
        totalBytes = 0;
        displayCurrentPage();
    }
}

// Load a specific page of log content
async function loadLogPage(filePath, page) {
    try {
        // Parse path and name from filePath
        const lastSlashIndex = filePath.lastIndexOf('/');
        const path = filePath.substring(0, lastSlashIndex);
        const name = filePath.substring(lastSlashIndex + 1);

        // Build the SQL-like query text
        const queryText = `select * from l8file where path=${path} and name = ${name} limit 1 page ${page}`;

        // Create the body parameter as JSON
        const bodyParam = JSON.stringify({ text: queryText });

        // Build query URL with body parameter
        const url = `/probler/87/logs?body=${encodeURIComponent(bodyParam)}`;

        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        // Update state with the returned data
        if (data.data) {
            currentLogContent = data.data.content || '';
            totalBytes = data.data.size || 0;
            currentPage = page;
        } else {
            currentLogContent = '';
            totalBytes = 0;
        }

        displayCurrentPage();
    } catch (error) {
        console.error('Error loading log page:', error);
        currentLogContent = 'Error loading log content. Please try again later.';
        totalBytes = 0;
        displayCurrentPage();
    }
}

// Display the current page of logs
function displayCurrentPage() {
    const logDisplay = document.getElementById('log-display');
    const pageInfo = document.getElementById('page-info');
    const firstButton = document.getElementById('first-page');
    const prevButton = document.getElementById('prev-page');
    const nextButton = document.getElementById('next-page');
    const lastButton = document.getElementById('last-page');

    if (totalBytes === 0) {
        logDisplay.textContent = 'No log content available';
        pageInfo.textContent = 'Showing bytes 0-0 of 0 (0 KB)';
        firstButton.disabled = true;
        prevButton.disabled = true;
        nextButton.disabled = true;
        lastButton.disabled = true;
        return;
    }

    // Display the content directly
    logDisplay.textContent = currentLogContent;

    // Calculate byte ranges
    const startByte = currentPage * bytesPerPage + 1;
    const endByte = Math.min(startByte + currentLogContent.length - 1, totalBytes);
    const totalPages = Math.ceil(totalBytes / bytesPerPage);

    // Format sizes in KB for display
    const startKB = (startByte / 1024).toFixed(2);
    const endKB = (endByte / 1024).toFixed(2);
    const totalKB = (totalBytes / 1024).toFixed(2);

    // Update pagination info with byte ranges
    pageInfo.textContent = `Showing bytes ${startByte}-${endByte} of ${totalBytes} (${startKB}-${endKB} KB of ${totalKB} KB) | Page ${currentPage + 1}/${totalPages}`;

    // Update button states
    const isFirstPage = currentPage === 0;
    const isLastPage = (currentPage + 1) >= totalPages;

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

    firstButton.addEventListener('click', async () => {
        if (currentPage > 0 && selectedFile) {
            const fullPath = selectedFile.path + '/' + selectedFile.name;
            await loadLogPage(fullPath, 0);
        }
    });

    prevButton.addEventListener('click', async () => {
        if (currentPage > 0 && selectedFile) {
            const fullPath = selectedFile.path + '/' + selectedFile.name;
            await loadLogPage(fullPath, currentPage - 1);
        }
    });

    nextButton.addEventListener('click', async () => {
        const maxPage = Math.ceil(totalBytes / bytesPerPage) - 1;
        if (currentPage < maxPage && selectedFile) {
            const fullPath = selectedFile.path + '/' + selectedFile.name;
            await loadLogPage(fullPath, currentPage + 1);
        }
    });

    lastButton.addEventListener('click', async () => {
        const maxPage = Math.ceil(totalBytes / bytesPerPage) - 1;
        if (currentPage < maxPage && selectedFile) {
            const fullPath = selectedFile.path + '/' + selectedFile.name;
            await loadLogPage(fullPath, maxPage);
        }
    });
}
