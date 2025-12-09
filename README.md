# Shopify Webhook Faker

Tiny Go web app to craft and deliver signed Shopify webhooks to any HTTP endpoint you control. Paste your shared secret, choose a topic, drop in JSON, and the app sends a POST with Shopify-style headers (including `X-Shopify-Hmac-Sha256`) so you can exercise webhook handling locally or in staging.

## Features
- Web UI to enter shared secret, target URL, topic, shop domain, and JSON body
- HMAC-SHA256 signing using the shared secret with a base64 encoded signature that matches Shopify’s docs
- Headers set for `X-Shopify-Hmac-Sha256`, `X-Shopify-Topic`, `X-Shopify-Shop-Domain`, `Content-Type`, and a simple user agent
- HTMX-powered inline feedback showing whether the downstream endpoint succeeded and what it returned

## Prerequisites
- Go 1.25+ installed locally

## Run locally
```bash
go run .
```
Then open http://localhost:8080 in your browser.

## Usage
1) Enter your Shopify app’s shared secret (it is only used to sign this single request).  
2) Set the target URL you want to receive the fake webhook (e.g., a local tunnel or staging endpoint).  
3) Optional: set a topic (default `orders/create`) and shop domain (default `example.myshopify.com`).  
4) Paste a JSON payload.  
5) Click **Send Signed Webhook**. The app posts your JSON with the computed `X-Shopify-Hmac-Sha256` header. The inline alert will show the downstream HTTP status and body (up to ~2 MB).

## Notes
- The HTTP client uses a 10s timeout.
- Non-JSON payloads or invalid URLs are rejected before sending.
- Headers and signing logic mirror Shopify’s webhook verification guide; use the same verification routine in your app to confirm the signature.
