# API reference

The API listens on port 8080 by default. Routes under `/api/v1` return JSON.

## Health & Authentication
- `GET /health` — Service health check
- `POST /api/v1/auth/signup` — Create account and queue verification email via outbox
- `POST /api/v1/auth/verify-email` — Verify account email token
- `POST /api/v1/auth/resend-verification` — Resend verification email
- `POST /api/v1/auth/login` — Authenticate and receive access + refresh tokens
- `POST /api/v1/auth/refresh` — Rotate refresh token and issue new access token
- `POST /api/v1/auth/logout` — Revoke current session
- `POST /api/v1/auth/forgot-password` — Request password reset link
- `POST /api/v1/auth/reset-password` — Reset password using token
- `GET /api/v1/auth/me` — Get current authenticated user profile

## Monitors, Incidents & Alert Channels (Bearer Auth)
- `GET /api/v1/dashboard/summary` — Aggregate uptime stats and recent incidents
- `POST /api/v1/monitors` — Create an HTTP/HTTPS monitor
- `GET /api/v1/monitors` — List monitors for the current user
- `GET /api/v1/monitors/{monitorID}` — Get monitor configuration and current status
- `PATCH /api/v1/monitors/{monitorID}` — Update monitor interval, timeout, or thresholds
- `DELETE /api/v1/monitors/{monitorID}` — Delete a monitor
- `POST /api/v1/monitors/{monitorID}/enable` — Resume scheduled checks
- `POST /api/v1/monitors/{monitorID}/disable` — Pause scheduled checks
- `GET /api/v1/monitors/{monitorID}/results` — List recent probe check results and latency breakdowns
- `GET /api/v1/monitors/{monitorID}/incidents` — List incident history for a monitor
- `GET /api/v1/alert-channels` — List configured email and webhook alert channels
- `POST /api/v1/alert-channels` — Add an email or HMAC-signed webhook alert channel
- `DELETE /api/v1/alert-channels/{channelID}` — Remove an alert channel
