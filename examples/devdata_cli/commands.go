// Copyright 2022 Searis AS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/jsonrpc"
	"github.com/clarify/clarify-go/query"
	"github.com/clarify/clarify-go/views"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
)

const (
	defaultSignalPrefix     = "devdata/signal-"
	defaultAnnotationPrefix = "clarify/clarify-go/examples/devdata_cli"

	defaultTransformVersion = "v1"
)

func rootCommand() *ffcli.Command {
	var p program
	fs := flag.NewFlagSet("devdata_cli", flag.ExitOnError)
	fs.StringVar(&p.credentialsFile, "credentials", "clarify-credentials.json", "Clarify credentials file location.")
	fs.BoolVar(&p.logRequests, "log-requests", false, "Log all RPC request, including trace information.")
	fs.BoolVar(&p.logOnly, "log-only", false, "Disable stdout content.")

	return &ffcli.Command{
		ShortUsage: "devdata_cli [flags] <subcommand>",
		ShortHelp:  "An example CLI on top of the Clarify SDK, allowing generation and selection of data for testing and experimentation purposes.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			return fmt.Errorf("subcommand required; try -help")
		},
		Options: []ff.Option{
			ff.WithConfigFileFlag("config"),
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithAllowMissingConfigFile(true),
			ff.WithEnvVarPrefix("DEVDATA"),
		},
		Subcommands: []*ffcli.Command{
			p.insertCommand(),
			p.saveSignalsCommand(),
			p.selectSignalsCommand(),
			p.selectItemsCommand(),
			p.publishSignalsCommand(),
			p.dataFrameCommand(),
		},
	}
}

type program struct {
	// raw parameters.
	credentialsFile string
	logRequests     bool
	logOnly         bool

	// runtime variables.
	defaultIntegration string
	client             *clarify.Client
	stdout             io.Writer
}

func (p program) Println(a ...any) (int, error) {
	return fmt.Fprintln(p.stdout, a...)
}

func (p program) Printf(format string, a ...any) (int, error) {
	return fmt.Fprintf(p.stdout, format, a...)
}

func (p program) EncodeJSON(data any) error {
	enc := json.NewEncoder(p.stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return err
	}
	return nil
}

func (p *program) init(ctx context.Context) {
	creds, err := clarify.CredentialsFromFile(p.credentialsFile)
	if err != nil {
		log.Println("fatal:", err)
		os.Exit(2)
	}
	h, err := creds.HTTPHandler(ctx)
	if err != nil {
		log.Println("fatal:", err)
		os.Exit(2)
	}
	if p.logRequests {
		h.RequestLogger = func(req jsonrpc.Request, trace string, latency time.Duration, err error) {
			log.Printf("JSONRPC request: %s, trace: %s, latency: %s, error: %v", req.Method, trace, latency, err)
		}
	}
	if p.logOnly {
		p.stdout = nilWriter{}
	} else {
		p.stdout = os.Stdout
	}

	p.defaultIntegration = creds.Integration
	p.client = clarify.NewClient(creds.Integration, h)
}

type insertConfig struct {
	signalPrefix, annotationPrefix string
	start, count, points           int

	publish   bool
	startTime time.Time
	truncate  time.Duration
	interval  time.Duration
}

func (p *program) insertCommand() *ffcli.Command {
	var config insertConfig
	config.startTime = time.Now().Truncate(time.Second) // set default start-time.

	fs := flag.NewFlagSet("devdata_cli insert", flag.ExitOnError)
	fs.StringVar(&config.signalPrefix, "prefix", defaultSignalPrefix, "Prefix to use for signal input keys.")
	fs.StringVar(&config.annotationPrefix, "annotation-prefix", defaultAnnotationPrefix, "Prefix to use annotations.")
	fs.IntVar(&config.start, "s", 0, "Number to use for first signal input key.")
	fs.IntVar(&config.count, "n", 1, "Number of signals to insert.")
	fs.IntVar(&config.points, "points", 1, "Number of data-points to insert per signal.")
	fs.BoolVar(&config.publish, "publish", false, "Update signal meta-data and publish signals.")
	fs.Var(timeFlag{target: &config.startTime}, "start-time", "first RFC 3339 timestamp to insert data for (default is current time).")
	fs.DurationVar(&config.truncate, "truncate-start-time", 1*time.Second, "Truncate the start time to align to this duration.")
	fs.DurationVar(&config.interval, "sample-interval", 1*time.Second, "Distance between data-points when points > 1; negative values inserts values backwards in time from start-time.")

	return &ffcli.Command{
		Name:       "insert",
		ShortUsage: "devdata_cli insert [flags]",
		ShortHelp:  "Insert random data into Clarify in range 0-100. You can customize how much data to insert.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			p.init(ctx)
			if err := p.insert(ctx, config); err != nil {
				return err
			}
			if !config.publish {
				return nil
			}
			saveConfig := saveSignalsConfig{
				signalPrefix:     config.signalPrefix,
				annotationPrefix: config.annotationPrefix,
				start:            config.start,
				count:            config.count,
				enablePublish:    true,
			}
			if err := p.saveSignals(ctx, saveConfig); err != nil {
				return err
			}
			publishConfig := publishSignalsConfig{
				signalPrefix:        config.signalPrefix,
				annotationPrefix:    config.annotationPrefix,
				annotationTransform: defaultAnnotationPrefix,
			}
			if err := p.publishSignals(ctx, publishConfig); err != nil {
				return err
			}
			return nil
		},
	}
}

func (p *program) insert(ctx context.Context, config insertConfig) error {
	if config.count < 1 {
		return fmt.Errorf("-n can not be below 1")
	}
	if config.points < 1 {
		return fmt.Errorf("-points can not be below 1")
	}
	rand.Seed(time.Now().UnixNano())

	startTime := config.startTime.Truncate(config.truncate)
	endTime := startTime.Add(config.interval * time.Duration(config.points))
	var compare func(t time.Time) bool
	switch {
	case config.interval == 0:
		return fmt.Errorf("-sample-interval can not be 0")
	case config.interval > 0:
		compare = func(t time.Time) bool { return t.Before(endTime) }
	default:
		compare = func(t time.Time) bool { return t.After(endTime) }
	}

	startSignal := config.start
	endSignal := config.start + config.count

	df := make(views.DataFrame, config.count)
	for i := startSignal; i < endSignal; i++ {
		k := config.signalPrefix + strconv.Itoa(i)
		ds := make(views.DataSeries, config.points)
		for t := startTime; compare(t); t = t.Add(config.interval) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			ds[fields.AsTimestamp(t)] = rand.Float64() * 100
		}
		df[k] = ds
	}
	log.Printf("Inserting data frame with %d signals and %d points per signal.", config.count, config.points)
	result, err := p.client.Insert(df).Do(ctx)
	if err != nil {
		return err
	}
	var created int
	for _, v := range result.SignalsByInput {
		if v.Created {
			created++
		}
	}
	log.Printf("Save summary: created: %d, total: %d", created, len(result.SignalsByInput))

	return p.EncodeJSON(result)
}

type saveSignalsConfig struct {
	signalPrefix, annotationPrefix string
	start, count                   int
	enablePublish, disablePublish  bool
}

func (p *program) saveSignalsCommand() *ffcli.Command {
	var config saveSignalsConfig
	fs := flag.NewFlagSet("devdata_cli save-signals", flag.ExitOnError)
	fs.StringVar(&config.signalPrefix, "prefix", defaultSignalPrefix, "Prefix to use for signal input keys.")
	fs.StringVar(&config.annotationPrefix, "annotation-prefix", defaultAnnotationPrefix, "Prefix to use annotations.")
	fs.IntVar(&config.start, "s", 0, "Number to use for first signal input key.")
	fs.IntVar(&config.count, "n", 1, "Number of signals to update.")
	fs.BoolVar(&config.enablePublish, "enable-publish", false, "Annotate the signal to allow publish (default is no change).")
	fs.BoolVar(&config.disablePublish, "disable-publish", false, "Annotate the signal to disallow publish (default is no change).")

	return &ffcli.Command{
		Name:       "save-signals",
		ShortUsage: "devdata_cli save-signals [flags]",
		ShortHelp:  "Provide signal meta-data (in this case, annotations).",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			p.init(ctx)
			return p.saveSignals(ctx, config)
		},
	}
}

func (p *program) saveSignals(ctx context.Context, config saveSignalsConfig) error {
	keyPublish := config.annotationPrefix + "publish"
	annotationsPatch := fields.Annotations{}
	if config.enablePublish {
		annotationsPatch[keyPublish] = "true"
	}
	if config.disablePublish {
		annotationsPatch[keyPublish] = ""
	}
	startSignal := config.start
	endSignal := config.start + config.count

	inputs := make(map[string]views.SignalSave, config.count)
	for i := startSignal; i < endSignal; i++ {
		k := config.signalPrefix + strconv.Itoa(i)
		inputs[k] = views.SignalSave{
			MetaSave: views.MetaSave{
				Annotations: annotationsPatch,
			},
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name: k,
			},
		}
	}
	log.Printf("Updating meta-data for %d signals.", config.count)
	result, err := p.client.SaveSignals(inputs).Do(ctx)
	if err != nil {
		return err
	}

	var created, updated int
	for _, v := range result.SignalsByInput {
		if v.Created {
			created++
		}
		if v.Updated {
			updated++
		}
	}
	log.Printf("Save summary: created: %d, updated: %d, total: %d", created, updated, len(result.SignalsByInput))

	return p.EncodeJSON(result)
}

type selectSignalsConfig struct {
	skip, limit         int
	integration, filter string
	includeItem         bool
	sort                []string
}

func (p *program) selectSignalsCommand() *ffcli.Command {
	var config selectSignalsConfig

	fs := flag.NewFlagSet("devdata_cli select-signals", flag.ExitOnError)
	fs.StringVar(&config.integration, "integration", "", "Integration to list signals for (defaults to integration from credentials file).")
	fs.IntVar(&config.skip, "s", 0, "Number of signals to skip.")
	fs.IntVar(&config.limit, "n", 50, "Maximum number of signals to return.")
	fs.StringVar(&config.filter, "filter", "", "Resource filter (JSON).")
	fs.Var(stringSlice{target: &config.sort}, "sort", "Comma-separated list of fields to sort the result by.")
	fs.BoolVar(&config.includeItem, "include-item", false, "Include related items.")

	return &ffcli.Command{
		Name:       "select-signals",
		ShortUsage: "devdata_cli select-signals [flags]",
		ShortHelp:  "List signal resources for a particular integration.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			p.init(ctx)
			return p.selectSignals(ctx, config)
		},
	}
}

func (p *program) selectSignals(ctx context.Context, config selectSignalsConfig) error {
	if config.integration == "" {
		config.integration = p.defaultIntegration
	}

	req := p.client.SelectSignals(config.integration).Skip(config.skip).Limit(config.limit)
	if config.includeItem {
		req = req.Include("item")
	}
	if len(config.sort) > 0 {
		req = req.Sort(config.sort...)
	}
	if config.filter != "" {
		var f query.Filter
		if err := json.Unmarshal([]byte(config.filter), &f); err != nil {
			return fmt.Errorf("-filter: %w", err)
		}
		req = req.Filter(f)
	}
	result, err := req.Do(ctx)
	if err != nil {
		return err
	}

	log.Printf("Selection summary: len(result.data): %d, len(result.included.items): %d", len(result.Data), len(result.Included.Items))
	return p.EncodeJSON(result)
}

type publishSignalsConfig struct {
	signalPrefix, annotationPrefix, annotationTransform string

	integration, filter string
	skip, limit         int
}

func (p *program) publishSignalsCommand() *ffcli.Command {
	var config publishSignalsConfig

	fs := flag.NewFlagSet("devdata_cli publish-signals", flag.ExitOnError)
	fs.StringVar(&config.signalPrefix, "prefix", defaultSignalPrefix, "Prefix to use for signal input keys.")
	fs.StringVar(&config.annotationPrefix, "annotation-prefix", defaultAnnotationPrefix, "Prefix to use for annotations.")
	fs.StringVar(&config.annotationTransform, "annotation-transform", defaultTransformVersion, "Value to search for in the name annotation.")
	fs.StringVar(&config.integration, "integration", "", "Integration to publish signals from (defaults to integration from credentials file).")
	fs.StringVar(&config.filter, "filter", "", "Resource filter (JSON) that's combined with the a forced annotations filter for publish=true.")
	fs.IntVar(&config.skip, "s", 0, "Number of signals to skip.")
	fs.IntVar(&config.limit, "n", 100, "Maximum number of signals to publish.")

	return &ffcli.Command{
		Name:       "publish-signals",
		ShortUsage: "devdata_cli publish-signals [flags]",
		ShortHelp:  "(Re-)publish signals matching the search criteria when signal meta-data or transform version is changed.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			p.init(ctx)

			return p.publishSignals(ctx, config)
		},
	}
}

func (p *program) publishSignals(ctx context.Context, config publishSignalsConfig) error {
	if config.integration == "" {
		config.integration = p.defaultIntegration
	}
	keyPublish := config.annotationPrefix + "publish"
	keyTransformVersion := config.annotationPrefix + "source/transform"
	keySignalAttributesHash := config.annotationPrefix + "source/signal-attributes-hash"
	keySignalID := config.annotationPrefix + "source/signal-id"

	selectReq := p.client.SelectSignals(config.integration).
		Skip(config.skip).
		Limit(config.limit).
		Filter(query.Comparisons{"annotations." + keyPublish: query.Equal("true")}).
		Include("item")
	if config.filter != "" {
		var f query.Filter
		if err := json.Unmarshal([]byte(config.filter), &f); err != nil {
			return fmt.Errorf("-filter: %w", err)
		}
		selectReq = selectReq.Filter(f)
	}
	selectResult, err := selectReq.Do(ctx)
	if err != nil {
		panic(err)
	}

	// This logic only updates items if the signals has changed since last time
	// the item was published, or if our transform version has changed.
	prevItemsBySignal := make(map[string]views.Item, len(selectResult.Included.Items))
	for _, item := range selectResult.Included.Items {
		prevItemsBySignal[item.Meta.Annotations.Get(keySignalID)] = item
	}

	newItems := make(map[string]views.ItemSave, len(selectResult.Data))

	for _, signal := range selectResult.Data {
		// Only update item if the source signal has changed.
		prevItem, ok := prevItemsBySignal[signal.ID]
		ok = ok && prevItem.Meta.Annotations.Get(keyTransformVersion) == defaultTransformVersion
		ok = ok && prevItem.Meta.Annotations.Get(keySignalAttributesHash) == signal.Meta.AttributesHash.String()
		if ok {
			log.Printf("Skip signal %s: item is up-to-date", signal.ID)
			continue
		}

		// Use visible status from previous item.
		visible := prevItem.Attributes.Visible
		if prevItem.ID == "" {
			visible = true
		}

		newItems[signal.ID] = views.PublishedItem(signal,
			func(item *views.ItemSave) {
				item.Annotations.Set(keyTransformVersion, defaultTransformVersion)
				item.Annotations.Set(keySignalAttributesHash, signal.Meta.AttributesHash.String())
				item.Annotations.Set(keySignalID, signal.ID)
				item.Visible = visible
			},
		)
	}

	if len(newItems) == 0 {
		log.Printf("All items up-to-date")
		return nil
	}
	log.Printf("Updating %d/%d items", len(newItems), len(selectResult.Data))
	result, err := p.client.PublishSignals(config.integration, newItems).Do(ctx)
	if err != nil {
		return err
	}

	var created, updated int
	for _, v := range result.ItemsBySignals {
		if v.Created {
			created++
		}
		if v.Updated {
			updated++
		}
	}
	log.Printf("Save summary: created: %d, updated: %d, total: %d", created, updated, len(result.ItemsBySignals))
	return p.EncodeJSON(result)
}

type selectItemsConfig struct {
	skip, limit int
	filter      string
	sort        []string
}

func (p *program) selectItemsCommand() *ffcli.Command {
	var config selectItemsConfig

	fs := flag.NewFlagSet("devdata_cli select-items", flag.ExitOnError)
	fs.IntVar(&config.skip, "s", 0, "Number of items to skip.")
	fs.IntVar(&config.limit, "n", 50, "Maximum number of items to return.")
	fs.StringVar(&config.filter, "filter", "", "Resource filter (JSON).")
	fs.Var(stringSlice{target: &config.sort}, "sort", "Comma-separated list of fields to sort the result by.")

	return &ffcli.Command{
		Name:       "select-items",
		ShortUsage: "devdata_cli select-items [flags]",
		ShortHelp:  "List item resources for organization.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			p.init(ctx)
			return p.selectItems(ctx, config)
		},
	}
}

func (p *program) selectItems(ctx context.Context, config selectItemsConfig) error {
	req := p.client.SelectItems().Skip(config.skip).Limit(config.limit)
	if config.filter != "" {
		var f query.Filter
		if err := json.Unmarshal([]byte(config.filter), &f); err != nil {
			return fmt.Errorf("-filter: %w", err)
		}
		req = req.Filter(f)
	}
	if len(config.sort) > 0 {
		req = req.Sort(config.sort...)
	}

	result, err := req.Do(ctx)
	if err != nil {
		return err
	}

	log.Printf("Selection summary: len(result.data): %d", len(result.Data))
	return p.EncodeJSON(result)
}

type dataFrameConfig struct {
	skip, limit int
	filter      string
	includeItem bool
	sort        []string

	startTime, endTime time.Time
	rollupDuration     time.Duration
	rollupWindow       bool
	last               int
}

func (p *program) dataFrameCommand() *ffcli.Command {
	var config dataFrameConfig

	// Set default-times.
	config.startTime = time.Now().Truncate(24 * time.Hour)
	config.endTime = config.startTime.Add(24 * time.Hour)

	fs := flag.NewFlagSet("devdata_cli data-frame", flag.ExitOnError)
	fs.IntVar(&config.skip, "s", 0, "Number of items to skip.")
	fs.IntVar(&config.limit, "n", 50, "Maximum number of items to return.")
	fs.IntVar(&config.last, "last", -1, "Limit response to last N samples.")
	fs.StringVar(&config.filter, "filter", "", "Resource filter (JSON).")
	fs.Var(stringSlice{target: &config.sort}, "sort", "Comma-separated list of fields to sort the result by.")
	fs.BoolVar(&config.includeItem, "include-item", false, "Include related items.")
	fs.Var(timeFlag{target: &config.startTime}, "start-time", "RFC 3339 timestamp of first data-point to include.")
	fs.Var(timeFlag{target: &config.endTime}, "end-time", "RFC 3339 timestamp of first data-point not to include.")
	fs.BoolVar(&config.rollupWindow, "rollup-window", false, "Use window rollup.")
	fs.DurationVar(&config.rollupDuration, "rollup-duration", 0, "Set to positive duration to enable time-bucket rollup.")

	return &ffcli.Command{
		Name:       "data-frame",
		ShortUsage: "devdata_cli data-frame [flags]",
		ShortHelp:  "Return a data frame.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			p.init(ctx)

			return p.dataFrame(ctx, config)
		},
	}
}

func (p *program) dataFrame(ctx context.Context, config dataFrameConfig) error {
	req := p.client.DataFrame().Include("item").Skip(config.skip).Limit(config.limit).TimeRange(config.startTime, config.endTime)
	expectSeriesPerItem := 1
	switch {
	case config.rollupWindow:
		expectSeriesPerItem = 5
		req = req.RollupWindow()
	case config.rollupDuration > 0:
		expectSeriesPerItem = 5
		req = req.RollupBucket(config.rollupDuration)
	}
	if config.last > 0 {
		req = req.Last(config.last)
	}
	if config.includeItem {
		req = req.Include("item")
	}
	if config.filter != "" {
		var f query.Filter
		if err := json.Unmarshal([]byte(config.filter), &f); err != nil {
			return fmt.Errorf("-filter: %w", err)
		}
		req = req.Filter(f)
	}
	if len(config.sort) > 0 {
		req = req.Sort(config.sort...)
	}

	result, err := req.Do(ctx)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(result.Data))
	for k := range result.Data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	var i, samples int
	for _, k := range keys {

		data := result.Data[k]
		samples += len(data)

		if config.includeItem {
			id, _, _ := strings.Cut(k, "_")
			for id != result.Included.Items[i].ID {
				i++
			}
			log.Printf("len(result.data['%s']): %d, name: %s\n", k, len(data), result.Included.Items[i].Attributes.Name)
		} else {
			log.Printf("len(result.data['%s']): %d\n", k, len(data))
		}
	}

	if config.includeItem {
		missing := len(result.Included.Items)*expectSeriesPerItem - len(result.Data)
		log.Printf("Data frame summary: len(result.data): %d, len(result.included.items): %d, total samples: %d, missing series: %d", len(result.Data), len(result.Included.Items), samples, missing)
	} else {
		log.Printf("Data frame summary: len(result.data): %d, len(result.included.items): %d, total samples: %d", len(result.Data), len(result.Included.Items), samples)
	}
	return p.EncodeJSON(result)
}
