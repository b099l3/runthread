# Next Codex Prompt

Implement Stage 14: Adaptation from imported Strava activities.

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
Runthread now has mock Strava normalisation, backend-only OAuth/token skeletons, backfill/import skeletons, webhook handling skeletons, and provider-neutral matching for imported activities. Stage 13 added `ProviderActivityMatchService`, which persists `WorkoutMatch` records for confident, uncertain, rejected, and manual matches without depending on Strava types.

Goal:
Wire imported Strava activities through provider-neutral workout result creation and deterministic adaptation after a confident match.

Requirements:

- Use existing provider-neutral domain types: `ImportedActivity`, `WorkoutMatch`, `WorkoutResult`, and `AdaptationEvent`.
- Reuse existing workout completion and adaptation engine behavior where possible.
- Do not add Strava-specific adaptation rules.
- Do not send Strava-derived activity data to AI prompts or use it for model training.
- Keep Strava-specific payloads and fields inside `services/api/internal/providers/strava`.
- Add tests for completed-as-planned, underperformed or partial completion, overperformed if currently supported, and no-adaptation for rejected/uncertain matches if that is the current boundary.
- Update `docs/roadmap.md` with Stage 14 progress.
- Do not build Flutter UI.
- Do not call Strava APIs.
- Do not add AI integration.

Acceptance criteria:

- A confidently matched imported Strava activity can produce a `WorkoutResult`.
- Deterministic adaptation can run from that result through provider-neutral code.
- Core domain, planning, matching, and adaptation packages do not depend on Strava types.
- Go tests pass.

Before coding:

- Make a short plan.
- List expected files to touch.

After coding:

- Summarise changes.
- Note any open questions.
- Write the next recommended Codex prompt to `.codex/next-prompt.md`.
