package aso

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// KeywordsCommand returns the "keywords" subcommand with analyze, recommend, and batch.
func KeywordsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "keywords",
		ShortUsage: "aso keywords <subcommand> [flags]",
		ShortHelp:  "Analyze popularity, difficulty, and get AI-powered suggestions.",
		LongHelp: `Keyword intelligence commands powered by ASO Maniac.

Subcommands:
  analyze    Score keyword popularity, difficulty, and top-ranking apps
  recommend  Get AI-powered keyword suggestions from a seed
  batch      Submit an async multi-keyword analysis job (waits + streams results by default)
  jobs       Inspect or cancel async jobs`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			keywordsAnalyzeCommand(),
			keywordsRecommendCommand(),
			keywordsBatchCommand(),
			keywordsJobsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing subcommand. Run 'aso keywords --help' for usage")
			}
			return fmt.Errorf("unknown subcommand %q. Run 'aso keywords --help' for usage", args[0])
		},
	}
}

func keywordsAnalyzeCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords analyze", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	fields := fs.String("fields", "", "Comma-separated fields to request from server: popularity,difficulty,topApps,relatedSearches")
	exclude := fs.String("exclude", "", "Comma-separated fields to hide from output (e.g. topApps,relatedSearches)")
	retryPending := fs.Bool("retry-pending", false, "Auto-retry pending keywords once if the request times out")
	full := fs.Bool("full", false, "Print the full response object including meta on stdout")

	return &ffcli.Command{
		Name:       "analyze",
		ShortUsage: "aso keywords analyze <keyword> [<keyword>...] [flags]",
		ShortHelp:  "Score keyword popularity, difficulty, and top-ranking apps.",
		LongHelp: `Analyze one or more keywords for a given storefront. Returns popularity
score (0-100), difficulty score, competition data, and top-ranking apps.

When the request times out, a warning listing pending keywords is written to
stderr (stdout JSON is unaffected). Use --retry-pending to automatically issue
one follow-up request for the pending keywords and merge the results.

Examples:
  aso keywords analyze "photo editor"
  aso keywords analyze camera photo --storefront GB
  aso keywords analyze vpn --fields popularity,difficulty
  aso keywords analyze camera --exclude topApps,relatedSearches
  aso keywords analyze camera photo --retry-pending`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			keywords := resolveArgs(fs, args, true)
			if len(keywords) == 0 {
				return fmt.Errorf("at least one keyword is required")
			}
			var fieldSlice []string
			if *fields != "" {
				fieldSlice = strings.Split(*fields, ",")
			}
			return runKeywordsAnalyze(ctx, asomaniac.DefaultConfigPath(), keywords, *storefront, fieldSlice, parseExclude(*exclude), *retryPending, *full, os.Stdout, os.Stderr)
		},
	}
}

func runKeywordsAnalyze(ctx context.Context, configPath string, keywords []string, storefront string, fields, exclude []string, retryPending, full bool, stdout, stderr io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	resp, err := client.AnalyzeKeywords(ctx, keywords, storefront, fields)
	if err != nil {
		return fmt.Errorf("analyze keywords: %w", err)
	}

	// Optional retry of pending keywords.
	if retryPending && resp.Meta.TimedOut && len(resp.Meta.Pending) > 0 {
		retry, retryErr := client.AnalyzeKeywords(ctx, resp.Meta.Pending, storefront, fields)
		if retryErr != nil {
			fmt.Fprintf(stderr, "warning: retry for pending keywords failed: %v\n", retryErr)
		} else {
			resp.Data = append(resp.Data, retry.Data...)
			// Replace meta with the retry meta — it represents the final state.
			resp.Meta = retry.Meta
		}
	}

	if resp.Meta.TimedOut && len(resp.Meta.Pending) > 0 {
		fmt.Fprintf(stderr, "warning: %d keywords pending: %s. Re-run to fetch.\n",
			len(resp.Meta.Pending), strings.Join(resp.Meta.Pending, ", "))
	}

	if full {
		return writeJSON(stdout, resp, exclude)
	}
	return writeJSON(stdout, resp.Data, exclude)
}

func keywordsRecommendCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords recommend", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	limit := fs.Int("limit", 50, "Maximum number of recommendations")
	exclude := fs.String("exclude", "", "Comma-separated fields to hide from output")
	full := fs.Bool("full", false, "Print the full response object including meta on stdout")

	return &ffcli.Command{
		Name:       "recommend",
		ShortUsage: "aso keywords recommend <seed> [flags]",
		ShortHelp:  "Get AI-powered keyword suggestions from a seed.",
		LongHelp: `Generate keyword suggestions based on a seed keyword. Returns related
keywords ranked by popularity and difficulty.

When the request times out, a warning is written to stderr (stdout JSON is
unaffected).

Examples:
  aso keywords recommend "photo editor"
  aso keywords recommend camera --storefront GB --limit 25
  aso keywords recommend camera --exclude topApps`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			positional := resolveArgs(fs, args, false)
			if len(positional) == 0 {
				return fmt.Errorf("seed keyword is required")
			}
			return runKeywordsRecommend(ctx, asomaniac.DefaultConfigPath(), positional[0], *storefront, *limit, parseExclude(*exclude), *full, os.Stdout, os.Stderr)
		},
	}
}

func runKeywordsRecommend(ctx context.Context, configPath string, seed, storefront string, limit int, exclude []string, full bool, stdout, stderr io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	resp, err := client.GetRecommendations(ctx, seed, storefront, limit)
	if err != nil {
		return fmt.Errorf("get recommendations: %w", err)
	}

	if resp.Meta.TimedOut {
		fmt.Fprintf(stderr, "warning: recommendations request timed out after %dms; partial results returned.\n", resp.Meta.ElapsedMs)
	}

	if full {
		return writeJSON(stdout, resp, exclude)
	}
	return writeJSON(stdout, resp.Data, exclude)
}

func keywordsBatchCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords batch", flag.ContinueOnError)
	storefronts := fs.String("storefronts", "US", "Comma-separated storefront codes")
	exclude := fs.String("exclude", "", "Comma-separated fields to hide from output")
	noWait := fs.Bool("no-wait", false, "Submit the job and exit; print {jobId, statusUrl} only")
	jsonOnly := fs.Bool("json", false, "Suppress progress output; print only the final JSON summary on stdout")
	pollInterval := fs.Duration("poll-interval", 2*time.Second, "How often to poll job status (default 2s)")
	timeout := fs.Duration("timeout", 30*time.Minute, "Max wall time before giving up on the job")

	return &ffcli.Command{
		Name:       "batch",
		ShortUsage: "aso keywords batch <keyword> [<keyword>...] --storefronts US,GB,DE",
		ShortHelp:  "Submit an async multi-keyword analysis job (waits + streams results).",
		LongHelp: `Batch-analyze keywords across one or more storefronts.

The submission is async: every (term, storefront) pair becomes a JobItem on
the server, the worker drains them serially (Apple Ads /recommendation is
~1 req/sec global), and this command polls until the job is COMPLETED.

By default, completed items stream to stdout as JSON-Lines (one JSON object
per line). Progress (e.g. "[24/100] processing spy camera") goes to stderr.
Final exit code is 0 on COMPLETED, 1 on FAILED/CANCELLED.

Use --no-wait to submit and exit without polling.
Use --json to suppress progress and print only the final summary on stdout.

Examples:
  aso keywords batch camera photo vpn
  aso keywords batch "photo editor" "video editor" --storefronts US,GB,DE
  aso keywords batch camera photo --no-wait                 # fire-and-forget
  aso keywords batch camera photo --json | jq               # machine mode`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			keywords := resolveArgs(fs, args, true)
			if len(keywords) == 0 {
				return fmt.Errorf("at least one keyword is required")
			}
			sfList := strings.Split(*storefronts, ",")
			return runKeywordsBatch(ctx, batchOptions{
				configPath:   asomaniac.DefaultConfigPath(),
				keywords:     keywords,
				storefronts:  sfList,
				exclude:      parseExclude(*exclude),
				noWait:       *noWait,
				jsonOnly:     *jsonOnly,
				pollInterval: *pollInterval,
				timeout:      *timeout,
				stdout:       os.Stdout,
				stderr:       os.Stderr,
			})
		},
	}
}

type batchOptions struct {
	configPath   string
	keywords     []string
	storefronts  []string
	exclude      []string
	noWait       bool
	jsonOnly     bool
	pollInterval time.Duration
	timeout      time.Duration
	stdout       io.Writer
	stderr       io.Writer
}

func runKeywordsBatch(ctx context.Context, opt batchOptions) error {
	client, err := requireAuth(opt.configPath)
	if err != nil {
		return err
	}

	submission, err := client.SubmitBatchAnalyze(ctx, opt.keywords, opt.storefronts)
	if err != nil {
		return fmt.Errorf("submit job: %w", err)
	}

	if opt.noWait {
		return writeJSON(opt.stdout, submission, opt.exclude)
	}

	if !opt.jsonOnly {
		fmt.Fprintf(opt.stderr, "submitted job %s (%d items)\n", submission.JobID, submission.TotalCount)
	}

	deadline := time.Now().Add(opt.timeout)
	streamed := make(map[string]struct{}) // (term::storefront) keys we've already emitted
	pollEvery := opt.pollInterval
	const maxPollInterval = 5 * time.Second
	lastProgress := -1
	stagnantSince := time.Time{}

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("job %s did not complete within %s", submission.JobID, opt.timeout)
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		job, err := client.GetJob(ctx, submission.JobID)
		if err != nil {
			return fmt.Errorf("poll job: %w", err)
		}

		// Stream newly-completed items to stdout.
		if !opt.jsonOnly {
			for _, it := range job.Items {
				if it.Status != "DONE" || it.Result == nil {
					continue
				}
				key := it.Term + "::" + it.Storefront
				if _, seen := streamed[key]; seen {
					continue
				}
				streamed[key] = struct{}{}
				if err := writeJSONL(opt.stdout, it.Result, opt.exclude); err != nil {
					return err
				}
			}
		}

		if !opt.jsonOnly {
			// Progress chip on stderr. Show next pending item if any.
			next := nextPending(job)
			if next != "" {
				fmt.Fprintf(opt.stderr, "\r[%d/%d] %s            ", job.ProcessedCount, job.TotalCount, next)
			} else {
				fmt.Fprintf(opt.stderr, "\r[%d/%d]            ", job.ProcessedCount, job.TotalCount)
			}
		}

		// Backoff: if processedCount hasn't advanced for 60s, slow polling to 5s.
		if job.ProcessedCount != lastProgress {
			lastProgress = job.ProcessedCount
			stagnantSince = time.Time{}
		} else if stagnantSince.IsZero() {
			stagnantSince = time.Now()
		} else if time.Since(stagnantSince) > 60*time.Second && pollEvery < maxPollInterval {
			pollEvery = maxPollInterval
		}

		switch job.Status {
		case "COMPLETED":
			if !opt.jsonOnly {
				fmt.Fprintln(opt.stderr) // newline after progress chip
				fmt.Fprintf(opt.stderr, "job %s completed (%d/%d)\n", job.ID, job.ProcessedCount, job.TotalCount)
			} else {
				return writeJSON(opt.stdout, job, opt.exclude)
			}
			return nil
		case "FAILED":
			if !opt.jsonOnly {
				fmt.Fprintln(opt.stderr)
			}
			msg := ""
			if job.Error != nil {
				msg = *job.Error
			}
			return fmt.Errorf("job %s failed: %s", job.ID, msg)
		case "CANCELLED":
			if !opt.jsonOnly {
				fmt.Fprintln(opt.stderr)
			}
			return fmt.Errorf("job %s cancelled", job.ID)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollEvery):
		}
	}
}

func nextPending(job *asomaniac.JobPollResponse) string {
	for _, it := range job.Items {
		if it.Status == "RUNNING" {
			return fmt.Sprintf("processing %q (%s)", it.Term, it.Storefront)
		}
	}
	for _, it := range job.Items {
		if it.Status == "PENDING" {
			return fmt.Sprintf("queued %q (%s)", it.Term, it.Storefront)
		}
	}
	return ""
}

// writeJSONL emits one JSON document per line. Used to stream completed items
// during async batch runs so AI agents can pipe and parse incrementally.
func writeJSONL(w io.Writer, v any, exclude []string) error {
	data, err := jsonMarshalExcluding(v, exclude)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	return nil
}

// jsonMarshalExcluding emits compact JSON with the named fields elided.
// Mirrors the visibility filter applied by writeJSON but produces a single
// line (no trailing newline) suitable for JSONL streaming.
func jsonMarshalExcluding(v any, exclude []string) ([]byte, error) {
	if len(exclude) == 0 {
		return json.Marshal(v)
	}
	// Round-trip through map to drop fields by tag-derived JSON keys.
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	stripped := stripFields(m, exclude)
	return json.Marshal(stripped)
}

func stripFields(v any, exclude []string) any {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range exclude {
			delete(t, k)
		}
		for k, vv := range t {
			t[k] = stripFields(vv, exclude)
		}
		return t
	case []any:
		for i, vv := range t {
			t[i] = stripFields(vv, exclude)
		}
		return t
	default:
		return v
	}
}

// ─── jobs subcommands ───────────────────────────────────────────────────────

func keywordsJobsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords jobs", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "jobs",
		ShortUsage: "aso keywords jobs <subcommand> | <id>",
		ShortHelp:  "Inspect or cancel async keyword analysis jobs.",
		LongHelp: `Examples:
  aso keywords jobs <id>            # show job status + items
  aso keywords jobs cancel <id>     # cancel a running job
  aso keywords jobs                 # list recent jobs (paginated)`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			keywordsJobsCancelCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			cli, err := requireAuth(asomaniac.DefaultConfigPath())
			if err != nil {
				return err
			}
			if len(args) == 0 {
				// TODO: list recent jobs once /jobs index endpoint UX firms up.
				return fmt.Errorf("usage: aso keywords jobs <id>  |  aso keywords jobs cancel <id>")
			}
			jobID := args[0]
			job, err := cli.GetJob(ctx, jobID)
			if err != nil {
				return fmt.Errorf("get job: %w", err)
			}
			return writeJSON(os.Stdout, job, nil)
		},
	}
}

func keywordsJobsCancelCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords jobs cancel", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "cancel",
		ShortUsage: "aso keywords jobs cancel <id>",
		ShortHelp:  "Cancel a running async job.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("job id required")
			}
			cli, err := requireAuth(asomaniac.DefaultConfigPath())
			if err != nil {
				return err
			}
			if err := cli.CancelJob(ctx, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "cancelled job %s\n", args[0])
			return nil
		},
	}
}
