# Next Codex Prompt

Implement Stage 13: Activity matching to planned workouts for imported Strava activities.

Read the project docs first:

- `README.md`
- `docs/product.md`
- `docs/architecture.md`
- `docs/domain.md`
- `docs/strava.md`
- `docs/roadmap.md`
- `docs/decisions.md`
- `.codex/working-agreement.md`

Context:
Runthread now has Stage 9 mock Strava normalisation, Stage 10 backend-only Strava OAuth/token skeletons, Stage 11 Strava backfill/import skeletons, and Stage 12 Strava webhook handling skeletons. Strava provider code must remain isolated under provider packages, and core matching/adaptation must stay provider-neutral.

Goal:
Wire the existing provider-neutral matching flow so imported activities from the Strava provider import path can be matched to planned workouts for the current athlete.

Requirements:

- Use `domain.ImportedActivity` and existing provider-neutral matching concepts.
- Do not add Strava-specific matching rules unless the signal is first mapped to an existing provider-neutral field.
- Keep Strava-specific payloads and fields inside `services/api/internal/providers/strava`.
- Add tests for confident match, uncertain match, rejected match, and manual override or deferred manual-review state if currently supported.
- Reuse existing matching, provider import, repository, and app-service boundaries where appropriate.
- Update `docs/roadmap.md` with Stage 13 progress.
- Do not build Flutter UI.
- Do not call Strava APIs.
- Do not add AI integration.

Acceptance criteria:

- Imported Strava activities can flow into provider-neutral workout matching.
- Core domain, planning, matching, and adaptation packages do not depend on Strava types.
- Go tests pass.

Before coding:

- Make a short plan.
- List expected files to touch.

After coding:

- Summarise changes.
- Note any open questions.
- Write the next recommended Codex prompt to `.codex/next-prompt.md`.
