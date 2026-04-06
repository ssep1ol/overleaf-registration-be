# overleaf-registration-be

Backend service for controlled Overleaf account registration.

This service exposes a single endpoint used by a frontend signup form. It validates:

- email domain allowlist
- Cloudflare Turnstile token

and then performs admin-authenticated user creation on an Overleaf instance.

## Features

- HTTP API for signup flow
- Domain-based admission control
- Turnstile CAPTCHA verification
- Admin session login against Overleaf
- Overleaf user registration through admin endpoint
- Configurable runtime via environment variables
- Docker image support

## How It Works

1. Client sends email and CAPTCHA token to `/signup`.
2. Service verifies email domain against `ALLOWED_DOMAINS`.
3. Service verifies token via Cloudflare Turnstile verify API.
4. Service logs in to Overleaf as admin.
5. Service posts user email to Overleaf admin registration endpoint.
6. Service returns success JSON message.

## Tech Stack

- Go (standard library net/http)
- Cloudflare Turnstile verification API
- goquery for CSRF token extraction
- rs/cors for CORS middleware
- godotenv for local env loading

## Project Structure

- `main.go`: server bootstrap, env validation, route setup
- `handlers.go`: request handling and signup flow orchestration
- `overleaf.go`: Overleaf client logic (CSRF, login, registration)
- `recaptcha.go`: Turnstile token verification
- `.env.example`: environment variable template
- `Dockerfile`: container image build instructions

## API

### POST /signup

Registers a user on Overleaf after domain and CAPTCHA checks.

Request body:

```json
{
	"email": "student@example.edu",
	"captcha": "turnstile_token_here"
}
```

Success response (`200`):

```json
{
	"message": "User registered successfully. Please check your email to set your password."
}
```

Typical error responses:

- `400`: invalid JSON or disallowed email domain
- `401`: invalid CAPTCHA token
- `500`: missing config, Turnstile failure, Overleaf login/registration failure

### Example curl

```bash
curl -X POST http://localhost:3000/signup \
	-H "Content-Type: application/json" \
	-d '{
		"email": "student@example.edu",
		"captcha": "TOKEN_FROM_TURNSTILE"
	}'
```

## Configuration

Set environment variables via system env or `.env` file.

Required:

- `ALLOWED_DOMAINS`: comma-separated domain allowlist
	- example: `example.edu,department.example.edu`
- `TURNSTILE_SECRET_KEY`: Cloudflare Turnstile secret key
- `OL_INSTANCE`: Overleaf base URL
	- example: `https://overleaf.yourdomain.tld`
- `OL_ADMIN_EMAIL`: Overleaf admin email
- `OL_ADMIN_PASSWORD`: Overleaf admin password

Optional:

- `PORT`: server port (default: `3000`)

Important note:

- The code uses `TURNSTILE_SECRET_KEY`.
- `.env.example` currently includes `CAPTCHA_SERVER_KEY`; this should be aligned before production usage.

## Run Locally

### 1) Create env file

```bash
cp .env.example .env
```

Edit `.env` and set required variables.

### 2) Install dependencies and run

```bash
go mod init main
go mod tidy
go run .
```

Service starts on `http://localhost:3000` unless `PORT` is set.

## Run With Docker

Build image:

```bash
docker build -t overleaf-registration-be .
```

Run container:

```bash
docker run --rm -p 3000:3000 --env-file .env overleaf-registration-be
```

## Operational Notes

- Startup fails fast when required env vars are missing.
- Overleaf integration relies on HTML CSRF token extraction from:
	- `/login`
	- `/admin/register`
- If Overleaf page markup changes, CSRF extraction may need update.

## CI

GitLab CI currently includes SAST template in `.gitlab-ci.yml`.

## Troubleshooting

- `environment variable ... is not set`
	- Ensure required variables are present in `.env` or runtime environment.

- `Invalid CAPTCHA`
	- Check Turnstile site/secret key pairing and token freshness.

- `Admin login failed`
	- Verify `OL_INSTANCE`, admin email/password, and Overleaf reachability.

- `CSRF token not found`
	- Overleaf login/admin page markup may have changed.

## License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.
