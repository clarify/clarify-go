// Copyright 2022-2023 Searis AS
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
	"fmt"
	"log/slog"

	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
)

const (
	selectSignalsPageSize  = 1000
	publishSignalsPageSize = 500
)

// PublishSignals allows you to automate signal publishing from one or more
// source integrations. The routine respects the DryRun and EarlyOut
// configurations.
type PublishSignals struct {
	// Integrations must list the IDs of the integrations to publish signals
	// from. If this list is empty, the rule set is a no-op.
	Integrations []string

	// SignalsFilter can optionally be specified to limit which signals to
	// publish.
	SignalsFilter fields.ResourceFilterType

	// TransformVersion should be changed if you want existing items to be
	// republished despite the source signal being unchanged.
	TransformVersion string

	// Transforms is a list of transforms to apply when publishing the signals.
	// The transforms are applied in order.
	Transforms []func(item *views.ItemSave)
}

var _ Routine = PublishSignals{}

func (p PublishSignals) Do(ctx context.Context, cfg *Config) error {
	logger := cfg.Logger()
	client := cfg.Client()
	earlyOut := cfg.EarlyOut()

	var publishCount, errorCount int

	defer func() {
		logger.LogAttrs(ctx, slog.LevelInfo, "Publish signals completed",
			slog.Int("integration_count", len(p.Integrations)),
			slog.Int("publish_count", publishCount),
			slog.Int("error_count", errorCount),
		)
	}()

	if err := ctx.Err(); err != nil {
		return err
	}

	// We iterate signals without requesting the total count. This is an
	// optimization bet that total % limit == 0 is uncommon.
	query := fields.Query().Sort("id").Limit(selectSignalsPageSize)
	if p.SignalsFilter != nil {
		query = query.Where(p.SignalsFilter)
	}

	items := make(map[string]views.ItemSave)
	flush := func(integrationID string) error {
		logger.LogAttrs(ctx, slog.LevelInfo, "Publish signals", slog.Int("publish_count", publishCount))
		logger.LogAttrs(ctx, slog.LevelDebug, "Publish parameters", slog.Group("params", slog.Any("itemBySignal", items)))

		if !cfg.DryRun() {
			result, err := client.Admin().PublishSignals(integrationID, items).Do(ctx)
			if err != nil {
				if earlyOut {
					return fmt.Errorf("publish signals: %w", err)
				} else {
					logger.LogAttrs(ctx, slog.LevelError, "Published items failed (flush)", AttrError(err), slog.Int("publish_count", len(items)))
				}
				errorCount += len(items)
			} else {
				logger.LogAttrs(ctx, slog.LevelInfo, "Published items (flush)", slog.Int("publish_count", len(items)))
				publishCount += len(items)
				logger.LogAttrs(ctx, slog.LevelDebug, "Publish results", slog.Any("result", result))
			}
		} else {
			publishCount += len(items)
		}

		items = make(map[string]views.ItemSave)
		return nil
	}

	for _, id := range p.Integrations {
		more := true
		for more {
			if err := ctx.Err(); err != nil {
				return err
			}

			var err error
			more, err = p.addItems(ctx, cfg, items, id, query)
			if err != nil {
				return err
			}
			if len(items) >= publishSignalsPageSize {
				if err := flush(id); err != nil {
					return err
				}
			}
			query = query.NextPage()
		}

		if err := flush(id); err != nil {
			return err
		}
	}

	return nil
}

// addItems adds items that require update to dest from all signals matching
// the integration ID and query.
func (p PublishSignals) addItems(ctx context.Context, cfg *Config, dest map[string]views.ItemSave, integrationID string, query fields.ResourceQuery) (bool, error) {
	logger := cfg.Logger()
	client := cfg.Client()

	results, err := client.Admin().SelectSignals(integrationID, query).Include("item").Do(ctx)
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
			logger.LogAttrs(
				ctx, slog.LevelDebug, "Item is up-to-date",
				slog.String("signal_id", signal.ID),
				slog.String("item_id", signal.Relationships.Item.Data.ID),
			)
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
		item.Annotations.Set(AnnotationPublisherName, cfg.AppName())
		item.Annotations.Set(AnnotationPublisherTransformVersion, p.TransformVersion)
		item.Annotations.Set(AnnotationPublisherSignalAttributes, signal.Meta.AttributesHash.String())
		item.Annotations.Set(AnnotationPublisherSignalID, signal.ID)

		dest[signal.ID] = item
	}

	var more bool
	if results.Meta.Total >= 0 {
		// More can be calculated exactly when the total count was requested (or
		// calculated for free by the backend).
		more = query.GetSkip()+len(results.Data) < results.Meta.Total
	} else {
		// Fallback to approximate the more value when total was not explicitly
		// requested. For large values of q.Limit, this approximation is likely
		// faster -- on average -- then to request a total count.
		more = (len(results.Data) == query.GetLimit())
	}
	return more, nil
}
