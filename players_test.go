package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/RLBot/go-interface/flat"
)

func TestExpandEnvironmentValue(t *testing.T) {
	t.Setenv("RLBOT_TEST_VALUE", "expanded")
	t.Setenv("RLBOT_TEST_EMPTY", "")
	t.Setenv("RLBOT_TEST_FILES(x86)", "/program files")
	t.Setenv("RLBOT_TEST_DOLLAR", "$RLBOT_TEST_VALUE")
	t.Setenv("RLBOT_TEST_PERCENT", "%RLBOT_TEST_VALUE%")
	t.Setenv("RLBOT_TEST_MISSING", "")
	if err := os.Unsetenv("RLBOT_TEST_MISSING"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, input, want string
	}{
		{"literal", "literal/path", "literal/path"},
		{"bare", "$RLBOT_TEST_VALUE/lib", "expanded/lib"},
		{"braced", "${RLBOT_TEST_VALUE}/lib", "expanded/lib"},
		{"percent", "%RLBOT_TEST_VALUE%/lib", "expanded/lib"},
		{"parentheses", "%RLBOT_TEST_FILES(x86)%/lib", "/program files/lib"},
		{"adjacent", "%RLBOT_TEST_VALUE%$RLBOT_TEST_VALUE${RLBOT_TEST_VALUE}", "expandedexpandedexpanded"},
		{"missing", "$RLBOT_TEST_MISSING/${RLBOT_TEST_MISSING}/%RLBOT_TEST_MISSING%", "$RLBOT_TEST_MISSING/${RLBOT_TEST_MISSING}/%RLBOT_TEST_MISSING%"},
		{"empty", "a${RLBOT_TEST_EMPTY}%RLBOT_TEST_EMPTY%$RLBOT_TEST_EMPTY/b", "a/b"},
		{"dollar in replacement", "%RLBOT_TEST_DOLLAR%", "$RLBOT_TEST_VALUE"},
		{"percent in replacement", "$RLBOT_TEST_PERCENT", "%RLBOT_TEST_VALUE%"},
		{"same syntax in replacement", "${RLBOT_TEST_DOLLAR}|%RLBOT_TEST_PERCENT%", "$RLBOT_TEST_VALUE|%RLBOT_TEST_VALUE%"},
		{"incomplete", "${RLBOT_TEST_VALUE/%RLBOT_TEST_VALUE", "${RLBOT_TEST_VALUE/%RLBOT_TEST_VALUE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandEnvironmentValue(tt.input); got != tt.want {
				t.Fatalf("expandEnvironmentValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPlatformEnvironmentLaunchAndJSON(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		t.Skip("Platform overrides require Windows or Linux")
	}
	t.Setenv("RLBOT_TEST_VALUE", "expanded")
	t.Setenv("RLBOT_TEST_LITERAL", "$RLBOT_TEST_VALUE")
	var config BotConfig
	_, err := toml.Decode(`
[settings.environment]
COMMON = "$RLBOT_TEST_VALUE/%RLBOT_TEST_VALUE%"
SHARED = "common"
[settings.environment.windows]
SHARED = "windows"
PLATFORM = "%RLBOT_TEST_VALUE%/lib"
LITERAL = "%RLBOT_TEST_LITERAL%"
[settings.environment.linux]
SHARED = "linux"
PLATFORM = "${RLBOT_TEST_VALUE}/lib"
LITERAL = "%RLBOT_TEST_LITERAL%"
`, &config)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"COMMON":   "$RLBOT_TEST_VALUE/%RLBOT_TEST_VALUE%",
		"SHARED":   runtime.GOOS,
		"PLATFORM": "expanded/lib",
		"LITERAL":  "$RLBOT_TEST_VALUE",
	}
	if got := config.Settings.Environment.Values(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Values() = %v, want %v", got, want)
	}

	info := BotInfo{Config: config, TomlPath: filepath.Join(t.TempDir(), "test.bot.toml")}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip BotInfo
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range []BotInfo{info, roundTrip} {
		player := candidate.ToPlayerConfig(0).Variety.Value.(*flat.CustomBotT)
		for name, environment := range map[string][]*flat.EnvironmentVariableT{
			"bot":    player.Environment,
			"script": candidate.ToScriptConfig().Environment,
		} {
			got := make(map[string]string)
			for _, variable := range environment {
				got[variable.Name] = variable.Value
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s environment = %v, want %v", name, got, want)
			}
		}
	}
}

func TestAgentLoadersSkipInvalidEnvironment(t *testing.T) {
	for _, kind := range []string{"bot", "script"} {
		t.Run(kind, func(t *testing.T) {
			directory := t.TempDir()
			configs := map[string]string{
				"valid":   "[settings]\nname = 'valid'\n[settings.environment]\nVALUE = 'value'\n",
				"invalid": "[settings]\nname = 'invalid'\n[settings.environment.windows]\nVALUE = 123\n",
			}
			for name, data := range configs {
				if err := os.WriteFile(filepath.Join(directory, name+"."+kind+".toml"), []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			app := &App{}
			var infos []BotInfo
			if kind == "bot" {
				infos = app.GetBots([]string{directory})
			} else {
				infos = app.GetScripts([]string{directory})
			}
			if len(infos) != 1 || infos[0].Config.Settings.Name != "valid" {
				t.Fatalf("Expected only the valid %s config, got %v", kind, infos)
			}
		})
	}
}
