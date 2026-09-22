const MAX_ICON_FILE_BYTES = 1024 * 1024;
const SUPPORTED_ICON_FILE_TYPES = new Set([
  "image/png",
  "image/jpeg",
  "image/gif",
  "image/webp",
  "image/svg+xml",
  "image/x-icon",
  "image/vnd.microsoft.icon",
]);

const createAppForm = document.getElementById("createAppForm");
const formError = document.getElementById("formError");
const iconFileInput = document.getElementById("iconFileInput");

// app.js
createAppForm.addEventListener("submit", function (event) {
  formError.textContent = "";

  var urlInput = document.getElementById("urlInput").value.trim();
  var iconInputs = Array.from(document.getElementsByClassName("iconInput"))
    .map((input) => input.value.trim())
    .filter((value) => value !== ""); // Remove empty values
  const iconFile = iconFileInput.files[0];

  if (iconFile && iconInputs.length > 0) {
    event.preventDefault();
    formError.textContent =
      "Choose either icon URLs or one uploaded icon file, not both.";
    return;
  }

  if (iconFile) {
    const validationError = validateIconFile(iconFile);
    if (validationError) {
      event.preventDefault();
      formError.textContent = validationError;
      return;
    }
  }

  if (urlInput) {
    try {
      // Ensure the URL has a protocol, adding 'https://' if missing for parsing purposes
      if (!urlInput.startsWith("http://") && !urlInput.startsWith("https://")) {
        urlInput = "https://" + urlInput;
      }

      // Parse the URL
      var url = new URL(urlInput);

      // Extract the domain and path
      var domain = url.hostname; // domain (e.g., google.com)
      var path = url.pathname + url.search + url.hash; // path including query and hash

      // Construct the redirection URL
      var redirectUrl =
        "https://intopwa.xyofn8h7t.workers.dev/a/" +
        encodeURIComponent(domain) +
        path;

      createAppForm.action = redirectUrl;

      // Add all icon URLs as hidden fields
      Array.from(
        createAppForm.querySelectorAll('input[name="icons[]"]'),
      ).forEach((input) => input.remove());
      iconInputs.forEach((iconUrl, _) => {
        const iconField = document.createElement("input");
        iconField.type = "hidden";
        iconField.name = "icons[]"; // Using array notation for backend processing
        iconField.value = iconUrl;
        createAppForm.appendChild(iconField);
      });
    } catch (error) {
      event.preventDefault();
      formError.textContent = "Please enter a valid URL";
      console.error("Invalid URL:", error);
    }
  } else {
    event.preventDefault();
    formError.textContent = "Please enter a URL";
  }
});

function validateIconFile(file) {
  if (!SUPPORTED_ICON_FILE_TYPES.has(file.type)) {
    return "Unsupported icon type. Upload a PNG, JPEG, GIF, WebP, SVG, or ICO file.";
  }
  if (file.size > MAX_ICON_FILE_BYTES) {
    return "Icon file is too large. Upload an image up to 1 MiB.";
  }
  return "";
}

// Add icon button handler
document.getElementById("addIconButton").addEventListener("click", function () {
  const iconFields = document.getElementById("iconFields");
  const newField = document.createElement("div");
  newField.className = "icon-field";
  newField.innerHTML = `
        <input type="text" class="iconInput" placeholder="Enter icon URL" />
        <button class="removeIcon" type="button">×</button>
    `;
  iconFields.appendChild(newField);

  newField.querySelector(".removeIcon").addEventListener("click", function () {
    iconFields.removeChild(newField);
  });
});

if ("serviceWorker" in navigator) {
  window.addEventListener("load", () => {
    navigator.serviceWorker
      .register("/sw.js")
      .then((_) => {
        console.log("ServiceWorker registration successful");
      })
      .catch((err) => {
        console.log("ServiceWorker registration failed: ", err);
      });
  });
}
