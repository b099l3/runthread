# Next Codex Prompt

Implement Stage 16 preparation: subscriptions and private beta readiness plan.

Read the project docs first:

- `README.md`
- `docs/product.md`
- `docs/architecture.md`
- `docs/domain.md`
- `docs/strava.md`
- `docs/garmin.md`
- `docs/garmin-access-findings.md`
- `docs/roadmap.md`
- `docs/decisions.md`
- `.codex/working-agreement.md`

Context:
Runthread now has a provider-neutral Strava MVP import loop through mock normalisation, backend-only OAuth/token skeletons, backfill/import skeletons, webhook skeletons, matching, workout result creation, and deterministic adaptation. Garmin direct integration is aligned with the provider-neutral pipeline but remains blocked by external access findings.

Goal:
Prepare Stage 16 by documenting and lightly structuring private beta readiness without implementing payments, production auth, live provider calls, or AI integration.

Requirements:

- Review product, roadmap, provider, privacy, and current backend/mobile readiness.
- Update docs with a private beta readiness checklist covering subscriptions, onboarding, auth/current athlete identity, provider connection readiness, operational monitoring, support, privacy, data deletion/export, and provider terms.
- Keep the beta plan Strava-first for MVP validation and keep Garmin direct integration explicitly blocked until findings are validated.
- Do not implement real subscriptions or payment provider integration.
- Do not implement production auth.
- Do not call Strava or Garmin APIs.
- Do not build Flutter UI unless docs already require a tiny copy/state-only change.
- Do not add AI integration.
- Add small backend skeletons/tests only if they clarify beta readiness boundaries without starting production implementation.

Acceptance criteria:

- Docs clearly describe what remains before private beta.
- Subscriptions and private beta scope are concrete enough for the next implementation stage.
- Provider data privacy and no-AI boundaries remain explicit.
- Go tests pass if backend code changes.

Before coding:

- Make a short plan.
- List expected files to touch.

After coding:

- Summarise changes.
- Note any open questions.
- Write the next recommended Codex prompt to `.codex/next-prompt.md`.
