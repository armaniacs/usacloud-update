// Package validation provides command validation functionality for usacloud-update
package validation

// IaaSCommands contains the complete dictionary of 48 IaaS-related usacloud commands
// and their supported subcommands. This dictionary serves as the foundation for
// command validation system.
var IaaSCommands = map[string][]string{
	// Storage & Archive commands
	"archive": {"list", "read", "create", "update", "delete", "download", "extract"},
	"cdrom":   {"list", "read", "create", "update", "delete", "upload"},
	"disk":    {"list", "read", "create", "update", "delete", "connect", "disconnect", "resize"},

	// Server & Compute commands
	"server":          {"list", "read", "create", "update", "delete", "boot", "shutdown", "reset", "send-key", "rdp", "ssh"},
	"serverplan":      {"list", "read"},
	"privatehost":     {"list", "read", "create", "update", "delete"},
	"privatehostplan": {"list", "read"},

	// Network commands
	"swytch":        {"list", "read", "create", "update", "delete", "connect", "disconnect"},
	"router":        {"list", "read", "create", "update", "delete", "boot", "shutdown"},
	"bridge":        {"list", "read", "create", "update", "delete"},
	"internet":      {"list", "read", "create", "update", "delete", "connect", "disconnect"},
	"internetplan":  {"list", "read"},
	"subnet":        {"list", "read", "create", "update", "delete"},
	"ipaddress":     {"list", "read", "update"},
	"ipv6addr":      {"list", "read", "create", "delete"},
	"ipv6net":       {"list", "read", "create", "update", "delete"},
	"iface":         {"list", "read", "create", "update", "delete", "connect", "disconnect"},
	"packetfilter":  {"list", "read", "create", "update", "delete"},
	"vpcrouter":     {"list", "read", "create", "update", "delete", "boot", "shutdown", "dhcp-server", "dhcp-static-mapping", "firewall", "l2tp", "port-forwarding", "pptp", "site-to-site-vpn", "static-nat", "static-route", "user"},
	"localrouter":   {"list", "read", "create", "update", "delete", "boot", "shutdown", "health", "peer", "static-route"},
	"mobilegateway": {"list", "read", "create", "update", "delete", "boot", "shutdown", "dns", "firewall", "interface", "sim-route", "static-route", "traffic-control"},

	// Load Balancer & Proxy commands
	"loadbalancer": {"list", "read", "create", "update", "delete", "boot", "shutdown", "monitor"},
	"proxylb":      {"list", "read", "create", "update", "delete", "bind-certificate", "unbind-certificate", "health"},
	"gslb":         {"list", "read", "create", "update", "delete", "health"},

	// DNS commands
	"dns": {"list", "read", "create", "update", "delete", "record"},

	// Database commands
	"database":   {"list", "read", "create", "update", "delete", "boot", "shutdown", "log", "monitor"},
	"enhanceddb": {"list", "read", "create", "update", "delete", "log"},

	// File System commands
	"nfs": {"list", "read", "create", "update", "delete", "plan"},

	// Container commands
	"containerregistry": {"list", "read", "create", "update", "delete", "tag", "users"},

	// Monitoring & Backup commands
	"autobackup":    {"list", "read", "create", "update", "delete"},
	"simplemonitor": {"list", "read", "create", "update", "delete", "health"},
	"autoscale":     {"list", "read", "create", "update", "delete", "status", "cpu-time", "scaling-history"},

	// Security & Authentication commands
	"sshkey":               {"list", "read", "create", "update", "delete", "generate"},
	"license":              {"list", "read"},
	"licenseinfo":          {"list", "read"},
	"certificateauthority": {"list", "read", "create", "update", "delete", "issue", "renew"},

	// Communication commands
	"esme": {"list", "read", "create", "update", "delete", "logs", "send"},
	"sim":  {"list", "read", "create", "update", "delete", "activate", "deactivate", "assign", "clear-ip", "info", "logs", "set-imei-lock", "clear-imei-lock"},

	// Management & Administration commands
	"icon":       {"list", "read", "create", "update", "delete", "upload"},
	"note":       {"list", "read", "create", "update", "delete"},
	"bill":       {"list", "read", "detail"},
	"coupon":     {"list", "read"},
	"authstatus": {"read"},
	"self":       {"read", "update"},

	// Plan & Service Class commands
	"diskplan":     {"list", "read"},
	"serviceclass": {"list", "read"},
	"category":     {"list", "read"},

	// Geographic & Zone commands
	"zone":   {"list", "read"},
	"region": {"list", "read"},
}

// GetIaaSCommandSubcommands returns the list of valid subcommands for the given IaaS command
func GetIaaSCommandSubcommands(command string) ([]string, bool) {
	subcommands, exists := IaaSCommands[command]
	return subcommands, exists
}

// IsValidIaaSCommand checks if the given command is a valid IaaS command
func IsValidIaaSCommand(command string) bool {
	_, exists := IaaSCommands[command]
	return exists
}

// IsValidIaaSSubcommand checks if the given subcommand is valid for the given IaaS command
func IsValidIaaSSubcommand(command, subcommand string) bool {
	subcommands, exists := IaaSCommands[command]
	if !exists {
		return false
	}

	for _, sc := range subcommands {
		if sc == subcommand {
			return true
		}
	}
	return false
}

// GetAllIaaSCommands returns a slice of all available IaaS commands
func GetAllIaaSCommands() []string {
	commands := make([]string, 0, len(IaaSCommands))
	for command := range IaaSCommands {
		commands = append(commands, command)
	}
	return commands
}

// GetTotalCommandCount returns the total number of defined IaaS commands
func GetTotalCommandCount() int {
	return len(IaaSCommands)
}
