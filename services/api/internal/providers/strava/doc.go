// Package strava will contain Strava-specific activity provider code.
//
// Keep Strava OAuth, webhook payloads, API responses, and activity mapping in
// this package or adjacent provider integration packages. Core domain,
// planning, matching, and adaptation code should only receive normalised
// domain.ImportedActivity values.
package strava

const ProviderName = "strava"
