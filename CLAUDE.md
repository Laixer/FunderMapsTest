# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FunderMaps is a Go REST API built with Fiber v2 web framework and PostgreSQL (via GORM). It provides user authentication (session + JWT + OAuth2), incident reporting, geocoding, and integrations with external services (Mailgun, S3, PDF processing).

## Build & Development Commands

```bash
make build          # Build binary (fundermapsapp)
make run            # Build and run
make test           # Run all tests (go test -v ./...)
make fmt            # Format code (go fmt ./...)
make lint           # Lint with golangci-lint
make clean          # Clean build artifacts
go test -v -run TestName ./path/to/package  # Run a single test
```

The server runs on port 3000 by default. Configuration is loaded from environment variables (prefixed `FM_`) and YAML files (`settings.yaml`, `settings.dev.yaml`) via Viper.

## Architecture

**Entry point**: `cmd/server/main.go` — sets up the Fiber app, middleware stack, and all route definitions.

**Core structure**:
- `app/config/` — Viper-based configuration with validation
- `app/database/` — GORM connection setup and all database models (uses multiple schemas: application, geocoder, data)
- `app/handlers/` — HTTP endpoint handlers organized by domain, with `management/` sub-package for admin endpoints
- `app/middleware/` — Auth (JWT/API key), admin authorization, request tracking, robots
- `app/platform/` — Business logic services (crypto, geocoder, incident, job, storage, user)
- `pkg/utils/` — Password hashing (Argon2id + legacy PBKDF2) and helper utilities
- `public/` — Static HTML (login page)
- `locales/` — i18n translations (Dutch)

**Route groups**: `/auth/*` (session), `/api/auth/*` (JWT), `/api/v1/oauth2/*` (OAuth2), `/api/user/*`, `/api/incident|mapset|product|report`, `/api/v1/management/*` (admin)

## Code Patterns

- Handlers use the Fiber `*fiber.Ctx` receiver pattern, returning `error`
- Services are structs with a DB dependency injected
- Input validation via `go-playground/validator` struct tags
- Database column names use snake_case; Go identifiers use camelCase
- Crypto service uses AES-256-GCM with PBKDF2 key derivation and versioned ciphertext
- Password hashing defaults to Argon2id with legacy PBKDF2 fallback for migration
