# Strava Integration Implementation Plan

This plan turns the existing Strava skeletons into the private beta MVP activity import path.

Target flow:

```text
Connect Strava
  -> backend OAuth callback
  -> secure token storage
  -> initial activity backfill
  -> webhook imports for future activity changes
  -> provider-neutral ImportedActivity
  -> workout matching
  -> workout result
  -> deterministic adaptation
  -> mobile plan refresh
```

The implementation should reuse the existing provider-neutral import, matching, completion, and adaptation boundaries. Strava-specific OAuth details, token data, API payloads, webhook bodies, activity IDs, scopes, and rate-limit behavior must stay inside Strava provider packages and provider persistence boundaries.

## Current Foundation

Already present:

- Mock Strava activity payloads and normalisation in `services/api/internal/providers/strava`.
- Backend-only Strava OAuth service shape for start/callback and token-reference storage.
- Backend-only Strava backfill service shape with rate-limit deferral.
- Backend-only Strava webhook service shape with verifier, deduper, create/update/delete routing, and provider import persistence.
- Provider persistence tables and repository boundaries for provider connections, provider activities, provider payloads, and import events.
- Provider-neutral app services for import, match, workout result creation, and deterministic adaptation.

Still missing:

- Strava as a first-class RPC/mobile provider enum.
- Production Strava configuration and server wiring.
- Real Strava code exchange, token refresh, token storage, API client, and webhook endpoint.
- Startup composition that wires Strava provider services to Postgres-backed repositories.
- Mobile Strava connection UX.
- Integration tests against Postgres-backed provider persistence and fake Strava HTTP behavior.

## Implementation Phases

### Phase 1: Provider Contract and Configuration

Goal: make Strava addressable through existing provider connection APIs without enabling live imports yet.

Backend changes:

- Add `PROVIDER_STRAVA` to `Provider` in `services/api/proto/runthread/v1/runthread.proto`.
- Regenerate Go protobuf/ConnectRPC code.
- Add Strava enum mapping in RPC handlers.
- Extend provider connection application validation so `strava` is supported alongside `garmin`.
- Keep Garmin behavior unchanged and still disabled for live connection.
- Add config fields and environment variables for:
  - Strava client ID.
  - Strava client secret.
  - Strava OAuth redirect URI.
  - Strava webhook verify token or equivalent verification secret.
  - Optional Strava API base URL for tests/staging.

Mobile changes:

- Add `Provider.strava` to mobile provider models.
- Keep Strava connect disabled until backend OAuth start returns a real authorization URL.

Tests:

- RPC mapping tests for Strava provider enum round trips.
- Provider connection service tests for Strava pending/status behavior.
- Mobile model tests for Strava provider parsing and labels.

### Phase 2: OAuth and Token Storage

Goal: support a backend-owned Strava OAuth connection flow.

Backend changes:

- Wire `StartProviderConnection` for `strava` to the Strava OAuth service instead of the generic readiness placeholder.
- Return `authorization_url`, `state`, and `oauth_ready=true` when Strava config is complete.
- Add an HTTP callback endpoint for Strava OAuth completion.
- Implement real Strava authorization-code exchange against Strava's token endpoint.
- Store token material through a production token store that returns a token reference.
- Persist provider user ID, token reference, token expiry, connected status, and callback errors on `provider_connections`.
- Add token refresh support behind a Strava token manager used by API fetch/backfill/webhook imports.
- Mark connections `error` or `disconnected` for invalid grants, revoked access, and unrecoverable refresh failures.

Security requirements:

- Mobile must never receive access tokens or refresh tokens.
- Core domain models must never contain token material.
- Logs must not print token values, authorization codes, or refresh tokens.

Tests:

- OAuth start with configured and missing Strava credentials.
- Callback success stores token reference and connects the provider connection.
- Invalid state, denied access, code exchange failure, token storage failure, and refresh failure paths.
- Token redaction/logging checks where practical.

### Phase 3: Real Strava API Client and Initial Backfill

Goal: import recent Strava activities after connection and record useful status for beta support.

Backend changes:

- Implement a real Strava activity client behind the existing `ActivityFetcher` interface.
- Fetch recent athlete activities after OAuth callback using the connected provider connection.
- Fetch full activity details before normalisation when the summary data is insufficient.
- Map Strava run-like activity types into provider-neutral `domain.ImportedActivity`.
- Record unsupported non-run activities as ignored provider imports.
- Preserve idempotency by provider connection ID plus Strava activity ID.
- Respect Strava rate-limit headers and return deferred backfill status when limits are reached.
- Store raw Strava payload snapshots only if allowed by the active privacy/retention decision; otherwise make payload storage policy-controlled.
- Update connection `last_sync_at`, `last_error`, and status after backfill attempts.

App-service integration:

- After import, route activities through provider-neutral match, workout result, and adaptation services.
- Confident matches may create workout results automatically.
- Uncertain/rejected matches should be persisted for later review and must not create workout results automatically.

Tests:

- Backfill imports a supported run and updates provider/imported activity state.
- Backfill ignores unsupported activity types.
- Re-running backfill is idempotent.
- Rate limits defer without marking the connection permanently failed.
- Partial failures preserve import events and support diagnostics.
- Imported Strava run can drive match, workout result, and deterministic adaptation.

### Phase 4: Webhook Endpoint and Ongoing Sync

Goal: keep Strava activity state current after the initial connection.

Backend changes:

- Add provider-facing Strava webhook endpoint to the HTTP server.
- Implement Strava webhook subscription validation/handshake if required by Strava's current API.
- Verify webhook requests using configured verification behavior.
- Persist webhook dedupe state so duplicate deliveries are ignored.
- Route create/update events through activity detail fetch, normalisation, provider import, matching, completion, and adaptation.
- Route delete events to provider import state and decide whether existing normalised activities/results remain, are marked stale, or require review.
- Record failed fetch/import webhook events with retry/support-visible errors.
- Add retry/backoff behavior for temporary Strava API failures and rate limits.

Tests:

- Webhook validation/handshake.
- Verified create/update imports activity detail.
- Delete event records provider state without provider payload leaking into core logic.
- Duplicate event is ignored.
- Unknown provider user ID and disconnected connection failures are recorded safely.
- Fetch failures and rate limits produce retryable/support-visible state.

### Phase 5: Mobile Strava Connection UX

Goal: let beta users connect Strava and see import state without mobile owning provider credentials.

Mobile changes:

- Replace Garmin-only provider assumptions in the Plan tab connection surface with Strava for private beta.
- Enable `Connect Strava` only when `StartProviderConnection` returns `oauth_ready=true` and a non-empty authorization URL.
- Open the backend-provided authorization URL externally.
- Show provider states: not connected, pending, connected, syncing, delayed/rate-limited, error/reconnect-needed, and disconnected.
- Continue reading plan, imported activity, match, workout result, and adaptation state through `GetCurrentPlanWeek`.
- Keep workout completion read-only from mobile; imported Strava activities drive completion through backend services.

Tests:

- Widget/model tests for all connection states.
- Connect button remains disabled when backend reports OAuth not ready.
- Connect button opens returned authorization URL when ready.
- Plan refresh reflects imported activity and adaptation state after backend updates.

### Phase 6: Beta Operations, Privacy, and Support

Goal: make the Strava integration safe enough for private beta users.

Backend changes:

- Add production-safe structured logs for OAuth, token refresh, backfill, webhook, import, match, completion, and adaptation failures.
- Add support diagnostics for provider connection status, last sync time, last error, import counts, and recent failed import events without exposing token material.
- Define and implement disconnect behavior:
  - Stop future imports.
  - Remove token material or invalidate token references.
  - Preserve or remove historical normalised activities according to the beta privacy decision.
- Define account deletion/export behavior for Strava-derived data before live beta.
- Document rate-limit behavior and operational response for provider outages.

Tests:

- Disconnect stops import and removes token material.
- Revoked access transitions connection into a recoverable state.
- Support diagnostics redact tokens and raw sensitive payload fields.
- Account deletion/export behavior covers provider connections, token references, provider activities, imported activities, workout results, and adaptation events as decided.

## Acceptance Criteria

The Strava integration is private-beta ready when:

- A beta user can connect Strava through backend-owned OAuth.
- Token material is stored securely and never exposed to mobile or core domain logic.
- Initial backfill imports supported running activities and records ignored/failure states.
- Webhooks keep new, updated, and deleted Strava activity state visible to backend workflows.
- Imported Strava activities are normalised into `ImportedActivity` and use provider-neutral matching, result, and adaptation services.
- Mobile shows Strava connection/import state and does not perform provider API calls directly.
- Backend and mobile tests cover success, retry, failure, rate-limit, duplicate, unsupported activity, disconnect, and revoked-access paths.
- Strava-derived activity data is not sent to AI prompts or used for model training.

## Assumptions

- Strava is the first live activity provider for private beta.
- Garmin remains disabled until Garmin access findings are validated.
- The backend owns OAuth, token refresh, webhook verification, API calls, imports, rate limits, and retries.
- Flutter only starts the connect flow, opens the returned authorization URL, and reads Runthread state.
- Payment, subscriptions, and AI explanation generation are out of scope for this plan.
- Production auth/current-athlete lookup is required before live beta use, but this plan keeps provider integration work isolated enough to wire auth later.
