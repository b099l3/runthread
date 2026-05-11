# Next Codex Prompt

Implement Stage 15 preparation: Garmin direct integration alignment with the provider-neutral pipeline.

Read the project docs first:

- `README.md`
- `docs/product.md`
- `docs/architecture.md`
- `docs/domain.md`
- `docs/garmin.md`
- `docs/garmin-access-findings.md`
- `docs/strava.md`
- `docs/roadmap.md`
- `docs/decisions.md`
- `.codex/working-agreement.md`

Context:
Runthread now has a provider-neutral Strava MVP import path through mock normalisation, OAuth/token skeletons, backfill/import skeletons, webhook skeletons, matching, workout result creation, and deterministic adaptation. Garmin direct integration remains on the roadmap but real Garmin access is still blocked until access findings are validated.

Goal:
Prepare Stage 15 by aligning Garmin direct integration docs and mock boundaries with the provider-neutral pipeline proven through Strava, without implementing real Garmin OAuth or API calls.

Requirements:

- Review existing Garmin mock packages and docs against the provider-neutral Strava pipeline.
- Identify any Garmin mock code that should eventually move under `services/api/internal/providers/garmin` or be bridged more clearly.
- Keep Garmin-specific payloads and fields out of core domain, planning, matching, and adaptation packages.
- Do not implement real Garmin OAuth, callbacks, polling, webhooks, token refresh, or API calls.
- Do not enable Flutter Garmin connect actions.
- Update docs/garmin.md and docs/roadmap.md with the smallest Stage 15 preparation state.
- Add or adjust small backend skeletons/tests only if they clarify Garmin's reuse of the provider-neutral pipeline without real provider calls.
- Do not add AI integration.

Acceptance criteria:

- Garmin direct integration remains clearly blocked by access findings.
- Docs explain how Garmin will reuse the provider-neutral import, matching, result, and adaptation pipeline.
- Any code changes are mock-only and keep core packages provider-neutral.
- Go tests pass if backend code changes.

Before coding:

- Make a short plan.
- List expected files to touch.

After coding:

- Summarise changes.
- Note any open questions.
- Write the next recommended Codex prompt to `.codex/next-prompt.md`.
