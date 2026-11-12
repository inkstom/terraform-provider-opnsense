package kea

import (
	"github.com/browningluke/opnsense-go/pkg/kea"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const ctrlAgentID = "kea_ctrl_agent"

type ctrlAgentResourceModel struct {
	Enabled  types.Bool   `tfsdk:"enabled"`
	HttpHost types.String `tfsdk:"http_host"`
	HttpPort types.Int64  `tfsdk:"http_port"`

	Id types.String `tfsdk:"id"`
}

func ctrlAgentResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Kea control agent (Services > Kea DHCP > Control Agent), the REST endpoint Kea high availability peers use to talk to each other. " +
			"This is a singleton that always exists, so it must be imported before Terraform manages it (`id = \"" + ctrlAgentID + "\"`); " +
			"destroying it only removes it from state. Attributes left unset keep the current value.",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Run the control agent. Required for Kea HA (`opnsense_kea_dhcpv4_peer`). OPNsense only starts it while Kea DHCPv4 or DHCPv6 is enabled.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"http_host": schema.StringAttribute{
				MarkdownDescription: "Address the agent listens on; HA peers must be able to reach it.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"http_port": schema.Int64Attribute{
				MarkdownDescription: "Port the agent listens on.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				Validators:          []validator.Int64{int64validator.Between(1, 65535)},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `" + ctrlAgentID + "`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func ctrlAgentDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Kea control agent settings (Services > Kea DHCP > Control Agent).",

		Attributes: map[string]dschema.Attribute{
			"enabled":   dschema.BoolAttribute{MarkdownDescription: "Whether the agent runs.", Computed: true},
			"http_host": dschema.StringAttribute{MarkdownDescription: "Listen address.", Computed: true},
			"http_port": dschema.Int64Attribute{MarkdownDescription: "Listen port.", Computed: true},
			"id":        dschema.StringAttribute{MarkdownDescription: "Always `" + ctrlAgentID + "`.", Computed: true},
		},
	}
}

func convertCtrlAgentSchemaToStruct(d *ctrlAgentResourceModel) *kea.CtrlAgentBody {
	return &kea.CtrlAgentBody{General: kea.CtrlAgentGeneral{
		Enabled:  tools.BoolToString(d.Enabled.ValueBool()),
		HttpHost: d.HttpHost.ValueString(),
		HttpPort: tools.Int64ToString(d.HttpPort.ValueInt64()),
	}}
}

func convertCtrlAgentStructToSchema(d *kea.CtrlAgentGeneral) *ctrlAgentResourceModel {
	return &ctrlAgentResourceModel{
		Enabled:  types.BoolValue(tools.StringToBool(d.Enabled)),
		HttpHost: types.StringValue(d.HttpHost),
		HttpPort: types.Int64Value(tools.StringToInt64(d.HttpPort)),
		Id:       types.StringValue(ctrlAgentID),
	}
}
