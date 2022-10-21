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

package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/query"
	"github.com/clarify/clarify-go/views"
	"golang.org/x/exp/maps"
)

const (
	selectSignalsPageSize  = 1000
	publishSignalsPageSize = 500
)

// Annotations used by automation tasks.
const (
	AnnotationPrefix                    = "clarify/clarify-go/"
	AnnotationPublisherName             = AnnotationPrefix + "publisher/name"
	AnnotationPublisherTransformVersion = AnnotationPrefix + "publisher/transform-version"
	AnnotationPublisherSignalID         = AnnotationPrefix + "publisher/signal-id"
	AnnotationPublisherSignalAttributes = AnnotationPrefix + "publisher/signal-attributes"
)

// LogOptions describe the options for operation logs.
type LogOptions struct {
	// Verbose, when true, enables detailed logs, such as full JSON summaries of
	// operations.
	Verbose bool

	// Out describe the destination writer for logs. If unset, all logging is
	// disabled.
	Out io.Writer
}

func (opts LogOptions) EncodeJSON(v any) {
	if opts.Out == nil {
		return
	}
	enc := json.NewEncoder(opts.Out)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func (opts LogOptions) Printf(format string, a ...any) {
	if opts.Out == nil {
		return
	}
	fmt.Fprintf(opts.Out, format, a...)
}

// PublishOptions describe options that can be supplied when running a
// PublishRuleSet.
type PublishOptions struct {
	LogOptions

	// DryRun, if true, does not run the automation, but rather logs what it's
	// planning to do.
	DryRun bool

	// Publisher is a name describing the publisher application. The default is
	// the declared path of the main module.
	Publisher string
}

func (opts PublishOptions) withDefaults() PublishOptions {
	if opts.Publisher == "" {
		info, ok := debug.ReadBuildInfo()
		if ok {
			opts.Publisher = info.Main.Path
		}
	}
	return opts
}

// PublishSignals allows you to automate signal publishing from one or more
// source integrations.
type PublishSignals struct {
	// Integrations must list the IDs of the integrations to publish signals
	// from. If this list is empty, the rule set is a no-op.
	Integrations []string

	// SignalsFilter can optionally be specified to limit which signals to
	// publish.
	SignalsFilter query.FilterType

	// TransformVersion should be changed if you want existing items to be
	// republished despite the source signal being unchanged.
	TransformVersion string

	// Transforms is a list of transforms to apply when publishing the signals.
	// The transforms are applied in order.
	Transforms []func(item *views.ItemSave)
}

// Do performs the automation against c with the passed in opts.
func (p PublishSignals) Do(ctx context.Context, c *clarify.Client, opts PublishOptions) error {
	var total int
	defer func() {
		var suffix string
		if opts.DryRun {
			suffix = " (dry-run)"
		}
		opts.Printf("-- Published %d signals from %d integrations%s.\n", total, len(p.Integrations), suffix)
	}()

	if err := ctx.Err(); err != nil {
		return err
	}

	opts = opts.withDefaults()

	// We iterate signals without requesting the total count. This is an
	// optimization bet that total % limit == 0 is uncommon.
	q := query.Query{
		Sort:  []string{"id"},
		Limit: selectSignalsPageSize,
	}
	if p.SignalsFilter != nil {
		q.Filter = p.SignalsFilter.Filter()
	}

	items := make(map[string]views.ItemSave)
	flush := func(integrationID string) error {
		var suffix string
		if opts.DryRun {
			suffix = " (dry-run)"
		}
		opts.Printf("Publish %d items%s...\n", len(items), suffix)
		if opts.Verbose {
			opts.Printf("itemsBySignal:\n")
			opts.EncodeJSON(items)
		}
		if !opts.DryRun {
			result, err := c.PublishSignals(integrationID, items).Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to publish signals: %w", err)
			}
			if opts.Verbose {
				opts.Printf("result:\n")
				opts.EncodeJSON(result)
			}
		}

		total += len(items)
		maps.Clear(items)
		return nil
	}

	for _, id := range p.Integrations {
		q.Skip = 0
		more := true
		for more {
			if err := ctx.Err(); err != nil {
				return err
			}

			var err error
			more, err = p.addItems(ctx, items, c, id, q, opts)
			if err != nil {
				return err
			}
			q.Skip += q.Limit

			if len(items) >= publishSignalsPageSize {
				if err := flush(id); err != nil {
					return err
				}
			}
		}

		if err := flush(id); err != nil {
			return err
		}
	}

	return nil
}

// addItems adds items that require update to dest from all signals matching
// the integration ID and query q.
func (p PublishSignals) addItems(ctx context.Context, dest map[string]views.ItemSave, c *clarify.Client, integrationID string, q query.Query, opts PublishOptions) (bool, error) {
	results, err := c.SelectSignals(integrationID).
		Query(q).
		Include("item").Do(ctx)
	if err != nil {
		return false, err
	}

	// This logic only updates items if the signals has changed since last time
	// the item was published, or if our transform version has changed.
	prevItemsBySignal := make(map[string]views.Item, len(results.Included.Items))
	for _, item := range results.Included.Items {
		prevItemsBySignal[item.Meta.Annotations.Get(AnnotationPublisherSignalID)] = item
	}

	for _, signal := range results.Data {
		// Only update item if either the source signal or transform version
		// have changed.
		prevItem, ok := prevItemsBySignal[signal.ID]
		ok = ok && prevItem.Meta.Annotations.Get(AnnotationPublisherTransformVersion) == p.TransformVersion
		ok = ok && prevItem.Meta.Annotations.Get(AnnotationPublisherSignalAttributes) == signal.Meta.AttributesHash.String()
		if ok {
			if opts.Verbose {
				opts.Printf("Skip signal %s: item is up-to-date\n", signal.ID)
			}
			continue
		}

		item := views.PublishedItem(signal)
		// Before running configured transforms, set the initial visible value
		// to the value from the existing item, or true for new items.
		if prevItem.ID != "" {
			item.Visible = prevItem.Attributes.Visible
		} else {
			item.Visible = true
		}

		for _, f := range p.Transforms {
			f(&item)
		}

		// After running configured transformations, set automation package
		// annotations.
		item.Annotations.Set(AnnotationPublisherName, opts.Publisher)
		item.Annotations.Set(AnnotationPublisherTransformVersion, p.TransformVersion)
		item.Annotations.Set(AnnotationPublisherSignalAttributes, signal.Meta.AttributesHash.String())
		item.Annotations.Set(AnnotationPublisherSignalID, signal.ID)

		dest[signal.ID] = item
	}

	var more bool
	if results.Meta.Total >= 0 {
		// More can be calculated exactly when the total count was requested (or
		// calculated for free by the backend).
		more = q.Skip+len(results.Data) < results.Meta.Total
	} else {
		// Fallback to approximate the more value when total was not explicitly
		// requested. For large values of q.Limit, this approximation is likely
		// faster -- on average -- then to request a total count.
		more = (len(results.Data) == q.Limit)
	}
	return more, nil

}
