package repository

import (
	"errors"
	"fmt"
	"time"
)

type ProviderConnection struct {
	ID               string
	AthleteID        string
	Provider         string
	ProviderUserID   string
	Status           ProviderConnectionStatus
	ConnectedAt      time.Time
	DisconnectedAt   time.Time
	LastSyncAt       time.Time
	LastImportCursor string
	TokenReference   string
	TokenExpiresAt   time.Time
	LastError        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (c ProviderConnection) Validate() error {
	if c.ID == "" {
		return errors.New("provider connection id is required")
	}
	if c.AthleteID == "" {
		return errors.New("provider connection athlete id is required")
	}
	if c.Provider == "" {
		return errors.New("provider connection provider is required")
	}
	if !c.Status.Valid() {
		return fmt.Errorf("invalid provider connection status %q", c.Status)
	}
	return nil
}

type ProviderConnectionStatus string

const (
	ProviderConnectionStatusPending      ProviderConnectionStatus = "pending"
	ProviderConnectionStatusConnected    ProviderConnectionStatus = "connected"
	ProviderConnectionStatusSyncing      ProviderConnectionStatus = "syncing"
	ProviderConnectionStatusError        ProviderConnectionStatus = "error"
	ProviderConnectionStatusDisconnected ProviderConnectionStatus = "disconnected"
)

func (s ProviderConnectionStatus) Valid() bool {
	switch s {
	case ProviderConnectionStatusPending,
		ProviderConnectionStatusConnected,
		ProviderConnectionStatusSyncing,
		ProviderConnectionStatusError,
		ProviderConnectionStatusDisconnected:
		return true
	default:
		return false
	}
}

type ProviderActivity struct {
	ID                   string
	ProviderConnectionID string
	AthleteID            string
	ImportedActivityID   string
	Provider             string
	ProviderActivityID   string
	ProviderActivityType string
	StartedAt            time.Time
	Status               ProviderActivityStatus
	FirstSeenAt          time.Time
	LastSyncedAt         time.Time
	LastError            string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (a ProviderActivity) Validate() error {
	if a.ID == "" {
		return errors.New("provider activity id is required")
	}
	if a.ProviderConnectionID == "" {
		return errors.New("provider activity connection id is required")
	}
	if a.AthleteID == "" {
		return errors.New("provider activity athlete id is required")
	}
	if a.Provider == "" {
		return errors.New("provider activity provider is required")
	}
	if a.ProviderActivityID == "" {
		return errors.New("provider activity provider activity id is required")
	}
	if !a.Status.Valid() {
		return fmt.Errorf("invalid provider activity status %q", a.Status)
	}
	return nil
}

type ProviderActivityStatus string

const (
	ProviderActivityStatusReceived   ProviderActivityStatus = "received"
	ProviderActivityStatusNormalised ProviderActivityStatus = "normalised"
	ProviderActivityStatusIgnored    ProviderActivityStatus = "ignored"
	ProviderActivityStatusFailed     ProviderActivityStatus = "failed"
	ProviderActivityStatusDeleted    ProviderActivityStatus = "deleted"
)

func (s ProviderActivityStatus) Valid() bool {
	switch s {
	case ProviderActivityStatusReceived,
		ProviderActivityStatusNormalised,
		ProviderActivityStatusIgnored,
		ProviderActivityStatusFailed,
		ProviderActivityStatusDeleted:
		return true
	default:
		return false
	}
}

type ProviderActivityPayload struct {
	ID                 string
	ProviderActivityID string
	Payload            []byte
	PayloadKind        string
	ReceivedAt         time.Time
}

func (p ProviderActivityPayload) Validate() error {
	if p.ID == "" {
		return errors.New("provider activity payload id is required")
	}
	if p.ProviderActivityID == "" {
		return errors.New("provider activity payload activity id is required")
	}
	if len(p.Payload) == 0 {
		return errors.New("provider activity payload is required")
	}
	if p.PayloadKind == "" {
		return errors.New("provider activity payload kind is required")
	}
	return nil
}

type ProviderImportEvent struct {
	ID                   string
	ProviderConnectionID string
	ProviderActivityID   string
	Provider             string
	EventType            string
	DeliveryID           string
	Status               ProviderImportEventStatus
	ReceivedAt           time.Time
	ProcessedAt          time.Time
	Error                string
}

func (e ProviderImportEvent) Validate() error {
	if e.ID == "" {
		return errors.New("provider import event id is required")
	}
	if e.Provider == "" {
		return errors.New("provider import event provider is required")
	}
	if e.EventType == "" {
		return errors.New("provider import event type is required")
	}
	if !e.Status.Valid() {
		return fmt.Errorf("invalid provider import event status %q", e.Status)
	}
	return nil
}

type ProviderImportEventStatus string

const (
	ProviderImportEventStatusReceived  ProviderImportEventStatus = "received"
	ProviderImportEventStatusProcessed ProviderImportEventStatus = "processed"
	ProviderImportEventStatusIgnored   ProviderImportEventStatus = "ignored"
	ProviderImportEventStatusFailed    ProviderImportEventStatus = "failed"
)

func (s ProviderImportEventStatus) Valid() bool {
	switch s {
	case ProviderImportEventStatusReceived,
		ProviderImportEventStatusProcessed,
		ProviderImportEventStatusIgnored,
		ProviderImportEventStatusFailed:
		return true
	default:
		return false
	}
}
