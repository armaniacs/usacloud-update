// Package validation provides command validation functionality for usacloud-update
package validation

// DiskSubcommands contains the complete list of all disk subcommands
// organized by functionality categories for better management
var DiskSubcommands = []string{
	// Basic CRUD operations
	"list",   // List disks
	"read",   // Read disk details
	"create", // Create new disk
	"update", // Update disk configuration
	"delete", // Delete disk

	// Attachment operations (disk-specific)
	"connect",    // Connect disk to server
	"disconnect", // Disconnect disk from server

	// Management operations
	"clone",  // Clone disk (create copy)
	"resize", // Resize disk capacity
}

// DiskSubcommandDescriptions provides Japanese descriptions for each disk subcommand
var DiskSubcommandDescriptions = map[string]string{
	// Basic CRUD operations
	"list":   "ディスク一覧を表示",
	"read":   "ディスクの詳細情報を表示",
	"create": "新しいディスクを作成",
	"update": "ディスクの設定を更新",
	"delete": "ディスクを削除",

	// Attachment operations
	"connect":    "ディスクをサーバーに接続",
	"disconnect": "ディスクをサーバーから切断",

	// Management operations
	"clone":  "ディスクのクローンを作成",
	"resize": "ディスクサイズを変更",
}

// DiskSubcommandCategories provides categorization for each disk subcommand
var DiskSubcommandCategories = map[string]string{
	// Basic CRUD operations
	"list":   "basic",
	"read":   "basic",
	"create": "basic",
	"update": "basic",
	"delete": "basic",

	// Attachment operations
	"connect":    "attachment",
	"disconnect": "attachment",

	// Management operations
	"clone":  "management",
	"resize": "management",
}

// GetDiskSubcommands returns the complete list of disk subcommands
func GetDiskSubcommands() []string {
	return DiskSubcommands
}

// GetDiskSubcommandCount returns the total number of disk subcommands
func GetDiskSubcommandCount() int {
	return len(DiskSubcommands)
}

// IsValidDiskSubcommand checks if the given subcommand is valid for the disk command
func IsValidDiskSubcommand(subcommand string) bool {
	for _, sc := range DiskSubcommands {
		if sc == subcommand {
			return true
		}
	}
	return false
}

// GetDiskSubcommandDescription returns the Japanese description for a disk subcommand
func GetDiskSubcommandDescription(subcommand string) (string, bool) {
	description, exists := DiskSubcommandDescriptions[subcommand]
	return description, exists
}

// GetDiskSubcommandCategory returns the category for a disk subcommand
func GetDiskSubcommandCategory(subcommand string) (string, bool) {
	category, exists := DiskSubcommandCategories[subcommand]
	return category, exists
}

// GetBasicDiskSubcommands returns the basic CRUD operation subcommands for disk
func GetBasicDiskSubcommands() []string {
	return []string{"list", "read", "create", "update", "delete"}
}

// GetAttachmentDiskSubcommands returns the attachment operation subcommands for disk
func GetAttachmentDiskSubcommands() []string {
	return []string{"connect", "disconnect"}
}

// GetManagementDiskSubcommands returns the management operation subcommands for disk
func GetManagementDiskSubcommands() []string {
	return []string{"clone", "resize"}
}

// IsBasicDiskSubcommand checks if the subcommand is a basic CRUD operation
func IsBasicDiskSubcommand(subcommand string) bool {
	basicCommands := GetBasicDiskSubcommands()
	for _, cmd := range basicCommands {
		if cmd == subcommand {
			return true
		}
	}
	return false
}

// IsAttachmentDiskSubcommand checks if the subcommand is an attachment operation
func IsAttachmentDiskSubcommand(subcommand string) bool {
	attachmentCommands := GetAttachmentDiskSubcommands()
	for _, cmd := range attachmentCommands {
		if cmd == subcommand {
			return true
		}
	}
	return false
}

// IsManagementDiskSubcommand checks if the subcommand is a management operation
func IsManagementDiskSubcommand(subcommand string) bool {
	managementCommands := GetManagementDiskSubcommands()
	for _, cmd := range managementCommands {
		if cmd == subcommand {
			return true
		}
	}
	return false
}

// GetDiskSubcommandsByCategory returns subcommands filtered by category
func GetDiskSubcommandsByCategory(category string) []string {
	var result []string
	for _, subcommand := range DiskSubcommands {
		if cat, exists := DiskSubcommandCategories[subcommand]; exists && cat == category {
			result = append(result, subcommand)
		}
	}
	return result
}

// GetAllDiskCategories returns all available disk subcommand categories
func GetAllDiskCategories() []string {
	categoryMap := make(map[string]bool)
	for _, category := range DiskSubcommandCategories {
		categoryMap[category] = true
	}

	var categories []string
	for category := range categoryMap {
		categories = append(categories, category)
	}
	return categories
}
