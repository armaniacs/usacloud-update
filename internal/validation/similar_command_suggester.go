// Package validation provides command validation functionality for usacloud-update
package validation

import (
	"sort"
	"strings"
)

// SimilarityResult represents a similarity search result
type SimilarityResult struct {
	Command  string  // Candidate command
	Distance int     // Levenshtein distance
	Score    float64 // Similarity score (0.0-1.0)
}

// SimilarCommandSuggester provides similar command suggestions
type SimilarCommandSuggester struct {
	allCommands        []string            // All available commands
	commandSubcommands map[string][]string // Command -> subcommand mapping
	maxDistance        int                 // Maximum allowable distance
	maxSuggestions     int                 // Maximum number of suggestions
}

// Common configuration constants
const (
	DefaultMaxDistance    = 3   // Maximum 3 character differences
	DefaultMaxSuggestions = 5   // Maximum 5 suggestions
	MinScore              = 0.5 // Minimum similarity score 50%
)

// CommonTypoPatterns maps common typo patterns
var CommonTypoPatterns = map[string][]string{
	"server":   {"sever", "serv", "srv", "servers"},
	"disk":     {"disc", "dsk", "disks"},
	"database": {"db", "databse", "datbase"},
	"list":     {"lst", "lis"},
	"create":   {"creat", "crate"},
	"delete":   {"delet", "del"},
	"switch":   {"swich", "swithc"},
	"note":     {"not", "notes"},
	"cdrom":    {"cd", "rom", "iso"},
	"archive":  {"arch", "archiv"},
	"snapshot": {"snap", "shot"},
}

// NewSimilarCommandSuggester creates a new command suggester
func NewSimilarCommandSuggester(maxDistance, maxSuggestions int) *SimilarCommandSuggester {
	return &SimilarCommandSuggester{
		allCommands:        getAllCommands(),
		commandSubcommands: getAllCommandSubcommands(),
		maxDistance:        maxDistance,
		maxSuggestions:     maxSuggestions,
	}
}

// NewDefaultSimilarCommandSuggester creates a suggester with default settings
func NewDefaultSimilarCommandSuggester() *SimilarCommandSuggester {
	return NewSimilarCommandSuggester(DefaultMaxDistance, DefaultMaxSuggestions)
}

// LevenshteinDistance calculates the Levenshtein distance between two strings
func (s *SimilarCommandSuggester) LevenshteinDistance(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 0
	}

	if len(s1) == 0 {
		return len(s2)
	}

	if len(s2) == 0 {
		return len(s1)
	}

	// Dynamic programming implementation
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize matrix
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Calculate distance
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			deletion := matrix[i-1][j] + 1
			insertion := matrix[i][j-1] + 1
			substitution := matrix[i-1][j-1] + cost

			matrix[i][j] = suggesterMin(deletion, suggesterMin(insertion, substitution))
		}
	}

	return matrix[len(s1)][len(s2)]
}

// SuggestMainCommands suggests main command candidates
func (s *SimilarCommandSuggester) SuggestMainCommands(input string) []SimilarityResult {
	if input == "" {
		return nil
	}

	var results []SimilarityResult
	maxDistance := s.getAdaptiveMaxDistance(input)

	// Filter candidates by prefix for performance
	candidates := s.filterByPrefix(input, s.allCommands)

	for _, command := range candidates {
		distance := s.LevenshteinDistance(input, command)

		if distance <= maxDistance {
			score := s.calculateScore(input, command, distance)
			if score >= MinScore {
				results = append(results, SimilarityResult{
					Command:  command,
					Distance: distance,
					Score:    score,
				})
			}
		}
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit to maximum suggestions
	if len(results) > s.maxSuggestions {
		results = results[:s.maxSuggestions]
	}

	return results
}

// SuggestSubcommands suggests subcommand candidates
func (s *SimilarCommandSuggester) SuggestSubcommands(mainCommand, input string) []SimilarityResult {
	if input == "" || mainCommand == "" {
		return nil
	}

	subcommands, exists := s.commandSubcommands[mainCommand]
	if !exists {
		return nil
	}

	var results []SimilarityResult
	maxDistance := s.getAdaptiveMaxDistance(input)

	// Filter candidates by prefix for performance
	candidates := s.filterByPrefix(input, subcommands)

	for _, subcommand := range candidates {
		distance := s.LevenshteinDistance(input, subcommand)

		if distance <= maxDistance {
			score := s.calculateScore(input, subcommand, distance)
			if score >= MinScore {
				results = append(results, SimilarityResult{
					Command:  subcommand,
					Distance: distance,
					Score:    score,
				})
			}
		}
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit to maximum suggestions
	if len(results) > s.maxSuggestions {
		results = results[:s.maxSuggestions]
	}

	return results
}

// getAdaptiveMaxDistance returns adaptive max distance based on input length
func (s *SimilarCommandSuggester) getAdaptiveMaxDistance(input string) int {
	length := len(input)
	switch {
	case length <= 3:
		return 1 // Strict for short strings
	case length <= 6:
		return 2 // Medium for medium strings
	default:
		return 3 // Lenient for long strings
	}
}

// calculateScore calculates similarity score with typo pattern bonus
func (s *SimilarCommandSuggester) calculateScore(input, candidate string, distance int) float64 {
	maxLength := suggesterMax(len(input), len(candidate))
	baseScore := 1.0 - float64(distance)/float64(maxLength)

	// Add typo pattern bonus
	typoBonus := s.getTypoScore(input, candidate)

	// Cap the final score at 1.0
	finalScore := baseScore + typoBonus
	if finalScore > 1.0 {
		finalScore = 1.0
	}

	return finalScore
}

// getTypoScore returns additional score for common typo patterns
func (s *SimilarCommandSuggester) getTypoScore(input, candidate string) float64 {
	patterns, exists := CommonTypoPatterns[candidate]
	if !exists {
		return 0.0
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range patterns {
		if inputLower == strings.ToLower(pattern) {
			return 0.2 // Typo pattern match bonus
		}
	}

	return 0.0
}

// filterByPrefix filters candidates by prefix for performance optimization
func (s *SimilarCommandSuggester) filterByPrefix(input string, candidates []string) []string {
	if len(input) < 2 {
		return candidates // Return all if input too short
	}

	prefix := strings.ToLower(input[:2])
	var filtered []string

	for _, candidate := range candidates {
		if strings.HasPrefix(strings.ToLower(candidate), prefix) {
			filtered = append(filtered, candidate)
		}
	}

	// Return all candidates if no prefix matches found
	if len(filtered) == 0 {
		return candidates
	}

	return filtered
}

// getAllCommands returns all available commands
func getAllCommands() []string {
	// Get commands from existing validators
	validator := NewMainCommandValidator()
	allCommands := make([]string, 0)

	// Add IaaS commands
	for command := range validator.iaasCommands {
		allCommands = append(allCommands, command)
	}

	// Add misc commands
	for command := range validator.miscCommands {
		allCommands = append(allCommands, command)
	}

	// Add root commands
	for command := range validator.rootCommands {
		allCommands = append(allCommands, command)
	}

	return allCommands
}

// getAllCommandSubcommands returns command to subcommand mapping
func getAllCommandSubcommands() map[string][]string {
	return map[string][]string{
		// IaaS commands with subcommands
		"server":                {"list", "read", "create", "update", "delete", "power-on", "power-off", "reset", "boot", "shutdown", "send-key", "vnc-info", "monitor", "activity-monitor"},
		"disk":                  {"list", "read", "create", "update", "delete", "connect", "disconnect"},
		"archive":               {"list", "read", "create", "update", "delete", "share", "open-ftp", "close-ftp"},
		"packet-filter":         {"list", "read", "create", "update", "delete", "add-rule", "clear-rules"},
		"internet":              {"list", "read", "create", "update", "delete", "monitor", "activity-monitor"},
		"switch":                {"list", "read", "create", "update", "delete", "connect", "disconnect"},
		"interface":             {"list", "read", "create", "update", "delete", "connect", "disconnect", "monitor", "activity-monitor"},
		"bridge":                {"list", "read", "create", "update", "delete"},
		"auto-backup":           {"list", "read", "create", "update", "delete"},
		"snapshot":              {"list", "read", "create", "update", "delete"},
		"load-balancer":         {"list", "read", "create", "update", "delete", "monitor", "activity-monitor"},
		"vpc-router":            {"list", "read", "create", "update", "delete", "boot", "shutdown", "reset", "monitor", "activity-monitor"},
		"database":              {"list", "read", "create", "update", "delete", "boot", "shutdown", "reset", "monitor", "activity-monitor"},
		"nfs":                   {"list", "read", "create", "update", "delete"},
		"mobile-gateway":        {"list", "read", "create", "update", "delete"},
		"sms":                   {"list", "send"},
		"cdrom":                 {"list", "read", "create", "update", "delete"},
		"note":                  {"list", "read", "create", "update", "delete"},
		"ipaddress":             {"list", "read", "create", "update", "delete"},
		"license":               {"list", "read", "info"},
		"bill":                  {"list", "read", "csv"},
		"coupon":                {"list", "read"},
		"authstatus":            {"read"},
		"zone":                  {"list", "read"},
		"region":                {"list", "read"},
		"icon":                  {"list", "read", "create", "update", "delete"},
		"private-host":          {"list", "read", "create", "update", "delete"},
		"dns":                   {"list", "read", "create", "update", "delete", "add-record", "clear-records"},
		"gslb":                  {"list", "read", "create", "update", "delete", "add-server", "clear-servers"},
		"simple-monitor":        {"list", "read", "create", "update", "delete", "activity-monitor"},
		"certificate-authority": {"list", "read", "create", "update", "delete"},
		"local-router":          {"list", "read", "create", "update", "delete", "health", "monitor", "activity-monitor"},
		"esme":                  {"list", "read", "send-message-status"},
		"proxylb":               {"list", "read", "create", "update", "delete", "add-bind-port", "update-bind-port", "delete-bind-port", "add-server", "update-server", "delete-server", "add-certificate", "update-certificate", "delete-certificate", "monitor", "activity-monitor"},
		"sim":                   {"list", "read", "create", "update", "delete", "activate", "deactivate", "assign-ip", "clear-ip", "set-imei-lock", "clear-imei-lock", "activity-monitor", "traffic-monitor"},
		"webaccel":              {"list", "read", "create", "update", "delete", "certificate"},
		"container-registry":    {"list", "read", "create", "update", "delete", "users", "add-user", "update-user", "delete-user"},
		"disk-plan":             {"list", "read"},
		"internet-plan":         {"list", "read"},
		"server-plan":           {"list", "read"},

		// Misc commands (most are standalone)
		"rest": {"get", "post", "put", "patch", "delete"},

		// Root commands (all standalone - no subcommands)
	}
}

// suggesterMin returns the minimum of two integers
func suggesterMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// suggesterMax returns the maximum of two integers
func suggesterMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
