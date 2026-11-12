package ipsec

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/ipsec"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const settingsID = "ipsec_settings"

type settingsResourceModel struct {
	Enabled             types.Bool   `tfsdk:"enabled"`
	PreferOldSa         types.Bool   `tfsdk:"prefer_old_sa"`
	DisableVpnRules     types.Bool   `tfsdk:"disable_vpn_rules"`
	PassthroughNetworks types.Set    `tfsdk:"passthrough_networks"`
	UserSources         types.Set    `tfsdk:"user_sources"`
	LocalGroup          types.String `tfsdk:"local_group"`

	Id types.String `tfsdk:"id"`
}

func settingsResourceSchema() schema.Schema {
	keepBool := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{MarkdownDescription: desc + " Unset keeps the current value.", Optional: true, Computed: true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}}
	}
	keepSet := func(desc string) schema.SetAttribute {
		return schema.SetAttribute{MarkdownDescription: desc + " Unset keeps the current value.", ElementType: types.StringType, Optional: true, Computed: true,
			PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()}}
	}
	return schema.Schema{
		MarkdownDescription: "IPsec general settings (VPN > IPsec > Advanced Settings, general section). Charon tuning and logging levels are not covered. " +
			"This is a singleton that always exists, so it must be imported before Terraform manages it (`id = \"" + settingsID + "\"`); " +
			"destroying it only removes it from state.",

		Attributes: map[string]schema.Attribute{
			"enabled":              keepBool("Enable IPsec (start strongSwan)."),
			"prefer_old_sa":        keepBool("Prefer older security associations over newer ones during rekeying."),
			"disable_vpn_rules":    keepBool("Don't add the automatic firewall rules for IPsec; you write them yourself."),
			"passthrough_networks": keepSet("Networks that bypass IPsec, e.g. local subnets also covered by a tunnel's policy."),
			"user_sources":         keepSet("Authentication servers for mobile users, by name (e.g. `Local Database`)."),
			"local_group": schema.StringAttribute{
				MarkdownDescription: "Only members of this group (by GID) may connect as mobile users; empty allows all. Unset keeps the current value.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `" + settingsID + "`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func settingsDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "IPsec general settings (VPN > IPsec > Advanced Settings, general section).",

		Attributes: map[string]dschema.Attribute{
			"enabled":              dschema.BoolAttribute{MarkdownDescription: "Whether IPsec is enabled.", Computed: true},
			"prefer_old_sa":        dschema.BoolAttribute{MarkdownDescription: "Whether older SAs are preferred.", Computed: true},
			"disable_vpn_rules":    dschema.BoolAttribute{MarkdownDescription: "Whether automatic rules are disabled.", Computed: true},
			"passthrough_networks": dschema.SetAttribute{MarkdownDescription: "Passthrough networks.", ElementType: types.StringType, Computed: true},
			"user_sources":         dschema.SetAttribute{MarkdownDescription: "Mobile user authentication servers.", ElementType: types.StringType, Computed: true},
			"local_group":          dschema.StringAttribute{MarkdownDescription: "Mobile user group.", Computed: true},
			"id":                   dschema.StringAttribute{MarkdownDescription: "Always `" + settingsID + "`.", Computed: true},
		},
	}
}

func convertSettingsSchemaToStruct(d *settingsResourceModel) *ipsec.SettingsBody {
	return &ipsec.SettingsBody{General: ipsec.SettingsGeneral{
		Enabled:             tools.BoolToString(d.Enabled.ValueBool()),
		PreferredOldSa:      tools.BoolToString(d.PreferOldSa.ValueBool()),
		DisableVpnRules:     tools.BoolToString(d.DisableVpnRules.ValueBool()),
		PassthroughNetworks: api.SelectedMapList(tools.SetToStringSlice(d.PassthroughNetworks)),
		UserSource:          api.SelectedMapList(tools.SetToStringSlice(d.UserSources)),
		LocalGroup:          api.SelectedMap(d.LocalGroup.ValueString()),
	}}
}

func convertSettingsStructToSchema(d *ipsec.SettingsGeneral) *settingsResourceModel {
	return &settingsResourceModel{
		Enabled:             types.BoolValue(tools.StringToBool(d.Enabled)),
		PreferOldSa:         types.BoolValue(tools.StringToBool(d.PreferredOldSa)),
		DisableVpnRules:     types.BoolValue(tools.StringToBool(d.DisableVpnRules)),
		PassthroughNetworks: tools.StringSliceToSet(d.PassthroughNetworks),
		UserSources:         tools.StringSliceToSet(d.UserSource),
		LocalGroup:          types.StringValue(d.LocalGroup.String()),
		Id:                  types.StringValue(settingsID),
	}
}
