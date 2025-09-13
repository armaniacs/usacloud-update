package preview

import (
	"time"
)

// CommandPreview represents a detailed preview of a command transformation
type CommandPreview struct {
	Original    string            `json:"original"`
	Transformed string            `json:"transformed"`
	Changes     []ChangeHighlight `json:"changes"`
	Description string            `json:"description"`
	Impact      *ImpactAnalysis   `json:"impact"`
	Warnings    []string          `json:"warnings"`
	Category    string            `json:"category"`
	Metadata    *PreviewMetadata  `json:"metadata"`
}

// ChangeHighlight represents a highlighted change in the command
type ChangeHighlight struct {
	Type        ChangeType `json:"type"`
	Position    Range      `json:"position"`
	Original    string     `json:"original"`
	Replacement string     `json:"replacement"`
	Reason      string     `json:"reason"`
	RuleName    string     `json:"rule_name"`
}

// ChangeType represents the type of change made to a command
type ChangeType string

const (
	ChangeTypeOption   ChangeType = "option"
	ChangeTypeArgument ChangeType = "argument"
	ChangeTypeCommand  ChangeType = "command"
	ChangeTypeFormat   ChangeType = "format"
	ChangeTypeRemoval  ChangeType = "removal"
	ChangeTypeAddition ChangeType = "addition"
)

// Range represents a position range in the text
type Range struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// ImpactAnalysis represents the analysis of command impact
type ImpactAnalysis struct {
	Risk         RiskLevel `json:"risk"`
	Description  string    `json:"description"`
	Resources    []string  `json:"resources"`
	Dependencies []string  `json:"dependencies"`
	Complexity   int       `json:"complexity"`
}

// RiskLevel represents the risk level of executing a command
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// PreviewMetadata contains additional metadata about the preview
type PreviewMetadata struct {
	LineNumber     int           `json:"line_number"`
	GeneratedAt    time.Time     `json:"generated_at"`
	ProcessingTime time.Duration `json:"processing_time"`
	Version        string        `json:"version"`
}

// CommandCategory represents the category of a command
type CommandCategory string

const (
	CategoryServer     CommandCategory = "server"
	CategoryDatabase   CommandCategory = "database"
	CategoryNetwork    CommandCategory = "network"
	CategoryStorage    CommandCategory = "storage"
	CategorySecurity   CommandCategory = "security"
	CategoryMonitoring CommandCategory = "monitoring"
	CategoryOther      CommandCategory = "other"
)

// PreviewFilter represents filtering options for previews
type PreviewFilter struct {
	ShowOnlyChanged bool              `json:"show_only_changed"`
	Categories      []CommandCategory `json:"categories"`
	RiskLevels      []RiskLevel       `json:"risk_levels"`
	SearchQuery     string            `json:"search_query"`
}

// PreviewOptions represents options for generating previews
type PreviewOptions struct {
	IncludeDescription bool          `json:"include_description"`
	IncludeImpact      bool          `json:"include_impact"`
	IncludeWarnings    bool          `json:"include_warnings"`
	MaxDescLength      int           `json:"max_desc_length"`
	Timeout            time.Duration `json:"timeout"`
}
