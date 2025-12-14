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
