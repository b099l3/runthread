# Garmin Access Findings

This file records validated Garmin access findings before real Garmin OAuth, callbacks, polling, webhooks, or production sync are implemented.

Do not treat assumptions as findings. Each row should be backed by a source, a date checked, and a clear confidence level.

## Status

Overall status: Not validated.

Blocking decision: Real Garmin implementation remains blocked until the required access, authorization, activity delivery, storage, and operational findings below are completed.

## Pause-Point Review

Stage 8 Garmin preparation is at a deliberate pause point. The codebase has enough internal shape to continue after external Garmin validation, but it should not move into real provider implementation yet.

Ready:

- Provider persistence tables, sqlc queries, in-memory repositories, and Postgres-backed repositories exist for provider connections, provider activities, provider activity payloads, and provider import events.
- `providerimport.Service` provides a provider-neutral persistence boundary for imported provider activity metadata and normalised `ImportedActivity` records.
- `app.ProviderImportService` connects provider import results to matching, workout result creation, adaptation, and core persistence without exposing Garmin-specific payloads to core training logic.
- `garminadapter` proves the dependency direction for mock Garmin payloads.
- `GetProviderConnectionStatus` and `StartProviderConnection` exist as readiness RPCs.
- Flutter has a read-only Garmin status surface and keeps `Connect Garmin` disabled.
- ADR-0007 blocks real Garmin implementation until validated findings are recorded.

Intentionally blocked:

- Real Garmin OAuth or any other production authorization flow.
- Callback handlers, webhook handlers, polling jobs, manual sync triggers, and real provider import scheduling.
- Storing production Garmin tokens or raw production Garmin payloads.
- Enabling `Connect Garmin` in Flutter.
- Changing provider protobuf contracts based on assumptions.
- Postgres provider store wiring in server startup.
- Live Garmin or live database integration tests.

Must be validated externally:

- Garmin access/approval path and commercial use constraints.
- Authorization model and whether the current `StartProviderConnection` shape is sufficient.
- Activity scopes and field availability.
- Activity delivery mechanism, retry behavior, idempotency, and latency.
- Historical import, activity edit, and activity delete support.
- Sandbox, test account, staging callback, and replay options.
- Raw payload, token, cursor, provider account ID, retention, deletion, encryption, and support-access rules.
- Rate limits, operational requirements, disconnect behavior, and revocation notifications.

Future session read list after findings are available:

- `README.md`
- `docs/garmin.md`
- `docs/garmin-access-findings.md`
- `docs/decisions.md`
- `docs/roadmap.md`
- `docs/persistence.md`
- `services/api/proto/runthread/v1/runthread.proto`
- `services/api/internal/app/provider_connection.go`
- `services/api/internal/app/provider_import.go`
- `services/api/internal/providerimport`
- `services/api/internal/garmin`
- `services/api/internal/garminadapter`
- `services/api/internal/repository/provider.go`
- `services/api/internal/postgres/provider_repositories.go`
- `services/api/internal/rpc/handler`
- `apps/mobile/lib/src/api/runthread_api.dart`
- `apps/mobile/lib/src/models/provider_connection.dart`
- `apps/mobile/lib/src/plan_week_screen.dart`

## External Validation Plan

These tasks should be completed outside the codebase before any real Garmin implementation starts. Record each validated answer in the matching finding section below.

1. Confirm the Garmin access path.
   Why it matters: Runthread needs to know whether Garmin activity import is available through a public developer flow, partner approval, business application, or another access path.
   Evidence to collect: official Garmin program name, application URL or contact path, approval requirements, sandbox/production distinction, and any commercial-use restrictions.
   Fills: Access And Approval.

2. Confirm the authorization model.
   Why it matters: the current `StartProviderConnection` contract assumes an OAuth-like URL/state flow, but Garmin may require a different provider-managed connection process.
   Evidence to collect: authorization flow description, redirect/callback requirements, token or subscription shape, state/CSRF requirements, and whether authorization happens in browser, app, or provider portal.
   Fills: Authorization Model.

3. Confirm required scopes and available activity fields.
   Why it matters: normalisation into `ImportedActivity`, matching quality, and workout result creation depend on which activity fields Garmin provides.
   Evidence to collect: required permissions/scopes and available fields for activity ID, type, start time, duration, distance, pace, heart rate, timezone, laps, splits, samples, elevation, and device data.
   Fills: Scopes And Activity Data.

4. Confirm activity delivery mechanism.
   Why it matters: webhook, polling, batch, and manual refresh models require different backend triggers and operational behavior.
   Evidence to collect: delivery mechanism, event types, latency expectations, retry semantics, duplicate delivery behavior, webhook signature rules, polling cursor rules, pagination, and backoff limits.
   Fills: Activity Delivery and Operations And Limits.

5. Confirm historical import and update/delete behavior.
   Why it matters: backfill and activity updates may require batch import jobs and update/delete handling rather than the current single received/processed flow.
   Evidence to collect: whether historical activities are available, how far back imports can go, whether edited/deleted activities are delivered, and how provider activity identity is maintained.
   Fills: Scopes And Activity Data and Activity Delivery.

6. Confirm sandbox, test account, and staging callback options.
   Why it matters: Runthread needs a repeatable way to test connection, import, duplicate delivery, failures, and retries before beta.
   Evidence to collect: sandbox availability, test user setup, sample payloads, event replay tooling, webhook tunneling or staging requirements, and approval needed for test data.
   Fills: Sandbox And Test Data.

7. Confirm payload, token, and provider metadata storage rules.
   Why it matters: storage policy determines whether `ProviderActivityPayload`, token metadata, encrypted token references, and import cursors can be used in production.
   Evidence to collect: raw payload retention allowance, token storage requirements, encryption/redaction obligations, support access limits, provider user ID rules, import cursor storage rules, and deletion/export requirements.
   Fills: Data Storage And Privacy.

8. Confirm rate limits, retry rules, and operational obligations.
   Why it matters: imports need safe retry/backoff behavior, monitoring, and failure handling before beta users rely on activity completion.
   Evidence to collect: rate limits, error codes, retryable/permanent failure categories, provider outage guidance, monitoring/audit expectations, and support escalation requirements.
   Fills: Operations And Limits.

9. Confirm disconnect and revocation behavior.
   Why it matters: Runthread needs to know how to stop imports, remove credentials, preserve or delete historical data, and represent disconnected status.
   Evidence to collect: user-initiated disconnect flow, Garmin-side revocation notification, reconnect behavior, token deletion requirements, raw payload deletion requirements, and normalised activity retention rules.
   Fills: Disconnect And Revocation and Data Storage And Privacy.

10. Review contract and UX impacts after findings are filled.
    Why it matters: validated answers may require changes to protobuf fields, provider persistence, app-service orchestration, scheduler/webhook boundaries, or mobile connection copy.
    Evidence to collect: a short implementation decision summary that references completed findings and lists required backend/mobile changes.
    Fills: Decision Gate.

## Finding Template

Copy this template for each finding.

```text
Topic:
Question:
Answer:
Source/link:
Date checked:
Confidence: low | medium | high
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open | waiting | resolved
Notes:
```

## Required Findings

### Access And Approval

```text
Topic: Access and approval
Question: Which Garmin developer, partner, or API access program does Runthread need for activity import?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

```text
Topic: Access and approval
Question: What approval, app information, privacy policy, callback URL, security, and business details are required before development or production launch?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

### Authorization Model

```text
Topic: Authorization model
Question: Is Garmin connection OAuth, another Garmin-managed authorization flow, webhook subscription, partner provisioning, or a hybrid?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

```text
Topic: Authorization model
Question: Are authorization_url, state, redirect_uri, and oauth_ready enough for StartProviderConnection, or does the RPC contract need another field?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

### Scopes And Activity Data

```text
Topic: Scopes and activity data
Question: Which permissions and activity fields are available for running activity import?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

```text
Topic: Scopes and activity data
Question: Are historical activities, activity edits, and activity deletes available after connection?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

### Activity Delivery

```text
Topic: Activity delivery
Question: Are activities delivered by webhook, polling, batch export, manual refresh, or another Garmin-specific mechanism?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

```text
Topic: Activity delivery
Question: What idempotency key, retry behavior, webhook signature, polling cursor, pagination, and backoff rules apply?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

### Sandbox And Test Data

```text
Topic: Sandbox and test data
Question: Is there a sandbox, test account, replay tool, staging callback setup, or representative activity data source?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

### Data Storage And Privacy

```text
Topic: Data storage and privacy
Question: May Runthread store raw Garmin activity payloads, token metadata, refresh tokens, provider account IDs, and import cursors?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

```text
Topic: Data storage and privacy
Question: What retention, deletion, export, redaction, encryption, support access, branding, or disclosure obligations apply?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

### Operations And Limits

```text
Topic: Operations and limits
Question: What rate limits, error codes, retry rules, outage behavior, monitoring, alerting, and audit requirements apply?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

### Disconnect And Revocation

```text
Topic: Disconnect and revocation
Question: How do user disconnect, Garmin-side revocation, reconnect, token deletion, payload deletion, and normalised activity retention work?
Answer:
Source/link:
Date checked:
Confidence:
Impact on backend:
Impact on mobile:
Follow-up owner:
Follow-up status: open
Notes:
```

## Decision Gate

Real Garmin implementation can start only after:

- access and approval path is confirmed
- authorization model is confirmed
- activity delivery mechanism is confirmed
- raw payload and token storage rules are confirmed
- sandbox or staging test strategy is confirmed
- disconnect and revocation behavior is confirmed
- any required proto, persistence, backend, or mobile boundary changes are documented
