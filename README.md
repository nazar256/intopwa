# IntoPWA

Turn any website into an installable Progressive Web App (PWA) with just one click.

> **Note**: This project is a prototype/proof-of-concept implementation focused on demonstrating functionality rather than production-ready code quality.

## Features

- Create PWAs from any website URL
- Custom icon support
- Automatic manifest generation

## Live Demo

Visit [into-progressive.web.app](https://into-progressive.web.app) to try it out.
Production backend API is available at https://intopwa.xyofn8h7t.workers.dev/.

## Development

The project consists of two parts:
1. Frontend (Firebase hosted)
2. Cloudflare Worker backend (Go)

## Create-PWA icon API

The generated app endpoint is `POST /a/{target-host-and-path}`. Icon customization supports exactly one explicit icon source mode per request:

1. **Hosted icon URLs**: submit one or more `icons[]` form fields with absolute `http(s)` URLs, or host/path values that the worker normalizes to `https://...`.
2. **Uploaded icon file**: submit `multipart/form-data` with one `iconFile` field and no `icons[]` fields.

When an icon file is uploaded, the worker validates the content server-side, stores the bytes in the existing icon KV cache, stores a stable content-hash icon reference for the generated app, and emits that reference in the generated manifest as `/i/uploaded-icons.intopwa.local/{sha256}.{ext}`. Browser/client-side validation is only UX help; the worker enforces the authoritative limits.

Uploaded icon URLs are content-addressed and stable: identical uploads map to the same URL, and the bytes persist indefinitely in the worker's KV icon cache (no expiry). The synthetic host is never fetched over the network — a missing cache entry resolves to an error rather than an outbound fetch.

Uploaded icon constraints:

* Max file size: 1 MiB.
* Accepted formats: PNG, JPEG, GIF, WebP, SVG, ICO.
* The worker sniffs and decodes image content instead of trusting filenames or client-provided MIME types.
* Requests that include both `icons[]` and `iconFile` are rejected as ambiguous.

If no explicit icon source is submitted, the worker preserves existing behavior: it scrapes icons from the target website and falls back to the default app icon when none are found.

### Prerequisites

- Node.js
- Go 1.21+
- TinyGo
- Firebase CLI
- Wrangler CLI

## Setup

For deployment you need to setup firebase client and wrangler cli.

### Frontend
Prepare firebase client:
```
npx firebase login
npx firebase init
```

### Backend
Prepare wrangler cli:
```
cd worker
npx wrangler login
```

### Deployment
Both front-end and back-end with:

```bash
cd worker
make deploy
```


## Project Structure
* /public - Frontend static files (firebase hosting)
* /worker - Cloudflare Worker backend written in Go
* /internal - Worker implementation
* /build - Compiled worker files

## Tech Stack
* Frontend: HTML, CSS, JavaScript
* Backend: Go (TinyGo for WASM)
* Infrastructure:
  * Firebase Hosting
  * Cloudflare Workers
  * Cloudflare KV Storage

## License
  MIT License
