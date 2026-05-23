## Why

The frontend container fails to start because the Nginx configuration file (`nginx.conf`) references an unknown variable `$proxy_addrs` on line 14: `proxy_set_header X-Forwarded-For $proxy_addrs;`. This prevents Nginx from booting up successfully.

## What Changes

- Change `$proxy_addrs` to `$proxy_add_x_forwarded_for` in `nginx.conf`.

## Capabilities

### New Capabilities

None

### Modified Capabilities

- `containerization`: Ensures that the frontend Nginx web server starts up and routes traffic successfully.

## Impact

- `frontend/nginx.conf`: Fixes the incorrect variable name for client IP forwarding.
