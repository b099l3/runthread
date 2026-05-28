# Private Beta Readiness

Stage 16 is a preparation stage. Runthread is not private-beta ready until the remaining product, backend, mobile, provider, operations, and privacy gates below are resolved.

The beta path should stay Strava-first. Strava is the MVP provider for validating the activity import loop; Garmin direct integration remains blocked until `docs/garmin-access-findings.md` has validated external answers.

Do not implement payment processing, production auth, real provider API calls, or AI integration as part of readiness documentation. Strava-derived activity data must not be sent to AI prompts or used for model training.

## Target Beta Loop

```text
Invite-only athlete
  -> onboard and identify current athlete
  -> connect Strava
  -> import and normalise activities
  -> match imported activities to planned workouts
  -> create workout results
  -> adapt plan deterministically
  -> show plan, activity state, and adaptation explanation
```

## Readiness Checklist

Product and beta scope:

- Define the invite-only beta cohort, success criteria, feedback channel, and support response expectations.
- Decide which platforms are supported for beta testing and how builds are distributed.
- Keep the beta focused on plan viewing, Strava-backed activity import, matching, and deterministic adaptation.
- Explicitly exclude AI-generated training decisions and AI use of Strava-derived activity data.

Onboarding and athlete identity:

- Replace demo `athlete-1` / `goal-1` assumptions with auth-backed current athlete and current goal lookup.
- Decide whether beta accounts are invite codes, email/password, passkeys, Sign in with Apple, or another production auth path.
- Define how a newly onboarded athlete gets an initial plan and how current-plan state is selected.
- Decide whether seeded demo fallback remains available in beta builds or is restricted to local development only.

Provider connection readiness:

- Implement Strava connect start and callback through backend-owned OAuth when ready.
- Store provider tokens securely behind token references; mobile must not receive refresh tokens or call Strava directly.
- Add refresh handling, reconnect, disconnect, and revoked-access states before relying on imports for beta users.
- Keep Strava payloads, provider IDs, scopes, webhook shapes, and rate-limit details inside Strava provider packages and provider persistence.
- Keep Garmin disabled for beta until access findings are validated.

Activity import and adaptation:

- Run initial Strava backfill after connection and record import job status for retry and support.
- Handle webhook-triggered activity create/update/delete events after OAuth is production-ready.
- Use `ImportedActivity` as the only activity input to matching, workout result, and adaptation logic.
- Provide manual review or user-visible status for uncertain and rejected workout matches.
- Keep deterministic adaptation as the only source of plan changes.

Subscriptions and entitlements:

- Decide the beta entitlement model before adding payment code: free invite-only beta, platform subscription, Stripe-backed subscription, or another backend entitlement.
- Define the minimal entitlement record needed by the backend before the app depends on paid access.
- Decide how beta users transition to paid access after beta.
- Do not add real payment provider calls until the subscription model and app-store constraints are documented.

Operations and support:

- Add production-safe logging for provider connection, import job, webhook, matching, and adaptation failures.
- Define monitoring for API health, provider import backlog, rate-limit delays, token refresh failures, webhook failures, and adaptation errors.
- Add support diagnostics that expose provider connection state and import job status without exposing raw provider payloads unnecessarily.
- Define backup, migration, rollback, and incident response expectations for beta.

Privacy, deletion, and export:

- Write a beta privacy policy before collecting real provider data.
- Define retention for raw provider payloads, provider activity records, imported activities, workout results, and adaptation events.
- Implement user disconnect behavior for Strava: stop imports, remove token material, and clearly define whether normalised historical activities remain.
- Define account deletion and data export behavior before private beta.
- Ensure Strava data is shown only to the authorised Runthread user who connected that Strava account.
- Review and comply with Strava provider terms and rate limits before live imports.

Mobile readiness:

- Keep ConnectRPC transport isolated behind the mobile API boundary.
- Replace demo identifiers with current-user/current-plan reads once auth exists.
- Show provider connection, importing, delayed sync, disconnected, and reconnect-needed states once live Strava connection is available.
- Keep completion read-only from the mobile app until imported activities drive completion through the backend.
- Run mobile checks before beta builds: `asdf exec flutter analyze` and `asdf exec flutter test`.

Backend readiness:

- Validate provider persistence and token storage startup wiring for the selected production database path.
- Exercise the Postgres-backed current-plan read behavior against a repeatable beta database workflow.
- Add live database integration tests before depending on Postgres in beta.
- Keep provider import, matching, workout result, and adaptation services provider-neutral.
- Run backend checks before beta builds: `asdf exec go test ./...`.

## Beta Launch Blockers

- Production auth/current athlete identity is not implemented.
- Strava OAuth, token refresh, webhook endpoint exposure, and live fetch paths exist, but need production hardening and live provider validation.
- Provider token storage implementation exists, but production key management and operational policy are not final.
- Data deletion/export and retention behavior are not implemented.
- Subscriptions and entitlements are undecided.
- Monitoring, support diagnostics, and operational runbooks are not in place.
- Garmin direct integration is blocked by external access findings.

## Smallest Next Implementation Order

1. Verify current backend and mobile automated checks.
2. Decide beta auth/current-athlete approach.
3. Decide beta entitlement/subscription model without adding payment calls.
4. Harden Strava OAuth/token storage behind backend-only boundaries for beta operations.
5. Validate Strava backfill and webhook imports through provider-neutral import services with live provider data.
6. Add manual review/status handling for uncertain workout matches.
7. Add privacy, deletion, export, monitoring, and support surfaces before inviting real users.
