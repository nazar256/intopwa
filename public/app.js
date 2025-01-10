// app.js
document.getElementById('createAppButton').addEventListener('click', function() {
    var urlInput = document.getElementById('urlInput').value.trim();
    var iconInputs = Array.from(document.getElementsByClassName('iconInput'))
        .map(input => input.value.trim())
        .filter(value => value !== ''); // Remove empty values

    if (urlInput) {
        try {
            // Ensure the URL has a protocol, adding 'https://' if missing for parsing purposes
            if (!urlInput.startsWith('http://') && !urlInput.startsWith('https://')) {
                urlInput = 'https://' + urlInput;
            }

            // Parse the URL
            var url = new URL(urlInput);

            // Extract the domain and path
            var domain = url.hostname; // domain (e.g., google.com)
            var path = url.pathname + url.search + url.hash; // path including query and hash

            // Construct the redirection URL
            var redirectUrl = 'https://intopwa.xyofn8h7t.workers.dev/a/' + encodeURIComponent(domain) + path;

            // Create a form and submit it programmatically
            const form = document.createElement('form');
            form.method = 'POST';
            form.action = redirectUrl;

            // Add all icon URLs as hidden fields
            iconInputs.forEach((iconUrl, _) => {
                const iconField = document.createElement('input');
                iconField.type = 'hidden';
                iconField.name = 'icons[]'; // Using array notation for backend processing
                iconField.value = iconUrl;
                form.appendChild(iconField);
            });

            document.body.appendChild(form);
            form.submit();
            document.body.removeChild(form);
        } catch (error) {
            alert('Please enter a valid URL');
            console.error('Invalid URL:', error);
        }
    } else {
        alert('Please enter a URL');
    }
});

// Add icon button handler
document.getElementById('addIconButton').addEventListener('click', function() {
        const iconFields = document.getElementById('iconFields');
        const newField = document.createElement('div');
        newField.className = 'icon-field';
        newField.innerHTML = `
        <input type="text" class="iconInput" placeholder="Enter icon URL" />
        <button class="removeIcon">×</button>
    `;
        iconFields.appendChild(newField);

        newField.querySelector('.removeIcon').addEventListener('click', function () {
            iconFields.removeChild(newField);
        });
    }
);

if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js')
            .then(_ => {
                console.log('ServiceWorker registration successful');
            })
            .catch(err => {
                console.log('ServiceWorker registration failed: ', err);
            });
    });
}
