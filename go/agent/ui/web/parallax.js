// Parallax Scrolling Module for Log Viewer

// Initialize parallax effect when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    initializeParallax();
});

// Initialize parallax effects
function initializeParallax() {
    const parallaxContainer = document.querySelector('.parallax-container');

    if (!parallaxContainer) {
        return;
    }

    const parallaxLayers = parallaxContainer.querySelectorAll('.parallax-layer');

    if (parallaxLayers.length === 0) {
        return;
    }

    // Mouse move parallax effect
    parallaxContainer.addEventListener('mousemove', function(e) {
        const rect = parallaxContainer.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        const centerX = rect.width / 2;
        const centerY = rect.height / 2;

        parallaxLayers.forEach(layer => {
            const speed = parseFloat(layer.getAttribute('data-speed')) || 1;
            const moveX = (x - centerX) / centerX * 20 * speed;
            const moveY = (y - centerY) / centerY * 20 * speed;

            layer.style.transform = `translate(${moveX}px, ${moveY}px)`;
        });
    });

    // Reset parallax when mouse leaves
    parallaxContainer.addEventListener('mouseleave', function() {
        parallaxLayers.forEach(layer => {
            layer.style.transform = 'translate(0, 0)';
        });
    });

    // Scroll parallax effect
    const sectionContainer = parallaxContainer.closest('.section-container');

    if (sectionContainer) {
        sectionContainer.addEventListener('scroll', function() {
            const scrollTop = sectionContainer.scrollTop;

            parallaxLayers.forEach(layer => {
                const speed = parseFloat(layer.getAttribute('data-speed')) || 1;
                const yPos = -(scrollTop * speed * 0.3);

                layer.style.transform = `translateY(${yPos}px)`;
            });
        });
    }
}

// Re-initialize parallax when content changes
if (typeof window !== 'undefined') {
    const observer = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutation) {
            if (mutation.addedNodes.length > 0) {
                mutation.addedNodes.forEach(function(node) {
                    if (node.nodeType === 1 && (node.classList.contains('parallax-container') || node.querySelector('.parallax-container'))) {
                        setTimeout(initializeParallax, 100);
                    }
                });
            }
        });
    });

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function() {
            const mainContent = document.querySelector('.main-content');
            if (mainContent) {
                observer.observe(mainContent, { childList: true, subtree: true });
            }
        });
    } else {
        const mainContent = document.querySelector('.main-content');
        if (mainContent) {
            observer.observe(mainContent, { childList: true, subtree: true });
        }
    }
}
