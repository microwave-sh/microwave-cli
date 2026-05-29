package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/config"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

type LoginCmd struct {
	Key string `arg:"" optional:"" help:"Management API key. If omitted, prompts (and prints the console URL)."`
}

func (c *LoginCmd) Run(g *Globals) error {
	key := strings.TrimSpace(c.Key)
	if key == "" {
		fmt.Printf("Create a management API key at %s\n", output.Bold.Render("https://app.microwave.sh"))
		fmt.Print("Paste your key: ")
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		key = strings.TrimSpace(line)
	}
	if key == "" {
		return fmt.Errorf("no key provided")
	}
	if err := config.WriteGlobalAuth(key); err != nil {
		return err
	}
	fmt.Printf("%s Saved to %s\n",
		output.Green.Render("✓"),
		output.Dim.Render(filepath.Join(config.GlobalConfigDir(), "config.toml")))
	return nil
}
