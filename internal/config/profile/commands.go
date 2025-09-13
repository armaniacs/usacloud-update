package profile

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// ProfileCommand implements CLI commands for profile management
type ProfileCommand struct {
	manager         *ProfileManager
	templateManager *TemplateManager
}

// NewProfileCommand creates a new profile command handler
func NewProfileCommand(manager *ProfileManager, templateManager *TemplateManager) *ProfileCommand {
	return &ProfileCommand{
		manager:         manager,
		templateManager: templateManager,
	}
}

// ListProfiles lists all profiles in a table format
func (pc *ProfileCommand) ListProfiles(cmd *cobra.Command, args []string) error {
	environment, _ := cmd.Flags().GetString("environment")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	sortBy, _ := cmd.Flags().GetString("sort")
	sortOrder, _ := cmd.Flags().GetString("order")

	opts := ProfileListOptions{
		Environment: environment,
		Tags:        tags,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
	}

	profiles := pc.manager.ListProfiles(opts)

	if len(profiles) == 0 {
		fmt.Println("プロファイルが見つかりません。")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("ID", "名前", "環境", "最終使用", "デフォルト", "タグ")

	activeProfile := pc.manager.GetActiveProfile()

	for _, profile := range profiles {
		isDefault := ""
		if profile.IsDefault {
			isDefault = "✓"
		}

		isActive := ""
		if activeProfile != nil && activeProfile.ID == profile.ID {
			isActive = " (現在)"
		}

		lastUsed := "未使用"
		if !profile.LastUsedAt.IsZero() {
			lastUsed = profile.LastUsedAt.Format("2006-01-02 15:04")
		}

		tags := strings.Join(profile.Tags, ", ")
		if len(tags) > 30 {
			tags = tags[:27] + "..."
		}

		_ = table.Append([]string{
			profile.ID[:8] + "...",
			profile.Name + isActive,
			profile.Environment,
			lastUsed,
			isDefault,
			tags,
		})
	}

	_ = table.Render()
	return nil
}

// ShowProfile displays detailed information about a profile
func (pc *ProfileCommand) ShowProfile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("プロファイルIDまたは名前を指定してください")
	}

	profile, err := pc.manager.GetProfile(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("プロファイル詳細\n")
	fmt.Printf("================\n")
	fmt.Printf("ID: %s\n", profile.ID)
	fmt.Printf("名前: %s\n", profile.Name)
	fmt.Printf("説明: %s\n", profile.Description)
	fmt.Printf("環境: %s\n", profile.Environment)

	if profile.ParentID != "" {
		parent, err := pc.manager.GetProfile(profile.ParentID)
		if err == nil {
			fmt.Printf("継承元: %s\n", parent.Name)
		} else {
			fmt.Printf("継承元: %s (見つかりません)\n", profile.ParentID)
		}
	}

	fmt.Printf("作成日時: %s\n", profile.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("更新日時: %s\n", profile.UpdatedAt.Format("2006-01-02 15:04:05"))

	if !profile.LastUsedAt.IsZero() {
		fmt.Printf("最終使用: %s\n", profile.LastUsedAt.Format("2006-01-02 15:04:05"))
	}

	if profile.IsDefault {
		fmt.Printf("デフォルト: ✓\n")
	}

	if len(profile.Tags) > 0 {
		fmt.Printf("タグ: %s\n", strings.Join(profile.Tags, ", "))
	}

	fmt.Printf("\n設定項目:\n")
	if len(profile.Config) == 0 {
		fmt.Printf("  (設定項目なし)\n")
	} else {
		for key, value := range profile.Config {
			if IsSensitiveKey(key) {
				fmt.Printf("  %s: %s\n", key, MaskValue(value))
			} else {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	return nil
}

// CreateProfile creates a new profile
func (pc *ProfileCommand) CreateProfile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("プロファイル名を指定してください")
	}

	name := args[0]
	description, _ := cmd.Flags().GetString("description")
	environment, _ := cmd.Flags().GetString("environment")
	templateName, _ := cmd.Flags().GetString("template")
	parentID, _ := cmd.Flags().GetString("parent")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	setDefault, _ := cmd.Flags().GetBool("default")
	configFlags, _ := cmd.Flags().GetStringSlice("config")

	// Parse config flags
	config := make(map[string]string)
	for _, configFlag := range configFlags {
		parts := strings.SplitN(configFlag, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("無効な設定フォーマット: %s (KEY=VALUE形式で指定してください)", configFlag)
		}
		config[parts[0]] = parts[1]
	}

	var profile *Profile
	var err error

	if templateName != "" {
		// Create from template
		profile, err = pc.templateManager.CreateProfileFromTemplate(templateName, name, config)
		if err != nil {
			return fmt.Errorf("テンプレートからプロファイルを作成できませんでした: %w", err)
		}

		// Override description if provided
		if description != "" {
			profile.Description = description
		}

		// Override environment if provided
		if environment != "" {
			profile.Environment = environment
		}

		// Merge tags
		if len(tags) > 0 {
			tagSet := make(map[string]bool)
			for _, tag := range profile.Tags {
				tagSet[tag] = true
			}
			for _, tag := range tags {
				if !tagSet[tag] {
					profile.Tags = append(profile.Tags, tag)
				}
			}
		}

		profile.IsDefault = setDefault

		// Save the profile through manager
		opts := ProfileCreateOptions{
			Name:        profile.Name,
			Description: profile.Description,
			Environment: profile.Environment,
			Config:      profile.Config,
			Tags:        profile.Tags,
			SetDefault:  setDefault,
		}
		profile, err = pc.manager.CreateProfile(opts)

	} else if parentID != "" {
		// Create from parent
		profile, err = pc.manager.CreateProfileFromParent(name, description, parentID, config)
		if err != nil {
			return fmt.Errorf("親プロファイルからプロファイルを作成できませんでした: %w", err)
		}

		// Override environment if provided
		if environment != "" {
			updateOpts := ProfileUpdateOptions{
				Environment: &environment,
			}
			err = pc.manager.UpdateProfile(profile.ID, updateOpts)
			if err != nil {
				return fmt.Errorf("プロファイルの環境を更新できませんでした: %w", err)
			}
		}

	} else {
		// Create new profile
		if environment == "" {
			environment = EnvironmentDevelopment // Default to development
		}

		opts := ProfileCreateOptions{
			Name:        name,
			Description: description,
			Environment: environment,
			Config:      config,
			Tags:        tags,
			SetDefault:  setDefault,
		}
		profile, err = pc.manager.CreateProfile(opts)
	}

	if err != nil {
		return fmt.Errorf("プロファイルを作成できませんでした: %w", err)
	}

	fmt.Printf("プロファイル '%s' を作成しました (ID: %s)\n", profile.Name, profile.ID[:8]+"...")

	if profile.IsDefault {
		fmt.Println("このプロファイルがデフォルトに設定されました。")
	}

	return nil
}

// UpdateProfile updates an existing profile
func (pc *ProfileCommand) UpdateProfile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("プロファイルIDまたは名前を指定してください")
	}

	profile, err := pc.manager.GetProfile(args[0])
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	environment, _ := cmd.Flags().GetString("environment")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	setDefault, _ := cmd.Flags().GetBool("default")
	configFlags, _ := cmd.Flags().GetStringSlice("config")

	opts := ProfileUpdateOptions{}

	if name != "" {
		opts.Name = &name
	}
	if description != "" {
		opts.Description = &description
	}
	if environment != "" {
		opts.Environment = &environment
	}
	if len(tags) > 0 {
		opts.Tags = tags
	}
	if setDefault {
		opts.SetDefault = &setDefault
	}

	// Parse config updates
	if len(configFlags) > 0 {
		opts.Config = make(map[string]string)
		for _, configFlag := range configFlags {
			parts := strings.SplitN(configFlag, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("無効な設定フォーマット: %s (KEY=VALUE形式で指定してください)", configFlag)
			}
			opts.Config[parts[0]] = parts[1]
		}
	}

	if err := pc.manager.UpdateProfile(profile.ID, opts); err != nil {
		return fmt.Errorf("プロファイルを更新できませんでした: %w", err)
	}

	fmt.Printf("プロファイル '%s' を更新しました。\n", profile.Name)
	return nil
}

// DeleteProfile deletes a profile
func (pc *ProfileCommand) DeleteProfile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("プロファイルIDまたは名前を指定してください")
	}

	profile, err := pc.manager.GetProfile(args[0])
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")

	if !force {
		fmt.Printf("プロファイル '%s' を削除しますか? [y/N]: ", profile.Name)
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("削除をキャンセルしました。")
			return nil
		}
	}

	if err := pc.manager.DeleteProfile(profile.ID); err != nil {
		return fmt.Errorf("プロファイルを削除できませんでした: %w", err)
	}

	fmt.Printf("プロファイル '%s' を削除しました。\n", profile.Name)
	return nil
}

// SwitchProfile switches to a different profile
func (pc *ProfileCommand) SwitchProfile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("プロファイルIDまたは名前を指定してください")
	}

	profile, err := pc.manager.GetProfile(args[0])
	if err != nil {
		return err
	}

	if err := pc.manager.SwitchProfile(profile.ID); err != nil {
		return fmt.Errorf("プロファイルを切り替えできませんでした: %w", err)
	}

	fmt.Printf("プロファイル '%s' に切り替えました。\n", profile.Name)
	return nil
}

// ExportProfile exports a profile to a file
func (pc *ProfileCommand) ExportProfile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("プロファイルIDまたは名前を指定してください")
	}

	// --output フラグから出力ファイルを取得
	outputFile, _ := cmd.Flags().GetString("output")
	if outputFile == "" {
		return fmt.Errorf("出力ファイルパス（--output）を指定してください")
	}

	profile, err := pc.manager.GetProfile(args[0])
	if err != nil {
		return err
	}

	if err := pc.manager.ExportProfile(profile.ID, outputFile); err != nil {
		return fmt.Errorf("プロファイルをエクスポートできませんでした: %w", err)
	}

	fmt.Printf("プロファイル '%s' を '%s' にエクスポートしました。\n", profile.Name, outputFile)
	return nil
}

// ImportProfile imports a profile from a file
func (pc *ProfileCommand) ImportProfile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("インポートファイルパスを指定してください")
	}

	inputFile := args[0]

	profile, err := pc.manager.ImportProfile(inputFile)
	if err != nil {
		return fmt.Errorf("プロファイルをインポートできませんでした: %w", err)
	}

	fmt.Printf("プロファイル '%s' を '%s' からインポートしました (ID: %s)。\n",
		profile.Name, inputFile, profile.ID[:8]+"...")
	return nil
}

// ListTemplates lists available templates
func (pc *ProfileCommand) ListTemplates(cmd *cobra.Command, args []string) error {
	environment, _ := cmd.Flags().GetString("environment")

	var templates []ProfileTemplate
	if environment != "" {
		templates = pc.templateManager.GetTemplateByEnvironment(environment)
	} else {
		templates = pc.templateManager.GetAllTemplates()
	}

	if len(templates) == 0 {
		fmt.Println("利用可能なテンプレートがありません。")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("名前", "環境", "説明", "タグ")

	for _, template := range templates {
		tags := strings.Join(template.Tags, ", ")
		if len(tags) > 30 {
			tags = tags[:27] + "..."
		}

		description := template.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		_ = table.Append([]string{
			template.Name,
			template.Environment,
			description,
			tags,
		})
	}

	_ = table.Render()
	return nil
}

// ShowTemplate displays detailed information about a template
func (pc *ProfileCommand) ShowTemplate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("テンプレート名を指定してください")
	}

	template, err := pc.templateManager.GetTemplate(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("テンプレート詳細\n")
	fmt.Printf("================\n")
	fmt.Printf("名前: %s\n", template.Name)
	fmt.Printf("説明: %s\n", template.Description)
	fmt.Printf("環境: %s\n", template.Environment)

	if len(template.Tags) > 0 {
		fmt.Printf("タグ: %s\n", strings.Join(template.Tags, ", "))
	}

	fmt.Printf("\n設定項目:\n")
	for _, keyDef := range template.ConfigKeys {
		required := ""
		if keyDef.Required {
			required = " (必須)"
		}

		defaultValue := ""
		if keyDef.Default != "" {
			defaultValue = fmt.Sprintf(" [デフォルト: %s]", keyDef.Default)
		}

		validation := ""
		if keyDef.Validation != "" {
			validation = fmt.Sprintf(" [検証: %s]", keyDef.Validation)
		}

		fmt.Printf("  %s (%s)%s%s%s\n", keyDef.Key, keyDef.Type, required, defaultValue, validation)
		fmt.Printf("    %s\n", keyDef.Description)
	}

	return nil
}
