package contracttest

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func appendNullPreserveIssues(ctx context.Context, handler iface.Handler, fixture string, skipFields []string, issues *[]string) {
	block := handler.PanelType() + "_config"
	if !panelkit.HasPanelConfigBlock(block) {
		return
	}
	sna, ok := handler.SchemaAttribute().(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	lp := collectLeafPaths(sna)

	item0, err := ParseDashboardPanel(fixture)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[NullPreserve] parse: %v", err))
		return
	}

	var baseline models.PanelModel
	if diags := handler.FromAPI(ctx, &baseline, nil, item0); diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("[NullPreserve] baseline FromAPI: %s", summarizeDiags(diags)))
		return
	}

	for _, leaf := range lp.optional {
		if len(leaf) != 1 {
			continue
		}
		sk := leaf[0]
		if slices.Contains(skipFields, sk) || skipHasPrefix(skipFields, sk) {
			continue
		}
		want, ok := readAttrLeaf(&baseline, block, leaf)
		if !ok {
			continue
		}
		nullPrior := synthesizePriorWithNull(handler, fixture, block, baseline, leaf, sk)
		if nullPrior != nil {
			assertFromAPI(ctx, handler, fixture, issues, "[NullPreserve] prior_null/"+sk, nullPrior,
				func(out *models.PanelModel) bool {
					got, ok := readAttrLeaf(out, block, leaf)
					if !ok {
						return false
					}
					if s, ok := got.(types.String); ok {
						return s.IsNull()
					}
					if b, ok := got.(types.Bool); ok {
						return b.IsNull()
					}
					if f, ok := got.(types.Float64); ok {
						return f.IsNull()
					}
					if i, ok := got.(types.Int64); ok {
						return i.IsNull()
					}
					return false
				})
		}
		knownPrior := synthesizePriorWithStaleString(handler, block, baseline, leaf, sk, want)
		if knownPrior != nil {
			assertFromAPI(ctx, handler, fixture, issues, "[NullPreserve] prior_known/"+sk, knownPrior,
				func(out *models.PanelModel) bool {
					got, ok := readAttrLeaf(out, block, leaf)
					if !ok {
						return false
					}
					return attrsComparableEqual(want, got)
				})
		}

		assertFromAPI(ctx, handler, fixture, issues, "[NullPreserve] fresh_import/"+sk, nil,
			func(out *models.PanelModel) bool {
				got, ok := readAttrLeaf(out, block, leaf)
				if !ok {
					return false
				}
				return attrsComparableEqual(want, got)
			})
	}
}

func skipHasPrefix(skip []string, sk string) bool {
	for _, s := range skip {
		if strings.HasPrefix(sk, s+".") {
			return true
		}
	}
	return false
}

func synthesizePriorWithNull(
	handler iface.Handler,
	fixture string,
	block string,
	baseline models.PanelModel,
	path []string,
	sk string,
) *models.PanelModel {
	_ = handler
	_ = fixture
	p, err := clonePanel(&baseline)
	if err != nil {
		return nil
	}
	panelkit.EnsureMutableTypedConfig(p, block)

	switch sniffLeafKind(sk) {
	case leafKindGuessString:
		_ = setStructLeaf(p, block, path, types.StringNull())
	case leafKindGuessBool:
		_ = setStructLeaf(p, block, path, types.BoolNull())
	default:
		return nil
	}
	p.Type = baseline.Type
	return p
}

func synthesizePriorWithStaleString(handler iface.Handler, block string, baseline models.PanelModel, path []string, sk string, want attr.Value) *models.PanelModel {
	_, ok := want.(types.String)
	if !ok || sniffLeafKind(sk) != leafKindGuessString {
		_ = handler
		return nil
	}
	p, err := clonePanel(&baseline)
	if err != nil {
		return nil
	}
	panelkit.EnsureMutableTypedConfig(p, block)
	p.Type = baseline.Type
	_ = setStructLeaf(p, block, path, types.StringValue("stale-prior-contracttest"))
	return p
}

type leafGuess int

const (
	leafGuessNone leafGuess = iota
	leafKindGuessString
	leafKindGuessBool
)

func sniffLeafKind(sk string) leafGuess {
	if strings.Contains(sk, "hide_") {
		return leafKindGuessBool
	}
	if strings.Contains(sk, "encode") {
		return leafKindGuessBool
	}
	if strings.Contains(sk, "open_in") {
		return leafKindGuessBool
	}
	switch {
	case strings.HasSuffix(sk, "_id"), strings.Contains(sk, "duration"), strings.HasSuffix(sk, "title"), strings.Contains(sk, "description"):
		return leafKindGuessString
	default:
		return leafGuessNone
	}
}

func assertFromAPI(
	ctx context.Context,
	handler iface.Handler,
	fixture string,
	issues *[]string,
	label string,
	prior *models.PanelModel,
	ok func(*models.PanelModel) bool,
) {
	item0, err := ParseDashboardPanel(fixture)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("%s: parse: %v", label, err))
		return
	}
	var pm models.PanelModel
	if prior != nil {
		pm = *prior
	}
	diags := handler.FromAPI(ctx, &pm, prior, item0)
	if diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("%s: FromAPI: %s", label, summarizeDiags(diags)))
		return
	}
	if !ok(&pm) {
		*issues = append(*issues, fmt.Sprintf("%s: post-condition failed", label))
	}
}
