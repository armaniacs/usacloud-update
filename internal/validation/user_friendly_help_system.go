// Package validation provides command validation functionality for usacloud-update
package validation

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// HelpContext represents help request context
type HelpContext struct {
	RequestedCommand string         // Requested command
	PreviousErrors   []ErrorHistory // Previous error history
	UserSkillLevel   SkillLevel     // User skill level
	PreferredFormat  HelpFormat     // Preferred help format
	LastAccessed     time.Time      // Last access time
}

// ErrorHistory represents error history entry
type ErrorHistory struct {
	Timestamp   time.Time // Error timestamp
	Command     string    // Input command
	ErrorType   string    // Error type
	WasResolved bool      // Whether resolved
	Resolution  string    // Resolution method
}

// SkillLevel represents user skill level
type SkillLevel int

const (
	SkillBeginner     SkillLevel = iota // Beginner
	SkillIntermediate                   // Intermediate
	SkillAdvanced                       // Advanced
	SkillExpert                         // Expert
)

// HelpFormat represents help display format
type HelpFormat int

const (
	FormatBasic       HelpFormat = iota // Basic format
	FormatDetailed                      // Detailed format
	FormatInteractive                   // Interactive format
	FormatExample                       // Example-focused format
)

// BuilderStep represents interactive builder step
type BuilderStep int

const (
	StepMainCommand BuilderStep = iota // Main command selection
	StepSubCommand                     // Subcommand selection
	StepOptions                        // Options configuration
	StepConfirm                        // Confirmation
)

// CommonMistake represents a common user mistake
type CommonMistake struct {
	Pattern         string   // Common mistake pattern
	Description     string   // Mistake description
	CorrectExamples []string // Correct examples
	Explanation     string   // Detailed explanation
	RelatedTopics   []string // Related topics
	Frequency       int      // Occurrence frequency
}

// TutorialStep represents a tutorial step
type TutorialStep struct {
	StepID      string   // Step ID
	Title       string   // Step title
	Description string   // Step description
	Commands    []string // Commands to try
	Tips        []string // Tips for this step
}

// ConceptExplanation represents a concept explanation
type ConceptExplanation struct {
	ConceptID   string   // Concept ID
	Title       string   // Concept title
	Description string   // Detailed description
	Examples    []string // Usage examples
	SeeAlso     []string // Related concepts
}

// MigrationGuide represents a migration guide
type MigrationGuide struct {
	FromVersion string             // Source version
	ToVersion   string             // Target version
	Changes     []MigrationChange  // List of changes
	Examples    []MigrationExample // Migration examples
}

// MigrationChange represents a single migration change
type MigrationChange struct {
	OldCommand string // Old command format
	NewCommand string // New command format
	Reason     string // Reason for change
	Impact     string // Impact assessment
}

// MigrationExample represents a migration example
type MigrationExample struct {
	Scenario    string // Usage scenario
	OldCommand  string // Old command
	NewCommand  string // New command
	Explanation string // Explanation
}

// CompletedTask represents a completed task
type CompletedTask struct {
	TaskID     string    // Task ID
	Command    string    // Executed command
	Timestamp  time.Time // Completion time
	Difficulty int       // Difficulty (1-10)
	Success    bool      // Success status
}

// LearningGoal represents a learning goal
type LearningGoal struct {
	GoalID      string     // Goal ID
	Title       string     // Goal title
	Description string     // Detailed description
	Steps       []string   // Achievement steps
	Progress    float64    // Progress rate (0-1)
	Deadline    *time.Time // Deadline
}

// Recommendation represents a learning recommendation
type Recommendation struct {
	Type        string // Recommendation type
	Title       string // Recommendation title
	Description string // Detailed description
	Priority    int    // Priority (1-10)
}

// PersonalizedHelp represents personalized help content
type PersonalizedHelp struct {
	RecommendedNextSteps []string         // Recommended next steps
	ReviewTopics         []string         // Topics to review
	SkillAssessment      string           // Current skill assessment
	PersonalizedTips     []string         // Personalized tips
	Recommendations      []Recommendation // Learning recommendations
}

// UserProfile represents user profile
type UserProfile struct {
	UserID          string          // User identifier
	SkillLevel      SkillLevel      // Current skill level
	PreferredFormat HelpFormat      // Preferred help format
	CompletedTasks  []CompletedTask // Completed tasks
	LearningGoals   []LearningGoal  // Learning goals
	LastActivity    time.Time       // Last activity time
	TotalCommands   int             // Total commands executed
	ErrorCount      int             // Total error count
	SuccessRate     float64         // Success rate
}

// HelpDatabase represents help content database
type HelpDatabase struct {
	commonMistakes  []CommonMistake
	tutorialSteps   []TutorialStep
	conceptMap      map[string]*ConceptExplanation
	migrationGuides map[string]*MigrationGuide
}

// UserFriendlyHelpSystem provides user-friendly help functionality
type UserFriendlyHelpSystem struct {
	commandValidator       *MainCommandValidator
	subcommandValidator    *SubcommandValidator
	errorFormatter         *ComprehensiveErrorFormatter
	helpDatabase           *HelpDatabase
	userProfile            *UserProfile
	interactiveModeEnabled bool
}

// InteractiveCommandBuilder provides interactive command building
type InteractiveCommandBuilder struct {
	helpSystem  *UserFriendlyHelpSystem
	currentStep BuilderStep
	command     []string
	options     map[string]string
}

// LearningTracker tracks learning progress
type LearningTracker struct {
	userProfile     *UserProfile
	completedTasks  []CompletedTask
	currentGoals    []LearningGoal
	recommendations []Recommendation
}

// NewUserFriendlyHelpSystem creates a new help system
func NewUserFriendlyHelpSystem(
	cmdValidator *MainCommandValidator,
	subValidator *SubcommandValidator,
	formatter *ComprehensiveErrorFormatter,
	interactive bool,
) *UserFriendlyHelpSystem {
	system := &UserFriendlyHelpSystem{
		commandValidator:       cmdValidator,
		subcommandValidator:    subValidator,
		errorFormatter:         formatter,
		helpDatabase:           NewHelpDatabase(),
		userProfile:            loadOrCreateUserProfile(),
		interactiveModeEnabled: interactive,
	}

	return system
}

// NewDefaultUserFriendlyHelpSystem creates a help system with default settings
func NewDefaultUserFriendlyHelpSystem() *UserFriendlyHelpSystem {
	return NewUserFriendlyHelpSystem(
		NewMainCommandValidator(),
		NewSubcommandValidator(NewMainCommandValidator()),
		NewDefaultComprehensiveErrorFormatter(),
		true,
	)
}

// ShowHelp displays context-dependent help
func (h *UserFriendlyHelpSystem) ShowHelp(context *HelpContext) error {
	if context == nil {
		context = h.createDefaultContext()
	}

	// Update last accessed time
	context.LastAccessed = time.Now()

	switch context.PreferredFormat {
	case FormatInteractive:
		return h.ShowInteractiveHelp()
	case FormatDetailed:
		return h.showDetailedHelp(context)
	case FormatExample:
		return h.showExampleHelp(context)
	default:
		return h.showBasicHelp(context)
	}
}

// ShowInteractiveHelp displays interactive help
func (h *UserFriendlyHelpSystem) ShowInteractiveHelp() error {
	if !h.interactiveModeEnabled {
		fmt.Println("インタラクティブモードが無効になっています。")
		return nil
	}

	builder := &InteractiveCommandBuilder{
		helpSystem: h,
		options:    make(map[string]string),
	}

	fmt.Println("🚀 usacloudコマンド構築ヘルパーへようこそ！")
	fmt.Println("   ステップごとにコマンドを作成していきます。")
	fmt.Println("   終了するには 'quit' と入力してください。")

	command, err := builder.BuildCommand()
	if err != nil {
		return fmt.Errorf("コマンド構築中にエラーが発生しました: %w", err)
	}

	if command != "" {
		fmt.Printf("✅ 生成されたコマンド: %s\n", command)
	}

	return nil
}

// ShowTutorial displays tutorial content
func (h *UserFriendlyHelpSystem) ShowTutorial() error {
	fmt.Println("📚 usacloud チュートリアル")
	fmt.Println("======================")

	steps := h.helpDatabase.tutorialSteps
	if len(steps) == 0 {
		fmt.Println("チュートリアルコンテンツを準備中です。")
		return nil
	}

	for i, step := range steps {
		fmt.Printf("\n%d. %s\n", i+1, step.Title)
		fmt.Printf("   %s\n", step.Description)

		if len(step.Commands) > 0 {
			fmt.Println("   試してみるコマンド:")
			for _, cmd := range step.Commands {
				fmt.Printf("   $ %s\n", cmd)
			}
		}

		if len(step.Tips) > 0 {
			fmt.Println("   💡 Tips:")
			for _, tip := range step.Tips {
				fmt.Printf("   • %s\n", tip)
			}
		}
	}

	return nil
}

// ShowCommonMistakes displays common mistakes and solutions
func (h *UserFriendlyHelpSystem) ShowCommonMistakes() error {
	fmt.Println("⚠️  よくある間違いと解決方法")
	fmt.Println("==========================")

	mistakes := h.helpDatabase.commonMistakes
	if len(mistakes) == 0 {
		fmt.Println("よくある間違いのデータベースを構築中です。")
		return nil
	}

	for i, mistake := range mistakes {
		fmt.Printf("\n%d. %s\n", i+1, mistake.Description)
		fmt.Printf("   間違い例: %s\n", mistake.Pattern)

		if len(mistake.CorrectExamples) > 0 {
			fmt.Println("   正しい例:")
			for _, example := range mistake.CorrectExamples {
				fmt.Printf("   ✅ %s\n", example)
			}
		}

		if mistake.Explanation != "" {
			fmt.Printf("   説明: %s\n", mistake.Explanation)
		}
	}

	return nil
}

// BuildCommand builds command interactively
func (b *InteractiveCommandBuilder) BuildCommand() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Step 1: Main command selection
	mainCmd, err := b.selectMainCommand(reader)
	if err != nil || mainCmd == "quit" {
		return "", err
	}

	// Step 2: Subcommand selection
	subCmd, err := b.selectSubCommand(reader, mainCmd)
	if err != nil || subCmd == "quit" {
		return "", err
	}

	// Step 3: Generate final command
	finalCommand := b.generateFinalCommand(mainCmd, subCmd, b.options)

	return finalCommand, nil
}

// selectMainCommand selects main command interactively
func (b *InteractiveCommandBuilder) selectMainCommand(reader *bufio.Reader) (string, error) {
	fmt.Println("📋 1. メインコマンドを選択してください:")
	fmt.Println("   よく使われるコマンド:")
	fmt.Println("   • server    - サーバー操作")
	fmt.Println("   • disk      - ディスク操作")
	fmt.Println("   • database  - データベース操作")
	fmt.Println("   • config    - 設定操作")
	fmt.Println("")
	fmt.Println("   すべてのコマンド: usacloud --help")
	fmt.Printf("\n入力してください: ")

	command, _ := reader.ReadString('\n')
	command = strings.TrimSpace(command)

	if command == "quit" {
		return command, nil
	}

	// Validate command
	if !b.helpSystem.commandValidator.IsValidCommand(command) {
		suggestions := b.helpSystem.commandValidator.getSimilarCommands(command, 3)
		if len(suggestions) > 0 {
			fmt.Printf("\n❓ '%s' は有効なコマンドではありません。\n", command)
			fmt.Println("   もしかして以下のコマンドですか？")
			for i, suggestion := range suggestions {
				fmt.Printf("   %d. %s\n", i+1, suggestion)
			}
			fmt.Printf("\n番号を選択するか、正しいコマンドを入力してください: ")

			choice, _ := reader.ReadString('\n')
			choice = strings.TrimSpace(choice)

			// Try to parse as number
			if choice >= "1" && choice <= fmt.Sprintf("%d", len(suggestions)) {
				idx := int(choice[0] - '1')
				if idx < len(suggestions) {
					command = suggestions[idx]
				}
			} else {
				command = choice
			}
		}
	}

	return command, nil
}

// selectSubCommand selects subcommand interactively
func (b *InteractiveCommandBuilder) selectSubCommand(reader *bufio.Reader, mainCmd string) (string, error) {
	fmt.Printf("\n📋 2. '%s' コマンドのサブコマンドを選択してください:\n", mainCmd)

	available := b.helpSystem.subcommandValidator.GetAvailableSubcommands(mainCmd)
	if len(available) == 0 {
		fmt.Println("   このコマンドにはサブコマンドがありません。")
		return "", nil
	}

	fmt.Println("   利用可能なサブコマンド:")
	for _, sub := range available {
		fmt.Printf("   • %s\n", sub)
	}
	fmt.Printf("\n入力してください: ")

	subCommand, _ := reader.ReadString('\n')
	subCommand = strings.TrimSpace(subCommand)

	if subCommand == "quit" {
		return subCommand, nil
	}

	// Validate subcommand
	if !b.helpSystem.subcommandValidator.IsValidSubcommand(mainCmd, subCommand) {
		suggestions := b.helpSystem.subcommandValidator.getSimilarSubcommands(mainCmd, subCommand)
		if len(suggestions) > 0 {
			fmt.Printf("\n❓ '%s' は有効なサブコマンドではありません。\n", subCommand)
			fmt.Println("   もしかして以下のサブコマンドですか？")
			for i, suggestion := range suggestions {
				fmt.Printf("   %d. %s\n", i+1, suggestion)
			}
		}
	}

	return subCommand, nil
}

// generateFinalCommand generates the final command string
func (b *InteractiveCommandBuilder) generateFinalCommand(mainCmd, subCmd string, options map[string]string) string {
	parts := []string{"usacloud", mainCmd}

	if subCmd != "" {
		parts = append(parts, subCmd)
	}

	for key, value := range options {
		if value != "" {
			parts = append(parts, fmt.Sprintf("--%s=%s", key, value))
		} else {
			parts = append(parts, fmt.Sprintf("--%s", key))
		}
	}

	return strings.Join(parts, " ")
}

// createDefaultContext creates a default help context
func (h *UserFriendlyHelpSystem) createDefaultContext() *HelpContext {
	return &HelpContext{
		RequestedCommand: "",
		PreviousErrors:   []ErrorHistory{},
		UserSkillLevel:   h.userProfile.SkillLevel,
		PreferredFormat:  h.userProfile.PreferredFormat,
		LastAccessed:     time.Now(),
	}
}

// showBasicHelp displays basic help content
func (h *UserFriendlyHelpSystem) showBasicHelp(context *HelpContext) error {
	skillLevel := context.UserSkillLevel

	switch skillLevel {
	case SkillBeginner:
		return h.showBeginnerHelp(context)
	case SkillAdvanced, SkillExpert:
		return h.showAdvancedHelp(context)
	default:
		return h.showIntermediateHelp(context)
	}
}

// showBeginnerHelp displays beginner-friendly help
func (h *UserFriendlyHelpSystem) showBeginnerHelp(context *HelpContext) error {
	fmt.Println("🎯 usacloud ヘルプ - 初心者向けガイド")
	fmt.Println("")
	fmt.Println("基本的な使い方:")
	fmt.Println("  usacloud [コマンド] [サブコマンド] [オプション]")
	fmt.Println("")
	fmt.Println("まず試してみましょう:")
	fmt.Println("  usacloud config list          # 設定確認")
	fmt.Println("  usacloud server list          # サーバー一覧表示")
	fmt.Println("")
	fmt.Println("よく使うコマンド:")
	fmt.Println("  • server  - サーバー操作 (作成、一覧、詳細等)")
	fmt.Println("  • disk    - ディスク操作 (作成、一覧、接続等)")
	fmt.Println("  • config  - 設定操作 (プロファイル管理等)")
	fmt.Println("")
	fmt.Println("🚀 インタラクティブモード: usacloud --interactive")
	fmt.Println("📚 詳細な学習ガイド: usacloud --tutorial")
	fmt.Println("❓ 困ったとき: usacloud --help [コマンド名]")

	return nil
}

// showIntermediateHelp displays intermediate help
func (h *UserFriendlyHelpSystem) showIntermediateHelp(context *HelpContext) error {
	fmt.Println("📋 usacloud ヘルプ - 中級者向けガイド")
	fmt.Println("")
	fmt.Println("よく使う操作パターン:")
	fmt.Println("  usacloud server list --output-type json")
	fmt.Println("  usacloud server read [ID] --format table")
	fmt.Println("  usacloud disk create --size 20 --name my-disk")
	fmt.Println("")
	fmt.Println("効率的な使い方:")
	fmt.Println("  • JSON出力でデータ処理: --output-type json")
	fmt.Println("  • フィルタリング: --selector や引数での絞り込み")
	fmt.Println("  • ゾーン指定: --zone で処理範囲を限定")
	fmt.Println("")
	fmt.Println("🔧 高度な機能:")
	fmt.Println("  usacloud rest get /v1/server  # 直接API呼び出し")
	fmt.Println("  usacloud config current        # 現在の設定確認")

	return nil
}

// showAdvancedHelp displays advanced help
func (h *UserFriendlyHelpSystem) showAdvancedHelp(context *HelpContext) error {
	fmt.Println("⚡ usacloud 高度な使用方法")
	fmt.Println("")
	fmt.Println("効率的な使い方:")
	fmt.Println("  usacloud server list --output-type json | jq '.[] | select(.Name | contains(\"web\"))'")
	fmt.Println("  usacloud server read --format table --selector 'Name=\"web-server\"'")
	fmt.Println("")
	fmt.Println("自動化Tips:")
	fmt.Println("  • JSON出力 + jq でのフィルタリング")
	fmt.Println("  • --output-type csv でのデータ処理")
	fmt.Println("  • 環境変数での認証設定")
	fmt.Println("")
	fmt.Println("パフォーマンス最適化:")
	fmt.Println("  • --zone 指定で検索範囲を限定")
	fmt.Println("  • --selector での効率的なフィルタリング")
	fmt.Println("  • バッチ処理でのAPI呼び出し削減")
	fmt.Println("")
	fmt.Println("🔧 開発者向け機能:")
	fmt.Println("  • usacloud rest - 直接API呼び出し")
	fmt.Println("  • --debug でデバッグ情報表示")
	fmt.Println("  • カスタムプロファイルでの環境切替")

	return nil
}

// showDetailedHelp displays detailed help
func (h *UserFriendlyHelpSystem) showDetailedHelp(context *HelpContext) error {
	fmt.Println("📖 usacloud 詳細ヘルプ")
	fmt.Println("===================")

	// Show basic help first
	err := h.showBasicHelp(context)
	if err != nil {
		return err
	}

	// Show common mistakes if available
	fmt.Println("\n⚠️  よくある間違い:")
	mistakes := h.helpDatabase.commonMistakes
	for i, mistake := range mistakes {
		if i >= 3 { // Show only top 3
			break
		}
		fmt.Printf("• %s\n", mistake.Description)
		fmt.Printf("  ❌ %s\n", mistake.Pattern)
		if len(mistake.CorrectExamples) > 0 {
			fmt.Printf("  ✅ %s\n", mistake.CorrectExamples[0])
		}
		fmt.Println()
	}

	return nil
}

// showExampleHelp displays example-focused help
func (h *UserFriendlyHelpSystem) showExampleHelp(context *HelpContext) error {
	fmt.Println("💡 usacloud 実例集")
	fmt.Println("================")
	fmt.Println("")
	fmt.Println("🖥️  サーバー操作:")
	fmt.Println("  usacloud server list                    # 全サーバー表示")
	fmt.Println("  usacloud server list web-server         # 名前で検索")
	fmt.Println("  usacloud server read 123456789          # サーバー詳細")
	fmt.Println("  usacloud server power-on 123456789      # サーバー起動")
	fmt.Println("")
	fmt.Println("💾 ディスク操作:")
	fmt.Println("  usacloud disk list                      # 全ディスク表示")
	fmt.Println("  usacloud disk create --size 20 --name my-disk  # ディスク作成")
	fmt.Println("  usacloud disk connect 123456789 987654321      # ディスク接続")
	fmt.Println("")
	fmt.Println("⚙️  設定操作:")
	fmt.Println("  usacloud config list                    # プロファイル一覧")
	fmt.Println("  usacloud config current                 # 現在の設定")
	fmt.Println("  usacloud config use production          # プロファイル切替")

	return nil
}

// NewHelpDatabase creates a new help database
func NewHelpDatabase() *HelpDatabase {
	db := &HelpDatabase{
		commonMistakes:  getCommonMistakes(),
		tutorialSteps:   getTutorialSteps(),
		conceptMap:      make(map[string]*ConceptExplanation),
		migrationGuides: make(map[string]*MigrationGuide),
	}

	// Initialize concept map
	db.initializeConceptMap()

	// Initialize migration guides
	db.initializeMigrationGuides()

	return db
}

// getCommonMistakes returns list of common mistakes
func getCommonMistakes() []CommonMistake {
	return []CommonMistake{
		{
			Pattern:         "usacloud server show",
			Description:     "v0での'show'は'read'に変更されました",
			CorrectExamples: []string{"usacloud server read [ID]"},
			Explanation:     "usacloud v1では一貫性のため、単一リソースの取得は'read'コマンドを使用します",
			RelatedTopics:   []string{"CRUD operations", "v0 to v1 migration"},
			Frequency:       95,
		},
		{
			Pattern:         "usacloud server list --selector",
			Description:     "セレクタ機能は廃止され、直接引数で指定します",
			CorrectExamples: []string{"usacloud server list [NAME_OR_ID]"},
			Explanation:     "--selectorオプションは廃止されました。名前やIDは直接引数として指定してください",
			RelatedTopics:   []string{"selector deprecation", "argument passing"},
			Frequency:       87,
		},
		{
			Pattern:         "usacloud iso-image list",
			Description:     "iso-imageコマンドは'cdrom'に名称変更されました",
			CorrectExamples: []string{"usacloud cdrom list"},
			Explanation:     "ISOイメージ関連の操作は'cdrom'コマンドに統合されました",
			RelatedTopics:   []string{"command renaming", "iso to cdrom migration"},
			Frequency:       76,
		},
	}
}

// getTutorialSteps returns tutorial steps
func getTutorialSteps() []TutorialStep {
	return []TutorialStep{
		{
			StepID:      "step1",
			Title:       "設定確認",
			Description: "まず現在の設定を確認しましょう",
			Commands:    []string{"usacloud config current", "usacloud config list"},
			Tips:        []string{"複数のプロファイルを使い分けることで、開発・本番環境を管理できます"},
		},
		{
			StepID:      "step2",
			Title:       "リソース一覧表示",
			Description: "基本的な一覧表示操作を学びます",
			Commands:    []string{"usacloud server list", "usacloud disk list"},
			Tips:        []string{"--output-type json を使うとデータ処理に便利です"},
		},
		{
			StepID:      "step3",
			Title:       "詳細情報取得",
			Description: "特定のリソースの詳細を確認します",
			Commands:    []string{"usacloud server read [ID]", "usacloud disk read [ID]"},
			Tips:        []string{"IDの代わりに名前でも検索できます"},
		},
	}
}

// initializeConceptMap initializes concept explanations
func (db *HelpDatabase) initializeConceptMap() {
	db.conceptMap["crud"] = &ConceptExplanation{
		ConceptID:   "crud",
		Title:       "CRUD操作",
		Description: "Create (作成), Read (読み取り), Update (更新), Delete (削除) の基本操作",
		Examples:    []string{"create", "read", "update", "delete"},
		SeeAlso:     []string{"commands", "resources"},
	}

	db.conceptMap["selector"] = &ConceptExplanation{
		ConceptID:   "selector",
		Title:       "セレクター機能",
		Description: "v0で使用されていたリソース絞り込み機能（v1では廃止）",
		Examples:    []string{"--selector 'Name=\"test\"'", "直接引数指定に変更"},
		SeeAlso:     []string{"migration", "filtering"},
	}
}

// initializeMigrationGuides initializes migration guides
func (db *HelpDatabase) initializeMigrationGuides() {
	db.migrationGuides["v0_to_v1"] = &MigrationGuide{
		FromVersion: "v0",
		ToVersion:   "v1",
		Changes: []MigrationChange{
			{
				OldCommand: "usacloud server show [ID]",
				NewCommand: "usacloud server read [ID]",
				Reason:     "CRUD操作の一貫性向上",
				Impact:     "コマンド名変更のみ、機能は同等",
			},
		},
		Examples: []MigrationExample{
			{
				Scenario:    "サーバー詳細表示",
				OldCommand:  "usacloud server show 123456789",
				NewCommand:  "usacloud server read 123456789",
				Explanation: "単一リソースの取得は'read'コマンドを使用",
			},
		},
	}
}

// loadOrCreateUserProfile loads or creates user profile
func loadOrCreateUserProfile() *UserProfile {
	// In a real implementation, this would load from a file or database
	// For now, return a default profile
	return &UserProfile{
		UserID:          "default",
		SkillLevel:      SkillBeginner,
		PreferredFormat: FormatBasic,
		CompletedTasks:  []CompletedTask{},
		LearningGoals:   []LearningGoal{},
		LastActivity:    time.Now(),
		TotalCommands:   0,
		ErrorCount:      0,
		SuccessRate:     0.0,
	}
}

// GetSkillLevelString returns skill level as string
func GetSkillLevelString(skill SkillLevel) string {
	switch skill {
	case SkillBeginner:
		return "Beginner"
	case SkillIntermediate:
		return "Intermediate"
	case SkillAdvanced:
		return "Advanced"
	case SkillExpert:
		return "Expert"
	default:
		return "Unknown"
	}
}

// GetHelpFormatString returns help format as string
func GetHelpFormatString(format HelpFormat) string {
	switch format {
	case FormatBasic:
		return "Basic"
	case FormatDetailed:
		return "Detailed"
	case FormatInteractive:
		return "Interactive"
	case FormatExample:
		return "Example"
	default:
		return "Unknown"
	}
}
