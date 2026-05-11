# Codex Working Agreement

Future Codex sessions should work this way:

- Read `README.md` and the relevant files in `docs/` before coding.
- Use the repo `.tool-versions` through `asdf` before running language tooling.
- Make a short plan before changing code.
- Keep changes small and tied to the current stage.
- Update docs when decisions, domain concepts, or architecture change.
- Do not build unrelated future features.
- Prefer simple, deterministic, testable domain logic.
- When adding provider integrations, start with docs and mock data before real OAuth, webhooks, polling, or API calls.
- Avoid provider-specific leakage into domain, planning, matching, and adaptation packages.
- Keep OAuth code exchange, token storage, token refresh, webhook handling, and provider API calls backend-only.
- Keep provider integrations small and testable, with provider-specific parsing isolated in provider packages.
- Respect provider API terms, rate limits, privacy rules, retention rules, and user data access boundaries.
- Keep AI out of core training decisions.
- Do not send Strava-derived activity data to AI prompts or use it for model training.
- Assume real Garmin access must be validated or applied for before direct Garmin implementation.
- Ask for clarification only when blocked.
- Leave clear TODOs for later stages when useful.
- Verify changes with focused tests or compile checks before finishing.
- At the end of each implementation session, write the next recommended Codex prompt to `.codex/next-prompt.md` and mention that it was updated in the final response.
