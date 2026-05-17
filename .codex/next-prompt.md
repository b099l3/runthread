# Next Codex Prompt

Implement the next Stage 16 step: decide and document the beta auth/current-athlete boundary, then add the smallest backend skeleton only if it clarifies that boundary.

Read the project docs first:

- `README.md`
- `docs/product.md`
- `docs/architecture.md`
- `docs/domain.md`
- `docs/strava.md`
- `docs/garmin.md`
- `docs/garmin-access-findings.md`
- `docs/private-beta.md`
- `docs/roadmap.md`
- `docs/decisions.md`
- `.codex/working-agreement.md`

Context:
Runthread has a provider-neutral Strava MVP import loop through mock provider code, backend-only OAuth/token skeletons, backfill/import skeletons, webhook skeletons, matching, workout result creation, deterministic adaptation, and Garmin alignment. Stage 16 now has a private beta readiness checklist. The app still uses demo `athlete-1` / `goal-1` identifiers and local demo fallback. VS Code launch configs have been checked; Flutter debug launch now has separate local/iOS and Android emulator backend URL configs.

Goal:
Prepare the private beta auth/current-athlete boundary so later provider connection and plan reads can stop depending on demo identifiers.

Requirements:

- Document the chosen near-term beta identity approach and how the backend should resolve the current athlete, current goal, and current plan.
- Keep this as a boundary/skeleton stage; do not implement full production auth unless explicitly requested.
- Do not implement subscriptions or payment provider calls.
- Do not call Strava or Garmin APIs.
- Do not enable Garmin.
- Do not add AI integration or send provider activity data to AI systems.
- If adding backend code, keep it small, provider-neutral, and testable.
- Preserve existing demo/local development behavior unless the docs explicitly say to change it.
- Run current automated checks after changes:
  - `cd services/api && asdf exec go test ./...`
  - `cd apps/mobile && asdf exec flutter analyze`
  - `cd apps/mobile && asdf exec flutter test`

Acceptance criteria:

- Docs clearly describe the beta auth/current-athlete boundary and remaining implementation work.
- Demo/local development behavior remains clear.
- Core domain and provider packages remain provider-neutral.
- No real auth, subscription, provider API, Garmin, or AI implementation is added.
- Go and Flutter checks pass, or failures are documented with exact next steps.

Before coding:

- Make a short plan.
- List expected files to touch.

After coding:

- Summarise changes.
- Note any open questions.
- Write the next recommended Codex prompt to `.codex/next-prompt.md`.
