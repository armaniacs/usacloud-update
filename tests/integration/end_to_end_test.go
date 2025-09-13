package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEndToEnd_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("短時間テストモードではスキップ")
	}

	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	scenarios := []struct {
		name         string
		scenarioFile string
	}{
		{
			name:         "Beginner user with typos",
			scenarioFile: "scenarios/beginner_user_typos.yaml",
		},
		{
			name:         "Expert user batch processing",
			scenarioFile: "scenarios/expert_batch_processing.yaml",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			if !fileExists(suite.testDataDir + "/" + sc.scenarioFile) {
				t.Skipf("シナリオファイルが見つかりません: %s", sc.scenarioFile)
				return
			}
			suite.RunScenarioFromFile(sc.scenarioFile)
		})
	}
}

func TestEndToEnd_BasicValidation(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	scenarios := []TestScenario{
		{
			Name: "Valid command processing",
			Input: ScenarioInput{
				Type:      "command",
				Arguments: []string{"--validate-only"},
				Content:   "usacloud server list",
			},
			Expected: ScenarioExpected{
				ExitCode: 0,
				OutputContains: []string{
					"検証完了",
				},
			},
		},
		{
			Name: "Invalid command with suggestion",
			Input: ScenarioInput{
				Type:      "command",
				Arguments: []string{"--validate-only"},
				Content:   "usacloud serv list",
			},
			Expected: ScenarioExpected{
				ExitCode: 1,
				ErrorContains: []string{
					"問題が見つかりました",
					"server",
				},
			},
		},
		{
			Name: "Deprecated command detection",
			Input: ScenarioInput{
				Type:      "command",
				Arguments: []string{"--validate-only"},
				Content:   "usacloud iso-image list",
			},
			Expected: ScenarioExpected{
				ExitCode: 1,
				ErrorContains: []string{
					"iso-image",
					"廃止",
					"cdrom",
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.RunScenario(scenario)
		})
	}
}

func TestEndToEnd_ErrorHandling(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	errorScenarios := []TestScenario{
		{
			Name: "Multiple validation errors in single line",
			Input: ScenarioInput{
				Type:      "command",
				Arguments: []string{"--validate-only"},
				Content:   "usacloud invalid-cmd invalid-sub",
			},
			Expected: ScenarioExpected{
				ExitCode: 1,
				ErrorContains: []string{
					"invalid-cmd",
					"有効なusacloudコマンドではありません",
				},
			},
		},
		{
			Name: "Empty input handling",
			Input: ScenarioInput{
				Type:    "stdin",
				Content: "",
			},
			Expected: ScenarioExpected{
				ExitCode: 0,
			},
		},
		{
			Name: "Long command line handling",
			Input: ScenarioInput{
				Type:      "command",
				Arguments: []string{"--validate-only"},
				Content:   "usacloud invalid-very-long-command-name list",
			},
			Expected: ScenarioExpected{
				ExitCode: 1,
				ErrorContains: []string{
					"invalid-very-long-command-name",
					"有効なusacloudコマンドではありません",
				},
			},
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.RunScenario(scenario)
		})
	}
}

func TestEndToEnd_ProfileIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	profiles := []struct {
		name           string
		profile        string
		command        string
		expectBehavior string
	}{
		{
			name:           "Default profile behavior",
			profile:        "default",
			command:        "usacloud invalid-command list",
			expectBehavior: "standard_error",
		},
		{
			name:           "Beginner profile verbose help",
			profile:        "beginner",
			command:        "usacloud invalid-command list",
			expectBehavior: "verbose_help",
		},
	}

	for _, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			scenario := TestScenario{
				Name: profile.name,
				Input: ScenarioInput{
					Type:      "command",
					Arguments: []string{"--validate-only"},
					Content:   profile.command,
				},
				Expected: ScenarioExpected{
					ExitCode: 1,
					ErrorContains: []string{
						"有効なusacloudコマンドではありません",
					},
				},
			}

			suite.RunScenario(scenario)
		})
	}
}

func TestEndToEnd_TransformationFlow(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	// 有効なコマンドのみのテストファイルを作成
	testContent := `#!/bin/bash
# Valid commands for transformation test
usacloud server list --output-type csv
usacloud disk list --output-type tsv
usacloud database list
echo "non-usacloud line"
`
	inputFile := filepath.Join(suite.tempDir, "transform-test.sh")
	err := os.WriteFile(inputFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("テストファイル作成エラー: %v", err)
	}

	scenario := TestScenario{
		Name: "File transformation workflow",
		Input: ScenarioInput{
			Type:      "file",
			FilePath:  inputFile,
			Arguments: []string{"--out", "-"},
		},
		Expected: ScenarioExpected{
			ExitCode: 0,
			OutputContains: []string{
				"usacloud-update",
			},
		},
	}

	suite.RunScenario(scenario)
}

func TestEndToEnd_ConfigIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	scenarios := []TestScenario{
		{
			Name: "Config file usage",
			Input: ScenarioInput{
				Type:      "command",
				Arguments: []string{"--validate-only"},
				Content:   "usacloud server list",
			},
			Expected: ScenarioExpected{
				ExitCode: 0,
			},
		},
		{
			Name: "Environment variable override",
			Input: ScenarioInput{
				Type:      "command",
				Arguments: []string{"--validate-only"},
				Content:   "usacloud server list",
				Environment: map[string]string{
					"USACLOUD_UPDATE_VERBOSE": "true",
				},
			},
			Expected: ScenarioExpected{
				ExitCode: 0,
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.RunScenario(scenario)
		})
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
