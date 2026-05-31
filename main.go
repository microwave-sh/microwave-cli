package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/microwave-sh/microwave-cli/cmd"
	"github.com/microwave-sh/microwave-cli/internal/config"
	"github.com/microwave-sh/microwave-cli/internal/version"
)

type CLI struct {
	Token   string `env:"MICROWAVE_TOKEN" help:"Management API key." hidden:""`
	APIURL  string `name:"api-url" env:"MICROWAVE_API_URL" help:"API base URL override." hidden:""`
	AuthURL string `name:"auth-url" env:"MICROWAVE_AUTH_URL" help:"Auth-plane base URL override." hidden:""`
	Output  string `name:"output" short:"o" enum:"table,json" default:"table" help:"Output format (table, json)."`
	Debug   bool   `name:"debug" help:"Enable debug logging."`

	Login          cmd.LoginCmd          `cmd:"" help:"Authenticate via browser device flow (or --key to paste a key)."`
	Logout         cmd.LogoutCmd         `cmd:"" help:"Clear stored credentials."`
	Whoami         cmd.WhoamiCmd         `cmd:"" help:"Print the authenticated identity."`
	Version        cmd.VersionCmd        `cmd:"" help:"Print version."`
	Completion     cmd.CompletionCmd     `cmd:"" help:"Print shell completion script."`
	PermissionSets cmd.PermissionSetsCmd `cmd:"" name:"permission-sets" help:"Manage permission sets."`
	KeySpecs       cmd.KeySpecsCmd       `cmd:"" name:"key-specs" help:"Manage key specs."`
	Keys           cmd.KeysCmd           `cmd:"" help:"Manage issued keys."`
	SigningKeySets cmd.SigningKeySetsCmd `cmd:"" name:"signing-key-sets" help:"Manage signing key sets."`
	TrustExchanges cmd.TrustExchangesCmd `cmd:"" name:"trust-exchanges" help:"Manage trust exchanges."`
	TrustProviders cmd.TrustProvidersCmd `cmd:"" name:"trust-providers" help:"Manage trust providers."`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cli := CLI{}
	kctx := kong.Parse(&cli,
		kong.Name("microwave"),
		kong.Description("Microwave — API key, JWKS, and OIDC federation management."),
		kong.UsageOnError(),
		kong.Bind(ctx),
	)

	config.CheckForUpdate(version.Version)

	err := kctx.Run(&cmd.Globals{
		Token:   cli.Token,
		APIURL:  cli.APIURL,
		AuthURL: cli.AuthURL,
		Output:  cli.Output,
		Debug:   cli.Debug,
		Version: version.Version,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, cmd.RenderError(err))
		os.Exit(1)
	}
}
