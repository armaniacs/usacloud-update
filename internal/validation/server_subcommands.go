// Package validation provides command validation functionality for usacloud-update
package validation

// ServerSubcommands contains the complete list of all 15 server subcommands
// organized by functionality categories for better management
var ServerSubcommands = []string{
	// Basic CRUD operations
	"list",   // List servers
	"read",   // Read server details
	"create", // Create new server
	"update", // Update server configuration
	"delete", // Delete server

	// Power control operations
	"boot",     // Boot/start server
	"shutdown", // Shutdown server
	"reset",    // Reset server (force restart)

	// Management operations
	"send-nmi", // Send Non-Maskable Interrupt

	// Monitoring operations
	"monitor-cpu", // Monitor CPU usage

	// Connection operations
	"ssh", // SSH connection
	"vnc", // VNC connection
	"rdp", // RDP connection

	// Wait operations
	"wait-until-ready",    // Wait until server is ready
	"wait-until-shutdown", // Wait until server is shutdown
}

// ServerSubcommandDescriptions provides Japanese descriptions for each server subcommand
var ServerSubcommandDescriptions = map[string]string{
	// Basic CRUD operations
	"list":   "サーバー一覧を表示",
	"read":   "サーバーの詳細情報を表示",
	"create": "新しいサーバーを作成",
	"update": "サーバーの設定を更新",
	"delete": "サーバーを削除",

	// Power control operations
	"boot":     "サーバーを起動",
	"shutdown": "サーバーをシャットダウン",
	"reset":    "サーバーをリセット（強制再起動）",

	// Management operations
	"send-nmi": "サーバーにNMI（Non-Maskable Interrupt）を送信",

	// Monitoring operations
	"monitor-cpu": "サーバーのCPU使用率を監視",

	// Connection operations
	"ssh": "サーバーにSSH接続",
	"vnc": "サーバーにVNC接続",
	"rdp": "サーバーにRDP接続",

	// Wait operations
	"wait-until-ready":    "サーバーの起動完了まで待機",
	"wait-until-shutdown": "サーバーのシャットダウン完了まで待機",
}

// GetServerSubcommands returns the complete list of server subcommands
func GetServerSubcommands() []string {
	return ServerSubcommands
}

// GetServerSubcommandCount returns the total number of server subcommands
func GetServerSubcommandCount() int {
	return len(ServerSubcommands)
}

// IsValidServerSubcommand checks if the given subcommand is valid for the server command
func IsValidServerSubcommand(subcommand string) bool {
	for _, sc := range ServerSubcommands {
		if sc == subcommand {
			return true
		}
	}
	return false
}

// GetServerSubcommandDescription returns the Japanese description for a server subcommand
func GetServerSubcommandDescription(subcommand string) (string, bool) {
	description, exists := ServerSubcommandDescriptions[subcommand]
	return description, exists
}

// GetBasicCRUDSubcommands returns the basic CRUD operation subcommands for server
func GetBasicCRUDSubcommands() []string {
	return []string{"list", "read", "create", "update", "delete"}
}

// GetPowerControlSubcommands returns the power control subcommands for server
func GetPowerControlSubcommands() []string {
	return []string{"boot", "shutdown", "reset"}
}

// GetConnectionSubcommands returns the connection subcommands for server
func GetConnectionSubcommands() []string {
	return []string{"ssh", "vnc", "rdp"}
}

// GetWaitSubcommands returns the wait operation subcommands for server
func GetWaitSubcommands() []string {
	return []string{"wait-until-ready", "wait-until-shutdown"}
}

// IsBasicCRUDSubcommand checks if the subcommand is a basic CRUD operation
func IsBasicCRUDSubcommand(subcommand string) bool {
	crudCommands := GetBasicCRUDSubcommands()
	for _, cmd := range crudCommands {
		if cmd == subcommand {
			return true
		}
	}
	return false
}

// IsPowerControlSubcommand checks if the subcommand is a power control operation
func IsPowerControlSubcommand(subcommand string) bool {
	powerCommands := GetPowerControlSubcommands()
	for _, cmd := range powerCommands {
		if cmd == subcommand {
			return true
		}
	}
	return false
}

// IsConnectionSubcommand checks if the subcommand is a connection operation
func IsConnectionSubcommand(subcommand string) bool {
	connectionCommands := GetConnectionSubcommands()
	for _, cmd := range connectionCommands {
		if cmd == subcommand {
			return true
		}
	}
	return false
}

// IsWaitSubcommand checks if the subcommand is a wait operation
func IsWaitSubcommand(subcommand string) bool {
	waitCommands := GetWaitSubcommands()
	for _, cmd := range waitCommands {
		if cmd == subcommand {
			return true
		}
	}
	return false
}

// IsManagementSubcommand checks if the subcommand is a management operation
func IsManagementSubcommand(subcommand string) bool {
	return subcommand == "send-nmi"
}

// IsMonitoringSubcommand checks if the subcommand is a monitoring operation
func IsMonitoringSubcommand(subcommand string) bool {
	return subcommand == "monitor-cpu"
}
