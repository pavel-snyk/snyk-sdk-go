package snyk

import "time"

const (
	tenantsBasePath = "tenants"
)

// Tenant represents a Snyk Tenant.
//
// See: https://docs.snyk.io/discover-snyk/getting-started/glossary#tenant
type Tenant struct {
	ID         string            `json:"id"`                   // The Tenant identifier.
	Type       string            `json:"type,omitempty"`       // The resource type `tenant`.
	Attributes *TenantAttributes `json:"attributes,omitempty"` // The Tenant resource data.
}

type TenantAttributes struct {
	CreatedAt time.Time `json:"created_at"` // The time the tenant was created.
	Name      string    `json:"name"`       // The display name of the tenant.
	Slug      string    `json:"slug"`       // The canonical (unique and URL-friendly) name of the tenant.
	UpdatedAt time.Time `json:"updated_at"` // The time the tenant was last modified.
}

type tenantRoot struct {
	Data *Tenant `json:"data,omitempty"`
}
