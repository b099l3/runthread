# Next Codex Prompt

Read README.md, docs/garmin.md, docs/garmin-access-findings.md, docs/decisions.md, docs/roadmap.md, docs/persistence.md, services/api/proto/runthread/v1/runthread.proto, services/api/internal/app/provider_connection.go, services/api/internal/app/provider_import.go, services/api/internal/providerimport, services/api/internal/repository/provider.go, services/api/internal/rpc/handler, and apps/mobile.

We are continuing Stage 8 preparation.

Do not implement real Garmin OAuth.
Do not add AI, auth, subscriptions, server startup Postgres provider wiring, live database integration tests, or Flutter Garmin OAuth screens yet.

Goal:
Use provided external Garmin validation findings to decide the smallest next implementation step.

Requirements:
- only proceed if the user provides actual Garmin findings or explicitly asks you to research them
- do not fill findings from assumptions
- update docs/garmin-access-findings.md with provided findings, including source/link, date checked, confidence, backend impact, mobile impact, and follow-up status
- review whether the findings require changes to protobuf contracts, provider persistence, provider import orchestration, server startup, or mobile UX
- propose the smallest next implementation step, but do not implement it unless the findings clearly unblock it and the user asks for implementation
- keep `Connect Garmin` disabled unless real authorization behavior is validated and implementation is explicitly requested
- do not add OAuth, callbacks, webhooks, polling jobs, or real sync unless the findings justify the path and the user explicitly asks for that implementation
- do not add or change protobuf methods
- do not call StartProviderConnection from Flutter
- run asdf exec flutter analyze
- run asdf exec flutter test
- run asdf exec go test ./... from services/api if backend docs/code change
- update docs/garmin.md and docs/roadmap.md
- give me the next prompt
