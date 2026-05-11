# Architecture Decisions

This file records important architecture decisions for Runthread. Add new decisions when the direction changes or when a meaningful tradeoff is resolved.

## ADR-0001: Use Go Backend

Status: Accepted

Decision: Runthread will use Go for the backend service and target Go 1.22 or newer.

Reason: Go is simple to deploy, works well for API services, has strong typing, and fits deterministic domain logic and integration workflows.

## ADR-0002: Use ConnectRPC

Status: Accepted

Decision: The Flutter app will communicate with the Go backend through ConnectRPC.

Reason: ConnectRPC gives a typed service contract while keeping HTTP-friendly operational behavior.

## ADR-0003: Use Postgres

Status: Accepted

Decision: Runthread will use Postgres as the primary application database.

Reason: The product needs reliable relational data for athletes, plans, workouts, activities, matches, and adaptation history.

## ADR-0004: Keep AI Out of Core Training Decisions

Status: Accepted

Decision: AI may explain decisions or help with copy, but deterministic code owns training plans and adaptations.

Reason: Training changes must be explainable, testable, auditable, and consistent.

## ADR-0005: Superseded Garmin-First Provider Plan

Status: Superseded by ADR-0008 and ADR-0012

Historical decision: Garmin was originally the first provider integration. COROS and Apple Watch were deferred.

Reason: Focusing on one provider reduces integration complexity and matches the initial target user. The domain should still remain provider-neutral.

Note: Real Garmin access is assumed to require validation or application before production integration work begins.

## ADR-0006: Use asdf for Tool Versions

Status: Accepted

Decision: Runthread will pin local development tool versions with repo-level `.tool-versions`.

Reason: Future sessions and local development should use the same toolchain without relying on machine-global defaults.

Initial pinned tool:

- Go `1.25.5`

## ADR-0007: Block Real Garmin Implementation Until Access Findings Are Validated

Status: Accepted

Decision: Real Garmin OAuth, callbacks, webhooks, polling, provider sync, and mobile connection actions are blocked until Garmin access findings are recorded in `docs/garmin-access-findings.md`.

Reason: Garmin access, authorization, activity delivery, data storage, rate limits, test data, and revocation rules can change the backend provider boundary, persistence policy, and mobile connection UX. Implementing production integration before those answers are validated would risk building the wrong flow.

Implications:

- `StartProviderConnection` remains a readiness placeholder.
- Flutter may read provider connection status, but `Connect Garmin` stays disabled.
- Provider import and app-service orchestration can continue using mock Garmin payloads for boundary tests.
- Any real provider implementation must first update the findings file and then document required contract, persistence, backend, or mobile changes.

## ADR-0008: Add Strava as First Real MVP Activity Provider

Status: Accepted

Decision: Strava will be the first likely real MVP activity import provider. Garmin direct integration remains important, but it moves after the provider-neutral import path is proven.

Reason: Strava can help validate the activity import loop faster because many runners already aggregate activities there and its OAuth/API path is likely more accessible for MVP testing than Garmin direct access.

Implications:

- The product is no longer Garmin-first in the MVP plan.
- Strava work starts with docs and mock payloads before real OAuth or API calls.
- The same provider-neutral pipeline must remain usable by Garmin later.

## ADR-0009: Keep Provider-Specific Data Out of Core Domain Logic

Status: Accepted

Decision: Provider-specific payloads, activity IDs, scopes, webhook bodies, token metadata, rate-limit state, and raw provider fields stay in provider packages and provider persistence boundaries. Core planning, matching, workout result, and adaptation logic uses provider-neutral `ImportedActivity` and related domain concepts.

Reason: Runthread should support Strava, Garmin, COROS, Apple Health, and future providers without coupling training behavior to one provider's schema.

Implications:

- Provider adapters normalise external data before core domain services see it.
- A provider field must be mapped to an explicit provider-neutral field before it affects matching or adaptation.
- Tests for provider mappings live near provider packages; tests for planning/adaptation stay provider-neutral.

## ADR-0010: Backend Owns OAuth, Token Refresh, Webhooks, and Activity Import

Status: Accepted

Decision: The backend owns provider OAuth code exchange, token storage, token refresh, webhook verification, import jobs, provider API calls, and provider rate-limit behavior. The mobile app may start connect flows and open authorization URLs, but it does not store provider tokens or call provider APIs directly.

Reason: Provider credentials and import workflows need central security, retry, audit, idempotency, and rate-limit handling.

Implications:

- Provider connection tokens must be stored securely outside core domain models.
- Webhook and import endpoints are backend/provider-facing surfaces.
- Flutter reads provider-neutral status and activity/plan state through Runthread APIs.

## ADR-0011: Do Not Send Strava-Derived Activity Data to AI Systems

Status: Accepted

Decision: Strava-derived activity data must not be sent to AI prompts and must not be used for AI model training.

Reason: Imported activity data is sensitive user fitness data. The MVP can validate deterministic planning and adaptation without exposing provider activity data to AI systems.

Implications:

- AI is not part of Strava import, matching, or adaptation stages.
- Any future explanation feature must avoid Strava-derived activity data unless a later explicit privacy and product decision changes this.
- Strava activity data should only be shown to the authorised user who connected the account.

## ADR-0012: Keep Garmin Direct Integration on the Roadmap

Status: Accepted

Decision: Garmin direct integration remains on the roadmap as a future premium/direct activity provider after the Strava-first MVP import path.

Reason: Garmin remains important for serious runners and direct device-platform integration, but its access path and provider requirements should not block validation of the activity import loop.

Implications:

- Garmin access validation remains required before production Garmin work.
- Garmin must share the provider-neutral import, normalisation, matching, and adaptation pipeline.
- Garmin-specific fields must not leak into core planning or adaptation logic.
