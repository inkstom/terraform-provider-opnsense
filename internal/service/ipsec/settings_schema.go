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
			"prefer_old_sa":        keepBool("By default, if several SAs match, the newest one is preferred if it's at least 30 seconds old. Select this option to always prefer old SAs over new ones."),
			"disable_vpn_rules":    keepBool("This option only applies to legacy tunnel configurations, connections do require manual firewall rules being setup."),
			"passthrough_networks": keepSet("This exempts traffic for one or more subnets from getting processed by the IPsec stack in the kernel. When sending all traffic to the remote location, you probably want to add your LAN network(s) here."),
			"user_sources":         keepSet("Select authentication methods to use, leave empty if no challenge response authentication is needed. Values are authentication server names, e.g. `Local Database`. Deprecated in OPNsense (Xauth)."),
			"local_group": schema.StringAttribute{
				MarkdownDescription: "Restrict access to users in the selected local group. Please be aware that other authentication backends will refuse to authenticate when using this option. Value is the group GID; empty allows all. Deprecated in OPNsense (Xauth). Unset keeps the current value.",
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
			"prefer_old_sa":        dschema.BoolAttribute{MarkdownDescription: "By default, if several SAs match, the newest one is preferred if it's at least 30 seconds old. Select this option to always prefer old SAs over new ones.", Computed: true},
			"disable_vpn_rules":    dschema.BoolAttribute{MarkdownDescription: "This option only applies to legacy tunnel configurations, connections do require manual firewall rules being setup.", Computed: true},
			"passthrough_networks": dschema.SetAttribute{MarkdownDescription: "This exempts traffic for one or more subnets from getting processed by the IPsec stack in the kernel. When sending all traffic to the remote location, you probably want to add your LAN network(s) here.", ElementType: types.StringType, Computed: true},
			"user_sources":         dschema.SetAttribute{MarkdownDescription: "Select authentication methods to use, leave empty if no challenge response authentication is needed. Values are authentication server names, e.g. `Local Database`. Deprecated in OPNsense (Xauth).", ElementType: types.StringType, Computed: true},
			"local_group":          dschema.StringAttribute{MarkdownDescription: "Restrict access to users in the selected local group. Please be aware that other authentication backends will refuse to authenticate when using this option. Value is the group GID; empty allows all. Deprecated in OPNsense (Xauth).", Computed: true},
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
