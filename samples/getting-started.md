# Quick Start Guide

## Introduction
This guide covers the complete workflow from installation to advanced usage, compatible with Node.js 16+ and modern browsers. Provides examples for both TypeScript and JavaScript environments.

+++ New Features
- Multi-environment installation guide
- Configuration examples
- Debugging techniques
+++

## Installation

```bash
# npm
npm install my-package@latest --save

# yarn
yarn add my-package@latest

# pnpm
pnpm add my-package@latest

# CDN (Browser environments)
<script src="https://unpkg.com/my-package@latest/dist/browser.min.js"></script>
```

## Basic Usage

### Initial Configuration
```javascript
// Basic configuration
const config = {
  env: process.env.NODE_ENV || 'development',
  port: process.env.PORT || 3000,
  cache: {
    maxAge: 3600,
    staleWhileRevalidate: 600
  }
};

// TypeScript type hints
interface AppConfig {
  env: 'development' | 'production';
  port: number;
  cache?: {
    maxAge: number;
    staleWhileRevalidate?: number;
  };
}
```

### Core API Usage
```javascript
import { initializeApp, createLogger } from 'my-package';

// Initialize application
const app = initializeApp({
  debug: process.env.NODE_ENV !== 'production'
});

// Async operation with error handling
async function fetchData() {
  try {
    const data = await app.fetch('/api/data', {
      timeout: 5000
    });
    logger.info('Data fetched:', data);
  } catch (error) {
    logger.error('Fetch failed:', error);
    throw new Error('DATA_FETCH_FAILED');
  }
}
```

## Configuration Options

| Option | Type | Default | Env Variable | Description |
|--------|------|--------|--------------|-------------|
| port | number | 3000 | APP_PORT | Service listening port |
| host | string | 'localhost' | APP_HOST | Binding host address |
| logLevel | string | 'info' | LOG_LEVEL | Logging level (debug/info/warn/error) |
| rateLimit | object | { window: 60, max: 100 } | - | Rate limiting configuration |

**Configuration Example:**
```yaml
# config.yml
production:
  host: 0.0.0.0
  port: 8080
  database:
    url: postgres://user:pass@localhost:5432/db
    poolSize: 10
```

## Best Practices
1. **Environment Isolation**: Use separate config files for dev/prod
2. **Error Handling**: Global unhandled promise rejection catcher
```javascript
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});
```
3. **Performance Optimization**: Enable gzip compression and HTTP/2
4. **Security Recommendations**: Set CSP headers for XSS protection

## FAQ
**Q: Peer dependencies warnings during installation?**  
A: Run `npm install --legacy-peer-deps` or upgrade to npm 8+

**Q: "window is not defined" in browser environment?**  
A: Use UMD build or configure global variables via bundler

**Q: TypeScript types not found?**  
A: Ensure @types/my-package is installed or check type references in tsconfig.json