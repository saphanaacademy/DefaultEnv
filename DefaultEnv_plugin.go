package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

type DefaultEnvPlugin struct{}

func handleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

type response struct {
	SystemEnvJson        map[string]interface{} `json:"system_env_json"`
	ApplicationEnvJson   map[string]interface{} `json:"application_env_json"`
	EnvironmentVariables map[string]interface{} `json:"environment_variables"`
}

func (c *DefaultEnvPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] != "default-env" {
		return
	}
	if len(args) != 2 {
		fmt.Println("Please specify an app")
		return
	}
	app, err := cliConnection.GetApp(args[1])
	handleError(err)

	url := fmt.Sprintf("/v3/apps/%s/env", app.Guid)
	env, err := cliConnection.CliCommandWithoutTerminalOutput("curl", url)
	handleError(err)

	var data response
	err = json.Unmarshal([]byte(strings.Join(env, "")), &data)
	handleError(err)

	f, err := os.Create("default-env.json")
	handleError(err)

	// Merge all environment variables into one map
	content := make(map[string]interface{})
	for k, v := range data.SystemEnvJson {
		content[k] = v
	}
	for k, v := range data.ApplicationEnvJson {
		content[k] = v
	}
	for k, v := range data.EnvironmentVariables {
		content[k] = v
	}

	write, err := json.MarshalIndent(content, "", "  ")
	handleError(err)

	_, err = f.Write(write)
	handleError(err)

	err = f.Close()
	handleError(err)
	fmt.Println("Environment variables for " + args[1] + " written to default-env.json")
}

func (c *DefaultEnvPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "DefaultEnv",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 1,
		},
		MinCliVersion: plugin.VersionType{
			Major: 7,
			Minor: 2,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "default-env",
				Alias:    "de",
				HelpText: "Create default-env.json file with environment variables of an app.",
				UsageDetails: plugin.Usage{
					Usage: "cf default-env APP",
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(DefaultEnvPlugin))
}
