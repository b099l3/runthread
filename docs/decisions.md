# Architecture Decisions

This file records important architecture decisions for Runthread. Add new decisions when the direction changes or when a meaningful tradeoff is resolved.

## ADR-0001: Use Go Backend

Status: Accepted

Decision: Runthread will use Go for the backend service and target Go 1.22 or newer.

Reason: Go is simple to deploy, works well for API services, has strong typing, and fits deterministic domain logic and integration workflows.

## ADR-0002: Use ConnectRPC

Status: Accepted

Decision: The Flutter app will communicate with the Go backend through ConnectRPC.

Reason: ConnectRPC gives a typed service contract while keeping HTTP-friendly operational behavior.

## ADR-0003: Use Postgres

Status: Accepted

Decision: Runthread will use Postgres as the primary application database.

Reason: The product needs reliable relational data for athletes, plans, workouts, activities, matches, and adaptation history.

## ADR-0004: Keep AI Out of Core Training Decisions

Status: Accepted

Decision: AI may explain decisions or help with copy, but deterministic code owns training plans and adaptations.

Reason: Training changes must be explainable, testable, auditable, and consistent.

## ADR-0005: Garmin First, Coros and Apple Watch Later

Status: Accepted

Decision: Garmin is the first provider integration. Coros and Apple Watch are deferred.

Reason: Focusing on one provider reduces integration complexity and matches the initial target user. The domain should still remain provider-neutral.

Note: Real Garmin access is assumed to require validation or application before production integration work begins.

## ADR-0006: Use asdf for Tool Versions

Status: Accepted

Decision: Runthread will pin local development tool versions with repo-level `.tool-versions`.

Reason: Future sessions and local development should use the same toolchain without relying on machine-global defaults.

Initial pinned tool:

- Go `1.25.5`
