package profile

import (
	"sync"
	"time"
)

// Profile represents a configuration profile for different environments
type Profile struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Environment string            `json:"environment" yaml:"environment"` // prod, staging, dev
	ParentID    string            `json:"parent_id,omitempty" yaml:"parent_id,omitempty"`
	Config      map[string]string `json:"config" yaml:"config"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	LastUsedAt  time.Time         `json:"last_used_at" yaml:"last_used_at"`
	Tags        []string          `json:"tags" yaml:"tags"`
	IsDefault   bool              `json:"is_default" yaml:"is_default"`
}

// ProfileManager manages multiple configuration profiles
type ProfileManager struct {
	mutex           sync.RWMutex
	profiles        map[string]*Profile
	activeProfile   *Profile
	configDir       string
	storage         ProfileStorage
	templateManager *TemplateManager
}

// ProfileStorage defines the interface for profile persistence
type ProfileStorage interface {
	Save(profile *Profile) error
	Load(id string) (*Profile, error)
	LoadAll() (map[string]*Profile, error)
	Delete(id string) error
	SetDefault(id string) error
	GetDefault() (*Profile, error)
}

// ProfileTemplate defines a template for creating new profiles
type ProfileTemplate struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	Environment string         `json:"environment" yaml:"environment"`
	ConfigKeys  []ConfigKeyDef `json:"config_keys" yaml:"config_keys"`
	Tags        []string       `json:"tags" yaml:"tags"`
}

// ConfigKeyDef defines a configuration key with validation rules
type ConfigKeyDef struct {
	Key         string `json:"key" yaml:"key"`
	Description string `json:"description" yaml:"description"`
	Required    bool   `json:"required" yaml:"required"`
	Default     string `json:"default" yaml:"default"`
	Type        string `json:"type" yaml:"type"`                                 // string, int, bool, url
	Validation  string `json:"validation,omitempty" yaml:"validation,omitempty"` // regex pattern
}

// ProfileCreateOptions contains options for creating a new profile
type ProfileCreateOptions struct {
	Name        string
	Description string
	Environment string
	Config      map[string]string
	ParentID    string
	Tags        []string
	SetDefault  bool
}

// ProfileUpdateOptions contains options for updating a profile
type ProfileUpdateOptions struct {
	Name        *string
	Description *string
	Environment *string
	Config      map[string]string
	Tags        []string
	SetDefault  *bool
}

// ProfileListOptions contains options for listing profiles
type ProfileListOptions struct {
	Environment string
	Tags        []string
	SortBy      string // name, created_at, last_used_at
	SortOrder   string // asc, desc
}

// Environment constants
const (
	EnvironmentProduction  = "production"
	EnvironmentStaging     = "staging"
	EnvironmentDevelopment = "development"
	EnvironmentTest        = "test"
)

// ConfigKey constants for common configuration keys
const (
	ConfigKeyAccessToken       = "SAKURACLOUD_ACCESS_TOKEN"       // #nosec G101 -- This is a configuration key name, not a credential
	ConfigKeyAccessTokenSecret = "SAKURACLOUD_ACCESS_TOKEN_SECRET" // #nosec G101 -- This is a configuration key name, not a credential
	ConfigKeyZone              = "SAKURACLOUD_ZONE"
	ConfigKeyDryRun            = "USACLOUD_UPDATE_DRY_RUN"
	ConfigKeyBatchMode         = "USACLOUD_UPDATE_BATCH"
	ConfigKeyInteractive       = "USACLOUD_UPDATE_INTERACTIVE"
)

// Default profile names
const (
	DefaultProfileName = "default"
)
