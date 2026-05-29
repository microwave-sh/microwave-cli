package cmd

import "fmt"

type CompletionCmd struct {
	Shell string `arg:"" enum:"bash,zsh,fish" help:"Shell type (bash, zsh, fish)."`
}

func (c *CompletionCmd) Run(g *Globals) error {
	switch c.Shell {
	case "bash":
		fmt.Println("# Add to ~/.bashrc:\n# source <(microwave completion bash)")
		fmt.Println(`complete -o default -C "microwave" microwave`)
	case "zsh":
		fmt.Println("# Add to ~/.zshrc:\n# source <(microwave completion zsh)")
		fmt.Println(`compdef _microwave microwave`)
	case "fish":
		fmt.Println("# microwave completion fish | source")
		fmt.Println(`complete -c microwave -f`)
	}
	return nil
}
