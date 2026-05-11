// Package garmin will contain Garmin-specific activity provider code for the
// provider-neutral import pipeline.
//
// Existing mock Garmin code currently lives in services/api/internal/garmin.
// Future direct Garmin work should keep Garmin payloads and provider rules
// isolated before normalising activities into domain.ImportedActivity.
package garmin

const ProviderName = "garmin"
