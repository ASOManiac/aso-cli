package certificates

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// CertificatesRelationshipsCommand returns the links command group.
func CertificatesRelationshipsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("links", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "links",
		ShortUsage: "aso certificates links <subcommand> [flags]",
		ShortHelp:  "View certificate relationship linkages.",
		LongHelp: `View certificate relationship linkages.

Examples:
  aso certificates links pass-type-id --id "CERT_ID"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			CertificatesRelationshipsPassTypeIDCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// CertificatesRelationshipsPassTypeIDCommand returns the pass-type-id links command.
func CertificatesRelationshipsPassTypeIDCommand() *ffcli.Command {
	fs := flag.NewFlagSet("pass-type-id", flag.ExitOnError)

	id := fs.String("id", "", "Certificate ID")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "pass-type-id",
		ShortUsage: "aso certificates links pass-type-id --id \"CERT_ID\"",
		ShortHelp:  "Get pass type ID relationship for a certificate.",
		LongHelp: `Get pass type ID relationship for a certificate.

Examples:
  aso certificates links pass-type-id --id "CERT_ID"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*id)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := shared.GetASCClient()
			if err != nil {
				return fmt.Errorf("certificates links pass-type-id: %w", err)
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetCertificatePassTypeIDRelationship(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("certificates links pass-type-id: failed to fetch: %w", err)
			}

			return shared.PrintOutput(resp, *output.Output, *output.Pretty)
		},
	}
}

// DeprecatedCertificatesRelationshipsAliasCommand preserves the legacy
// relationships surface as a hidden compatibility alias.
func DeprecatedCertificatesRelationshipsAliasCommand() *ffcli.Command {
	fs := flag.NewFlagSet("relationships", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "relationships",
		ShortUsage: "aso certificates links <subcommand> [flags]",
		ShortHelp:  "DEPRECATED: use `aso certificates links ...`.",
		LongHelp:   "Deprecated compatibility alias for `aso certificates links ...`.",
		FlagSet:    fs,
		UsageFunc:  shared.DeprecatedUsageFunc,
		Subcommands: []*ffcli.Command{
			shared.DeprecatedAliasLeafCommand(
				CertificatesRelationshipsPassTypeIDCommand(),
				"pass-type-id",
				"aso certificates links pass-type-id --id \"CERT_ID\"",
				"aso certificates links pass-type-id",
				"Warning: `aso certificates relationships pass-type-id` is deprecated. Use `aso certificates links pass-type-id`.",
			),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
