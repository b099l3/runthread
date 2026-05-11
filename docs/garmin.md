# Garmin Integration

Garmin is a future direct integration target for Runthread.

Strava is now the first likely MVP activity provider because it can help validate the activity import loop faster. Garmin remains important as a premium/direct integration target after the provider-neutral import pipeline is proven.

The integration should be isolated from the core domain. Garmin-specific identifiers, payloads, API details, and provider terminology should stay inside the integration and persistence layers. Core training logic should use Runthread domain models such as `ImportedActivity`, `WorkoutMatch`, and `WorkoutResult`.

Garmin and Strava should share the same provider-neutral activity import pipeline:

```text
Provider connection
  -> provider-specific delivery or fetch
  -> provider adapter normalisation
  -> ImportedActivity
  -> workout matching
  -> workout result
  -> adaptation event
```

Garmin-specific fields must not leak into core planning, matching, workout result, or adaptation logic. If a Garmin field is useful to the core product, it should first be mapped to an explicit provider-neutral field.

Real Garmin API access is not assumed to be available during early development. Garmin's current access process, API capabilities, provider terms, webhook or polling model, and test data options must be validated before implementing production OAuth or import flows.

## Intended Flow

1. The user connects Garmin from the mobile app when Garmin direct integration is available.
2. The backend completes the provider connection flow.
3. The backend stores a provider connection for the athlete.
4. Garmin activities are imported by webhook, polling, or another supported provider mechanism.
5. Raw provider data is stored only where useful for audit/debugging.
6. Activities are normalised into Runthread's `ImportedActivity` model.
7. Imported activities are matched to planned workouts.
8. Matched activities produce workout results.
9. Workout results may trigger adaptation events.

## Future Garmin Flow

Garmin work should start by validating Garmin access before implementing OAuth. The first useful output is confidence about Garmin's current developer approval path, data access model, test data options, and whether activities arrive by webhook, polling, or another supported delivery model.

Once access is understood, the backend flow should be:

1. Athlete starts a Garmin connection from the mobile app.
2. Backend creates or resumes a `provider_connection` record in a pending state.
3. Backend sends the athlete through Garmin's supported authorization flow.
4. Backend stores the connected provider account metadata and token metadata allowed by Garmin's terms.
5. Backend imports new Garmin activities by the supported delivery mechanism.
6. Garmin-specific activity records are stored in provider tables for idempotency, audit, and retry.
7. Each relevant Garmin activity is normalised into the provider-neutral `domain.ImportedActivity`.
8. Matching links the imported activity to a planned workout.
9. Workout result and adaptation logic run exactly as they do for provider-neutral activities.
10. Flutter continues to read plan, completion, and adaptation state through `GetCurrentPlanWeek`.

Garmin direct integration should not require Garmin details in the planner, matcher, adaptation engine, or mobile plan views.

## Garmin Access Validation Checklist

Before enabling `StartProviderConnection`, implementing OAuth, or adding provider callbacks, Runthread needs written answers to the following questions. The answers should be captured in this document or a linked decision record before production-facing Garmin work starts.

Use `docs/garmin-access-findings.md` to record validated answers. That file is the decision gate for moving into real Garmin implementation.

The same findings file includes the first external validation tasks in recommended order. Complete those tasks before changing protobuf contracts, enabling mobile connection actions, adding callbacks, or implementing sync.

The Garmin work is now at a preparation pause point. Provider persistence, provider import orchestration, provider connection RPC placeholders, and the read-only mobile Garmin status surface are ready to continue from, but production Garmin implementation remains blocked by ADR-0007 until the findings file has validated answers.

Access and approval:

- What Garmin developer, partner, or API access program does Runthread need for activity import?
- Is approval required before development, before production launch, or both?
- Are there separate sandbox, beta, and production credentials?
- What app information, privacy policy, security information, callback URLs, or business details are required for approval?
- Are there restrictions on commercial coaching, training-plan, or subscription products using imported Garmin data?

Authorization model:

- Is the connection flow OAuth, another Garmin-managed authorization flow, webhook subscription, partner provisioning, or a hybrid?
- Does Runthread receive user-level tokens, provider-managed subscriptions, webhook-only deliveries, or another credential shape?
- What redirect URI, callback URI, state, and CSRF requirements apply?
- Does authorization happen in an external browser, embedded web view, Garmin Connect app, or another surface?
- Does the current `StartProviderConnection` response shape need additional fields beyond `authorization_url`, `state`, and `oauth_ready`?

Scopes and activity data:

- Which scopes or permissions are required to read running activities?
- Which activity fields are available: provider activity ID, activity type, start time, duration, distance, pace, heart rate, elevation, laps, splits, samples, device, and timezone?
- Are historical activities available after connection, or only new activities after authorization?
- Are updates/deletes sent when an activity changes in Garmin Connect?
- Which activity types should be imported, ignored, or treated as non-running cross-training?

Activity delivery:

- Are activities delivered by webhook, polling, batch export, manual refresh, or a Garmin-specific push mechanism?
- If webhooks are available, what event types, retry behavior, delivery IDs, and signature verification rules apply?
- If polling is required, what cursor, pagination, time-window, and backoff rules apply?
- How soon after Garmin activity completion should Runthread expect activity data to arrive?
- Can the same provider activity be delivered more than once, and what fields should define idempotency?

Sandbox and test data:

- Is there a sandbox environment or test athlete account flow?
- Can Runthread create representative test runs, edited activities, deleted activities, failed deliveries, and duplicate deliveries?
- Can local development receive webhook callbacks, or is a public tunnel/staging deployment required?
- Are there provider tools for replaying webhook events or import batches?

Data storage, retention, and privacy:

- May Runthread store raw Garmin activity payloads? If yes, for how long and for what support/debugging purposes?
- Must raw payloads be redacted, encrypted, access-limited, or omitted entirely?
- May Runthread store token metadata, refresh tokens, provider account IDs, and import cursors?
- What user deletion, disconnect, export, and data retention obligations apply?
- Are there provider branding or disclosure requirements in the mobile app?

Operations and limits:

- What rate limits apply to authorization, activity import, polling, and retry?
- What error codes should trigger retry, reconnect, support escalation, or permanent disconnect?
- How should token expiry, refresh failure, revoked access, and provider outages be represented?
- What monitoring, alerting, and audit logs are required before beta launch?

Disconnect and revocation:

- Can users disconnect Garmin from Runthread, Garmin Connect, or both?
- Does Garmin notify Runthread when access is revoked?
- Should disconnect remove tokens only, stop imports, delete raw payloads, or delete normalised imported activities too?
- Should reconnect reuse the existing `ProviderConnection` record or create a new connection?

Answers that would change backend boundaries:

- If Garmin is webhook-only, add a provider-facing webhook handler and signature verification before polling or manual sync.
- If Garmin is polling-only, add a scheduler/manual sync trigger plus cursor persistence before webhook work.
- If Garmin does not allow raw payload storage, make `ProviderActivityPayload` optional-by-policy and avoid storing payload bytes by default.
- If Garmin activity IDs are not globally unique, keep idempotency scoped by provider connection ID and provider activity ID.
- If historical backfill is available, add an import job boundary that can process batches without assuming a single workout match.
- If activity updates/deletes are delivered, extend provider import orchestration to handle update and delete events rather than only received/processed/ignored/failed.
- If token storage is allowed, add encrypted token storage outside core domain tables before production OAuth.
- If token storage is not allowed, keep provider connection records metadata-only and rely on Garmin's supported delivery mechanism.

Answers that would change mobile UX:

- If Garmin requires external browser authorization, `Connect Garmin` should open the returned `authorization_url` and show pending state until callback completion.
- If Garmin uses a provider-managed connection surface instead of OAuth URLs, the mobile copy and action should reflect that provider flow.
- If connection approval is delayed or asynchronous, pending state should explain that Runthread is waiting for Garmin access rather than asking the athlete to retry.
- If historical backfill is available, connected state should show initial sync/import progress.
- If activity delivery can be delayed, the workout detail empty state should say that Garmin activity may still be syncing.
- If disconnect must delete data, the future disconnect flow needs explicit confirmation copy.
- If Garmin requires branding or disclosure copy, the Plan tab provider section and future connection screens must include it.

## Mock Provider Import Boundary

Earlier preparation includes a small mock import service in `services/api/internal/garmin`.

The service is still not real Garmin integration. It keeps mock Garmin payload handling in the Garmin package, then delegates persistence to the provider-neutral `services/api/internal/providerimport` orchestration service. It:

- accepts a `ProviderConnection` and mock Garmin activity payload
- normalises the activity into provider-neutral `domain.ImportedActivity`
- builds a `providerimport.ImportRequest`
- uses the provider import orchestration service to store provider activity state, optional raw payload, imported activity, provider activity link, and import event status

This keeps Garmin-shaped payload handling inside the Garmin package while preserving the core domain boundary. It is not wired into ConnectRPC, mobile, startup, or real OAuth.

An in-memory test now exercises the prepared boundary end to end: seed athlete, goal, plan week, and provider connection; import a mock Garmin activity; persist provider activity, raw payload, imported activity, and import event; match the imported activity to a workout; create a workout result; and run adaptation logic. This proves the service boundaries fit together before real Garmin access or endpoint work begins.

Stage 15 preparation adds a small bridge in `services/api/internal/providers/garmin` that implements the shared `providers.ActivityProvider` interface for mock Garmin payloads. It delegates to the existing `services/api/internal/garmin` mock normaliser instead of moving or rewriting Garmin logic.

Current package split:

- `services/api/internal/garmin`: legacy mock Garmin payload structs, mock normalisation, and mock import service tests.
- `services/api/internal/garminadapter`: bridge from legacy mock Garmin payloads into app-service completion orchestration.
- `services/api/internal/providers/garmin`: provider-neutral package anchor and mock `ActivityProvider` bridge.

Future cleanup, after Garmin access findings are validated, should move or replace legacy mock boundaries deliberately. Until then, keeping the bridge small avoids churn while proving Garmin can share the same provider-neutral shape Strava uses.

## Garmin Readiness Review

Reusable for real Garmin:

- Provider persistence models: `ProviderConnection`, `ProviderActivity`, `ProviderActivityPayload`, and `ProviderImportEvent` are provider-oriented rather than mock-only.
- Provider repository interfaces and stores: the in-memory and Postgres-backed implementations can support real import orchestration without exposing Garmin details to core training logic.
- Core domain boundary: real Garmin activities should still normalise into `domain.ImportedActivity` before matching, workout result creation, or adaptation.
- Provider activity idempotency shape: real imports can use provider connection ID plus Garmin activity ID to avoid duplicate imported activities.
- Import event audit shape: webhook deliveries, polling attempts, manual refreshes, and failures can all be represented as provider import events.
- Mock e2e test shape: the sequence of provider import, normalisation, matching, result creation, and adaptation is the right production flow even though the payload source is fake.

Still mock-only:

- `MockActivityPayload` is not a real Garmin API contract.
- `NormalizeMockActivity` only handles representative local fields and simple activity type labels.
- `MockImportService` uses deterministic local IDs and does not process Garmin authorization, token refresh, pagination, webhook signatures, or retry semantics.
- `services/api/internal/providers/garmin.MockProvider` is only a mock adapter for the shared `ActivityProvider` interface.
- Raw payload storage currently accepts caller-provided bytes without provider terms, privacy, retention, or redaction policy.
- The import flow is not wired into app services, ConnectRPC, server startup, scheduler jobs, mobile UI, or Postgres integration tests.

## Next Production-Facing Boundary

The internal provider import orchestration service lives in `services/api/internal/providerimport`. It is still internal-only and has no HTTP, RPC, startup, scheduler, auth, or OAuth wiring.

Current shape:

- Inputs: athlete ID, provider connection ID, provider delivery metadata, and a provider activity envelope supplied by a Garmin adapter.
- Dependencies: `repository.ProviderStore` and `repository.ImportedActivityRepository`.
- Garmin adapter responsibility: validate Garmin-specific payloads, map Garmin activity fields into a normalisation input, expose provider activity ID/type/start time, and provide raw payload bytes if storage is allowed.
- Orchestration responsibility: load and validate the provider connection, create or update `ProviderActivity`, record `ProviderImportEvent`, optionally store `ProviderActivityPayload`, save `ImportedActivity`, link provider activity to imported activity, and record processed, ignored, or failed terminal state.
- Output: provider connection, provider activity, optional imported activity, import event, payload storage flag, and a clear error for retry or support.

This boundary should remain independent of the trigger mechanism. The same orchestration service should be callable later from a webhook handler, scheduled poller, manual refresh RPC, or local test command.

Matching, workout result creation, and adaptation remain outside this package for now. They should be connected by an application service after the import trigger shape and current-plan lookup are clearer.

## Provider Import Completion Boundary

`services/api/internal/app.ProviderImportService` now connects provider import orchestration to the existing core loop pieces without adding any endpoint or startup wiring.

The service:

- imports a provider activity through `providerimport.Service`
- accepts or loads a `PlanWeek`
- accepts, loads, or finds the relevant `PlannedWorkout`
- matches the imported activity to the planned workout with existing matching logic
- creates a `WorkoutResult` through the existing domain workout completion flow
- runs adaptation logic when the result requires it
- persists the workout match, workout result, updated workout, updated plan week, and adaptation event when present

Garmin-specific parsing still stays outside this app service. A real Garmin adapter should feed already-normalised provider import data into this boundary after authorization and delivery handling are implemented.

## Mock Garmin App Adapter

`services/api/internal/garminadapter` now contains a small mock adapter that turns `garmin.MockActivityPayload` into `app.CompleteProviderImportRequest`.

This package exists to keep the dependency direction clean:

- `garmin` owns Garmin-shaped payload parsing and mock normalisation.
- `providerimport` owns provider-neutral import persistence.
- `app.ProviderImportService` owns matching, workout result creation, adaptation, and core persistence.
- `garminadapter` is the thin bridge used by tests and future trigger code.

It is still internal-only and has no endpoint, server startup wiring, OAuth handling, or mobile UI.

## Pre-Endpoint Backend Readiness

The backend is now close enough to define production-facing Garmin API boundaries, but not to implement real Garmin access yet.

Ready to expose later through RPC or provider triggers:

- `repository.ProviderStore` can persist provider connections, provider activities, provider payload snapshots, and provider import events.
- `providerimport.Service` can idempotently receive provider activity metadata, save optional raw payload bytes, save a provider-neutral `ImportedActivity`, link it back to the provider activity, and record processed, ignored, or failed import events.
- `app.ProviderImportService` can take a completed provider import through matching, workout result creation, conservative adaptation, and core persistence.
- `app.ProviderActivityMatchService` and `app.ProviderActivityCompletionService` can separately persist match/review state, create workout results, and run deterministic adaptation after a confident match.
- `garminadapter` proves the desired dependency direction with mock Garmin data: Garmin-shaped input becomes provider-neutral app-service input before core training logic sees it.
- `providers/garmin.MockProvider` proves Garmin-shaped mock payloads can be exposed through the same `ActivityProvider` interface as Strava mock payloads.

Still blocked by Garmin access validation:

- The supported connection flow and whether Garmin uses OAuth, partner APIs, webhooks, polling, or another delivery model for Runthread's use case.
- Provider account identifiers, available activity fields, data scopes, rate limits, retry semantics, and test account or sandbox support.
- Whether raw activity payloads and token metadata can be stored, and what retention, redaction, encryption, or support-access rules Garmin requires.
- Webhook signature verification or polling cursor rules, depending on the available import mechanism.

Still mock-only:

- `garmin.MockActivityPayload`, `NormalizeMockActivity`, and `garminadapter.BuildMockCompleteProviderImportRequest`.
- The deterministic local ID strategy for mock imported activities.
- Request-supplied plan week and planned workout context in `app.ProviderImportService`; production should load the current plan/workout for the authenticated athlete.

Smallest endpoint plan, once Garmin access is understood:

1. Add provider connection RPCs for starting a Garmin connection, reading connection status, and disconnecting.
2. Add the provider callback, webhook, poller, or manual sync trigger required by Garmin's actual delivery model.
3. Keep activity import endpoints internal/provider-facing, not mobile completion actions.
4. Have the trigger call a Garmin adapter that validates Garmin data and feeds `providerimport.Service` or `app.ProviderImportService`.
5. Continue exposing mobile state through `GetCurrentPlanWeek` until a narrower read model is needed.

## Provider Connection API Contract

The protobuf contract now defines the first two provider connection RPC shapes:

- `GetProviderConnectionStatus`
- `StartProviderConnection`

These are intentionally provider-neutral contracts with a `provider` enum that currently allows `garmin`. The Go handlers are implemented as readiness placeholders backed by `repository.ProviderStore`; no OAuth, callbacks, mobile screens, auth, or provider startup wiring exists yet.

Temporary placeholder fields:

- `athlete_id`: request-supplied until auth/session context exists.
- `provider_connection_id`: optional lookup hint until the backend has current-athlete provider connection lookup.
- `redirect_uri`: included for likely OAuth-style flows, but Garmin's actual connection mechanism still needs validation.
- `authorization_url`: reserved for the provider authorization URL if Garmin's supported flow returns one.
- `state`: reserved for CSRF/session correlation if the real connection flow needs it.
- `oauth_ready`: explicitly false until real Garmin access is validated and implemented.

Boundary decision:

- `placeholder_note` was removed from the proto. Development caveats belong in docs and tests, not in the mobile-facing API contract.
- `redirect_uri`, `authorization_url`, and `state` stay in the proto because they are the smallest useful fields for a likely OAuth-style provider flow. If Garmin requires a different connection model, these fields can remain empty while the trigger/callback shape is adjusted.

Current behavior:

- `StartProviderConnection` supports `provider = garmin`, creates a pending provider connection when none exists, and reuses an existing pending connection for the athlete.
- `StartProviderConnection` always returns `oauth_ready=false` and an empty `authorization_url` until Garmin access is validated.
- `GetProviderConnectionStatus` returns an existing connection for the request-supplied athlete/provider or `has_connection=false`.
- The application service uses `repository.ProviderStore`, and current tests exercise the in-memory repository path only.
- HTTP-level tests now use the mounted ConnectRPC service and generated client to verify `StartProviderConnection` followed by `GetProviderConnectionStatus` through the real server mux.

The next implementation step should tighten the provider connection boundary before any mobile UI by reviewing whether the request/response shape is sufficient for a future Garmin connection screen.

## Mobile Connection Entry Point

The Flutter MVP now has a small read-only Garmin connection surface on the main Plan tab. It appears near the week header and before plan adaptation summaries, because Garmin status affects how workouts become completed but should not distract from the plan itself.

Current behavior:

- Shows provider `Garmin`.
- Calls `GetProviderConnectionStatus` through the hand-written Connect JSON mobile API client.
- Sends demo `athlete_id = athlete-1` and `provider = garmin`; this stays temporary until auth and current-athlete lookup exist.
- Shows `Not connected`, `Pending`, `Connected`, `Syncing`, `Needs attention`, or `Disconnected` when the backend returns a status.
- Shows a disabled `Connect Garmin` action.
- Explains that run completion will come from imported Garmin activity once provider access is ready.
- Falls back to the read-only `Not connected` state when provider status cannot be loaded.
- Does not call `StartProviderConnection` from Flutter yet.

Planned states:

- `not connected`: show a calm status and enabled connect action once Garmin access is validated.
- `pending`: show that connection has started and the athlete may need to finish provider authorization.
- `connected`: show connected state, last sync/import status, and the most recent successful activity import time when available.
- `syncing`: show that imports are being processed without implying the plan has changed yet.
- `error`: show recovery text and a retry/reconnect path after the backend has real provider error semantics.
- `disconnected`: show that Runthread is no longer importing Garmin activity and offer reconnect when supported.

The near-term mobile implementation should keep using `GetCurrentPlanWeek` for plan, activity, completion, and adaptation state. `GetProviderConnectionStatus` is read-only in Flutter for now; enabling the connect action should wait until Garmin access and the real authorization flow are validated.

## Provider Connection Concepts

`ProviderConnection` is the athlete-owned link between Runthread and an external activity provider.

Practical fields:

- app-owned connection ID
- athlete ID
- provider name, initially `garmin`
- provider athlete/account/user ID when Garmin provides one
- status
- connected/disconnected timestamps
- last successful sync timestamp
- last import cursor or checkpoint if the provider supports one
- token metadata or encrypted token references if Garmin's terms allow token storage
- last error text for support and recovery

Connection statuses should stay simple:

- `pending`: connection started but not completed
- `connected`: provider account is usable for imports
- `syncing`: import is currently running
- `error`: connection or sync failed and may need retry
- `disconnected`: athlete or provider disconnected the account

Lifecycle:

1. Create pending connection.
2. Mark connected after provider authorization succeeds.
3. Update sync metadata after each import attempt.
4. Move to error if authorization, token refresh, webhook processing, or polling fails.
5. Move to disconnected when the athlete disconnects or provider access is revoked.
6. Reconnect by creating a new connection or reactivating the existing one, depending on Garmin's terms and identifiers.

Core domain logic should not depend on `ProviderConnection`. It belongs to integration and persistence boundaries.

## Provider Activity Concepts

`ProviderActivity` records a provider-specific activity before or alongside normalisation.

Practical fields:

- app-owned provider activity ID
- provider connection ID
- athlete ID
- provider name, initially `garmin`
- provider activity ID from Garmin
- provider activity type string from Garmin
- provider start time
- import status
- linked `imported_activity_id` once normalisation succeeds
- first seen/imported timestamp
- last sync timestamp
- last error text

Import statuses should stay simple:

- `received`: provider activity was seen
- `normalised`: provider activity produced an `ImportedActivity`
- `ignored`: provider activity is not relevant to Runthread, such as unsupported activity type
- `failed`: normalisation or persistence failed and can be retried

The unique key should be provider plus provider activity ID, scoped by connection if Garmin IDs are not globally unique.

## Garmin-Specific Data Boundary

Garmin-specific data may be stored outside the core domain for integration reliability:

- Garmin provider user/account ID
- Garmin activity ID
- Garmin activity type string
- Garmin webhook delivery ID or polling cursor
- raw Garmin activity payload for audit/debugging if allowed by Garmin's terms
- token metadata, encrypted token references, expiry, and refresh state if allowed
- import status, retry count, last error, and delivery timestamps

Garmin-specific data must not be stored in `domain.ImportedActivity`, `WorkoutMatch`, `WorkoutResult`, planning rules, matching rules, or adaptation rules.

`domain.ImportedActivity` remains provider-neutral and should contain only Runthread's normalised activity facts: athlete, activity type, start time, duration, distance, pace, and heart rate.

## Provider Persistence Targets

Provider schema design should keep provider tables outside the core domain model:

- `provider_connections`
- `provider_activities`
- `provider_activity_payloads`
- optionally `provider_import_events` if webhook/polling delivery audit is needed immediately

These tables should exist before real OAuth or provider callbacks so integrations can be idempotent and debuggable from the first real import.

## Stage 4 Mock Import

Stage 4 introduces a mock Garmin activity importer in `services/api/internal/garmin`.

The mock importer accepts representative Garmin-shaped activity payloads with Garmin-specific fields such as:

- Garmin activity ID
- Garmin activity type
- start time
- duration
- distance
- average heart rate

It normalises those payloads into the provider-neutral `domain.ImportedActivity` model.

The current mock mapping supports:

- Garmin road running labels to `run`
- Garmin trail running labels to `trail_run`
- Garmin treadmill labels to `treadmill`
- Garmin walking labels to `walk`
- unknown activity labels to `other`

This stage does not include real Garmin OAuth, webhooks, polling, persistence, ConnectRPC endpoints, or raw payload storage.

## Boundary Rule

Garmin-specific data must not leak into core domain logic.

The planning, matching, and adaptation packages should depend on provider-neutral concepts. If COROS or Apple Health is added later, those integrations should also normalise into the same domain models.
