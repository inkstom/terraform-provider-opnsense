package routing

import (
	"context"
	"fmt"
	"regexp"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/routing"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// maxGatewayGroupTiers is the number of tiers (item, item2..item5) in the OPNsense model.
const maxGatewayGroupTiers = 5

var gatewayNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,32}$`)

var gatewayTierType = types.SetType{ElemType: types.StringType}

// gatewayGroupResourceModel describes the resource data model.
type gatewayGroupResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Tiers       types.List   `tfsdk:"tiers"`
	Trigger     types.String `tfsdk:"trigger"`
	PoolOptions types.String `tfsdk:"pool_options"`
	Description types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func gatewayGroupResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Gateway groups combine gateways into tiers for failover and load balancing (System > Gateways > Group). " +
			"Requires OPNsense 26.7 or later. A group cannot be deleted while it is referenced, e.g. by a firewall rule.",
		Version: 1,

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Group name. Letters, digits, `_` and `-` only, max 32 characters. Firewall rules refer to the group by this name.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(gatewayNameRegex, "must match ^[a-zA-Z0-9_-]{1,32}$"),
				},
			},
			"tiers": schema.ListAttribute{
				MarkdownDescription: "Ordered list of tiers, each a set of gateway **names**. The first element is tier 1 (highest priority). " +
					"Gateways in the same tier load-balance; lower tiers are used when all gateways in the higher tiers are down. Between 1 and 5 tiers. " +
					"Example: `[[\"WAN1_DHCP\"], [\"WAN2_DHCP\"]]` for failover.",
				Required:    true,
				ElementType: gatewayTierType,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, maxGatewayGroupTiers),
					listvalidator.ValueSetsAre(setvalidator.SizeAtLeast(1)),
				},
			},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "When to consider a member down: `down` (member down), `downloss` (packet loss), `downlatency` (high latency) or `downlosslatency` (packet loss or high latency). Defaults to `down`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("down"),
				Validators: []validator.String{
					stringvalidator.OneOf("down", "downloss", "downlatency", "downlosslatency"),
				},
			},
			"pool_options": schema.StringAttribute{
				MarkdownDescription: "Load-balancing pool option: `\"\"` (default), `round-robin` or `round-robin sticky-address`. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.OneOf("", "round-robin", "round-robin sticky-address"),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description here for your reference (not parsed).",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the gateway group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func gatewayGroupDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Gateway groups combine gateways into tiers for failover and load balancing.",

		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				MarkdownDescription: "UUID of the gateway group.",
				Required:            true,
			},
			"name": dschema.StringAttribute{
				MarkdownDescription: "Group name.",
				Computed:            true,
			},
			"tiers": dschema.ListAttribute{
				MarkdownDescription: "Ordered list of tiers, each a set of gateway names.",
				Computed:            true,
				ElementType:         gatewayTierType,
			},
			"trigger": dschema.StringAttribute{
				MarkdownDescription: "Trigger level.",
				Computed:            true,
			},
			"pool_options": dschema.StringAttribute{
				MarkdownDescription: "Load-balancing pool option.",
				Computed:            true,
			},
			"description": dschema.StringAttribute{
				MarkdownDescription: "Description.",
				Computed:            true,
			},
		},
	}
}

func convertGatewayGroupSchemaToStruct(d *gatewayGroupResourceModel) (*routing.GatewayGroup, error) {
	var tiers [][]string
	if diags := d.Tiers.ElementsAs(context.Background(), &tiers, false); diags.HasError() {
		return nil, fmt.Errorf("unable to parse tiers: %v", diags)
	}
	if len(tiers) > maxGatewayGroupTiers {
		return nil, fmt.Errorf("at most %d tiers are supported, got %d", maxGatewayGroupTiers, len(tiers))
	}

	// Always send all five tiers so tiers removed from config are cleared upstream.
	padded := make([]api.SelectedMapList, maxGatewayGroupTiers)
	for i := range padded {
		padded[i] = api.SelectedMapList{}
		if i < len(tiers) {
			padded[i] = tiers[i]
		}
	}

	return &routing.GatewayGroup{
		Name:        d.Name.ValueString(),
		Tier1:       padded[0],
		Tier2:       padded[1],
		Tier3:       padded[2],
		Tier4:       padded[3],
		Tier5:       padded[4],
		Trigger:     api.SelectedMap(d.Trigger.ValueString()),
		PoolOptions: api.SelectedMap(d.PoolOptions.ValueString()),
		Description: d.Description.ValueString(),
	}, nil
}

func convertGatewayGroupStructToSchema(d *routing.GatewayGroup) (*gatewayGroupResourceModel, error) {
	all := [][]string{d.Tier1, d.Tier2, d.Tier3, d.Tier4, d.Tier5}

	// Trim trailing empty tiers; interior empty tiers are kept so positions stay stable.
	last := -1
	for i, t := range all {
		if len(nonEmpty(t)) > 0 {
			last = i
		}
	}
	tiers := make([][]string, 0, last+1)
	for i := 0; i <= last; i++ {
		tiers = append(tiers, nonEmpty(all[i]))
	}

	tierList, diags := types.ListValueFrom(context.Background(), gatewayTierType, tiers)
	if diags.HasError() {
		return nil, fmt.Errorf("unable to build tiers: %v", diags)
	}

	return &gatewayGroupResourceModel{
		Name:        types.StringValue(d.Name),
		Tiers:       tierList,
		Trigger:     types.StringValue(d.Trigger.String()),
		PoolOptions: types.StringValue(d.PoolOptions.String()),
		Description: tools.StringOrNull(d.Description),
	}, nil
}

func nonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
