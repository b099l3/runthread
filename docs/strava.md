# Strava Integration

Strava is the first real MVP activity provider for Runthread.

The purpose of the Strava integration is to validate the activity import loop quickly:

1. User connects Strava using OAuth.
2. Backend exchanges the authorization code and stores provider connection tokens securely.
3. Backend refreshes tokens as needed.
4. Backend receives Strava webhook events.
5. Backend fetches activity details from Strava when needed.
6. Backend normalises Strava activity data into Runthread's `ImportedActivity` model.
7. Backend matches imported activities to planned workouts.
8. Workout results drive deterministic adaptation.

The mock provider path is documented and tested. Strava OAuth start/callback, token storage, token refresh, webhook handling, backfill, and HTTP activity fetch boundaries now exist behind backend-owned provider packages. Remaining work should focus on live provider validation, production hardening, operational observability, and privacy/support behavior.

## Backend Responsibilities

The backend owns the full Strava integration lifecycle:

- Start the OAuth flow and generate state.
- Exchange authorization codes for provider tokens.
- Store token references securely, outside core domain models.
- Refresh access tokens before activity import when required.
- Receive, verify, and deduplicate webhook events.
- Fetch activity details from Strava under rate limits.
- Normalise activity data into `domain.ImportedActivity`.
- Record provider import state for audit, idempotency, retry, and support.

The Flutter app may start the connect flow and open a backend-provided authorization URL. It should not exchange OAuth codes, store tokens, refresh tokens, or call the Strava API directly.

## Data Boundary

Strava payloads, Strava activity IDs, scopes, webhook event shapes, and rate-limit behavior belong in Strava provider packages and provider persistence boundaries.

Core domain logic should only receive `ImportedActivity`, `WorkoutMatch`, `WorkoutResult`, and `AdaptationEvent` values. Strava-specific fields should not leak into planning, matching, adaptation, or mobile plan-read models unless they are explicitly mapped to provider-neutral fields.

## Rate Limits

The backend must respect Strava API rate limits. Import jobs should be idempotent and retryable, with backoff for temporary failures and clear status when rate limits delay sync.

Backfill, webhook-triggered imports, manual retries, and future polling should share the same rate-limit-aware import path.

## Privacy and AI Boundary

Strava-derived activity data must not be sent to AI prompts and must not be used for AI model training.

Strava activity data should only be shown to the authorised Runthread user who connected that Strava account. Future support, debugging, export, deletion, and retention behavior must respect Strava terms and Runthread privacy policy decisions.

## MVP Scope

Production-facing Strava work is staged as:

1. Mock Strava provider payloads and normalisation. Done.
2. OAuth start/callback and secure token reference storage. Implemented; needs production hardening.
3. Backfill/import jobs for recent activities. Implemented behind provider-neutral import services; needs live validation.
4. Webhook subscription and event handling. Implemented behind backend routes/services; needs live validation and operations coverage.
5. Matching imported activities to planned workouts. Implemented through provider-neutral matching paths.
6. Adaptation from matched workout results. Implemented through deterministic adaptation paths.

Before beta, validate the live Strava app configuration, rate-limit behavior, webhook delivery, token key management, revoked-access transitions, and support/debugging surfaces.
