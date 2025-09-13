// Package validation provides command validation functionality for usacloud-update
package validation

import (
	"fmt"
	"strings"
)

// DeprecationInfo represents deprecated command information
type DeprecationInfo struct {
	Command            string   // Deprecated command
	ReplacementCommand string   // Replacement command (empty if completely discontinued)
	DeprecationType    string   // "renamed" or "discontinued"
	Message            string   // Detailed explanation message
	AlternativeActions []string // Alternative methods (for discontinued commands)
	DocumentationURL   string   // Related documentation URL
}

// DeprecatedCommandDetector represents a deprecated command detector
type DeprecatedCommandDetector struct {
	deprecatedCommands map[string]*DeprecationInfo
}

// NewDeprecatedCommandDetector creates a new deprecated command detector
func NewDeprecatedCommandDetector() *DeprecatedCommandDetector {
	detector := &DeprecatedCommandDetector{
		deprecatedCommands: make(map[string]*DeprecationInfo),
	}

	// Initialize deprecated command information
	detector.initializeDeprecatedCommands()

	return detector
}

// initializeDeprecatedCommands initializes deprecated command information
func (d *DeprecatedCommandDetector) initializeDeprecatedCommands() {
	// Renamed commands
	d.deprecatedCommands["iso-image"] = &DeprecationInfo{
		Command:            "iso-image",
		ReplacementCommand: "cdrom",
		DeprecationType:    "renamed",
		Message:            "iso-imageコマンドはv1で廃止されました。cdromコマンドを使用してください。",
		DocumentationURL:   "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}

	d.deprecatedCommands["startup-script"] = &DeprecationInfo{
		Command:            "startup-script",
		ReplacementCommand: "note",
		DeprecationType:    "renamed",
		Message:            "startup-scriptコマンドはv1で廃止されました。noteコマンドを使用してください。",
		DocumentationURL:   "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}

	d.deprecatedCommands["ipv4"] = &DeprecationInfo{
		Command:            "ipv4",
		ReplacementCommand: "ipaddress",
		DeprecationType:    "renamed",
		Message:            "ipv4コマンドはv1で廃止されました。ipaddressコマンドを使用してください。",
		DocumentationURL:   "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}

	d.deprecatedCommands["product-disk"] = &DeprecationInfo{
		Command:            "product-disk",
		ReplacementCommand: "disk-plan",
		DeprecationType:    "renamed",
		Message:            "product-diskコマンドはv1で廃止されました。disk-planコマンドを使用してください。",
		DocumentationURL:   "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}

	d.deprecatedCommands["product-internet"] = &DeprecationInfo{
		Command:            "product-internet",
		ReplacementCommand: "internet-plan",
		DeprecationType:    "renamed",
		Message:            "product-internetコマンドはv1で廃止されました。internet-planコマンドを使用してください。",
		DocumentationURL:   "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}

	d.deprecatedCommands["product-server"] = &DeprecationInfo{
		Command:            "product-server",
		ReplacementCommand: "server-plan",
		DeprecationType:    "renamed",
		Message:            "product-serverコマンドはv1で廃止されました。server-planコマンドを使用してください。",
		DocumentationURL:   "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}

	// Completely discontinued commands
	d.deprecatedCommands["summary"] = &DeprecationInfo{
		Command:            "summary",
		ReplacementCommand: "",
		DeprecationType:    "discontinued",
		Message:            "summaryコマンドはv1で廃止されました。",
		AlternativeActions: []string{
			"請求情報は 'usacloud bill list' を使用してください",
			"アカウント情報は 'usacloud self read' を使用してください",
			"個別リソース情報は各リソースの 'list' コマンドを使用してください",
			"詳細な情報が必要な場合は 'usacloud rest' コマンドを使用してください",
		},
		DocumentationURL: "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}

	d.deprecatedCommands["object-storage"] = &DeprecationInfo{
		Command:            "object-storage",
		ReplacementCommand: "",
		DeprecationType:    "discontinued",
		Message:            "object-storageコマンドはv1で廃止されました。Sakura Cloudのオブジェクトストレージサービス終了に伴い利用できません。",
		AlternativeActions: []string{
			"S3互換ツール (aws-cli, s3cmd等) の使用を検討してください",
			"Terraformによるインフラ管理への移行を検討してください",
			"他のクラウドストレージサービスの利用を検討してください",
		},
		DocumentationURL: "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}

	d.deprecatedCommands["ojs"] = &DeprecationInfo{
		Command:            "ojs",
		ReplacementCommand: "",
		DeprecationType:    "discontinued",
		Message:            "ojsコマンド（object-storageのエイリアス）はv1で廃止されました。Sakura Cloudのオブジェクトストレージサービス終了に伴い利用できません。",
		AlternativeActions: []string{
			"S3互換ツール (aws-cli, s3cmd等) の使用を検討してください",
			"Terraformによるインフラ管理への移行を検討してください",
			"他のクラウドストレージサービスの利用を検討してください",
		},
		DocumentationURL: "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	}
}

// Detect detects deprecated commands
func (d *DeprecatedCommandDetector) Detect(command string) *DeprecationInfo {
	normalized := strings.ToLower(strings.TrimSpace(command))
	return d.deprecatedCommands[normalized]
}

// IsDeprecated checks if a command is deprecated
func (d *DeprecatedCommandDetector) IsDeprecated(command string) bool {
	normalized := strings.ToLower(strings.TrimSpace(command))
	return d.deprecatedCommands[normalized] != nil
}

// GetReplacementCommand returns the replacement command
func (d *DeprecatedCommandDetector) GetReplacementCommand(command string) string {
	if info := d.Detect(command); info != nil {
		return info.ReplacementCommand
	}
	return ""
}

// GetDeprecationType returns the deprecation type
func (d *DeprecatedCommandDetector) GetDeprecationType(command string) string {
	if info := d.Detect(command); info != nil {
		return info.DeprecationType
	}
	return ""
}

// GetDeprecationMessage returns the deprecation message
func (d *DeprecatedCommandDetector) GetDeprecationMessage(command string) string {
	if info := d.Detect(command); info != nil {
		return info.Message
	}
	return ""
}

// GetAlternativeActions returns alternative actions for discontinued commands
func (d *DeprecatedCommandDetector) GetAlternativeActions(command string) []string {
	if info := d.Detect(command); info != nil {
		return info.AlternativeActions
	}
	return []string{}
}

// GenerateMigrationMessage generates a comprehensive migration message
func (d *DeprecatedCommandDetector) GenerateMigrationMessage(command string) string {
	info := d.Detect(command)
	if info == nil {
		return ""
	}

	switch info.DeprecationType {
	case "renamed":
		return d.handleRenamedCommand(info)
	case "discontinued":
		return d.handleDiscontinuedCommand(info)
	default:
		return info.Message
	}
}

// handleRenamedCommand handles renamed command messages
func (d *DeprecatedCommandDetector) handleRenamedCommand(info *DeprecationInfo) string {
	return fmt.Sprintf(
		"%s は %s に名称変更されました。%s を使用してください。\n\n詳細: %s",
		info.Command,
		info.ReplacementCommand,
		info.ReplacementCommand,
		info.DocumentationURL,
	)
}

// handleDiscontinuedCommand handles discontinued command messages
func (d *DeprecatedCommandDetector) handleDiscontinuedCommand(info *DeprecationInfo) string {
	message := fmt.Sprintf("%s\n\n代替手段:\n", info.Message)
	for _, action := range info.AlternativeActions {
		message += fmt.Sprintf("  • %s\n", action)
	}
	message += fmt.Sprintf("\n詳細: %s", info.DocumentationURL)
	return message
}

// GetAllDeprecatedCommands returns all deprecated commands
func (d *DeprecatedCommandDetector) GetAllDeprecatedCommands() map[string]*DeprecationInfo {
	result := make(map[string]*DeprecationInfo)
	for cmd, info := range d.deprecatedCommands {
		// Return a copy to prevent modification
		result[cmd] = &DeprecationInfo{
			Command:            info.Command,
			ReplacementCommand: info.ReplacementCommand,
			DeprecationType:    info.DeprecationType,
			Message:            info.Message,
			AlternativeActions: make([]string, len(info.AlternativeActions)),
			DocumentationURL:   info.DocumentationURL,
		}
		copy(result[cmd].AlternativeActions, info.AlternativeActions)
	}
	return result
}

// GetRenamedCommands returns only renamed commands
func (d *DeprecatedCommandDetector) GetRenamedCommands() map[string]string {
	result := make(map[string]string)
	for cmd, info := range d.deprecatedCommands {
		if info.DeprecationType == "renamed" {
			result[cmd] = info.ReplacementCommand
		}
	}
	return result
}

// GetDiscontinuedCommands returns only discontinued commands
func (d *DeprecatedCommandDetector) GetDiscontinuedCommands() []string {
	var result []string
	for cmd, info := range d.deprecatedCommands {
		if info.DeprecationType == "discontinued" {
			result = append(result, cmd)
		}
	}
	return result
}

// GetDeprecatedCommandCount returns the total number of deprecated commands
func (d *DeprecatedCommandDetector) GetDeprecatedCommandCount() int {
	return len(d.deprecatedCommands)
}

// GetRenamedCommandCount returns the number of renamed commands
func (d *DeprecatedCommandDetector) GetRenamedCommandCount() int {
	count := 0
	for _, info := range d.deprecatedCommands {
		if info.DeprecationType == "renamed" {
			count++
		}
	}
	return count
}

// GetDiscontinuedCommandCount returns the number of discontinued commands
func (d *DeprecatedCommandDetector) GetDiscontinuedCommandCount() int {
	count := 0
	for _, info := range d.deprecatedCommands {
		if info.DeprecationType == "discontinued" {
			count++
		}
	}
	return count
}

// ValidateConsistencyWithDeprecatedCommands checks consistency with existing deprecated commands mapping
func (d *DeprecatedCommandDetector) ValidateConsistencyWithDeprecatedCommands() error {
	// Get the existing deprecated commands from PBI-006 implementation
	existingCommands := GetDeprecatedCommands()

	// Check if all existing deprecated commands are covered
	for existingCmd, existingReplacement := range existingCommands {
		info := d.Detect(existingCmd)
		if info == nil {
			return fmt.Errorf("deprecated command '%s' from existing mapping is not covered", existingCmd)
		}

		if info.ReplacementCommand != existingReplacement {
			return fmt.Errorf("replacement mismatch for '%s': existing='%s', detector='%s'",
				existingCmd, existingReplacement, info.ReplacementCommand)
		}
	}

	return nil
}
