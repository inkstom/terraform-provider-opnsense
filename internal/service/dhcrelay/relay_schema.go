package dhcrelay

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/dhcrelay"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type relayResourceModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	Interface    types.String `tfsdk:"interface"`
	Destination  types.String `tfsdk:"destination"`
	AgentInfo    types.Bool   `tfsdk:"agent_info"`
	CircuitId    types.String `tfsdk:"circuit_id"`
	RemoteId     types.String `tfsdk:"remote_id"`
	CarpDependOn types.String `tfsdk:"carp_depend_on"`

	Id types.String `tfsdk:"id"`
}

func relayResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "DHCP relays: forward DHCP requests arriving on an interface to a destination (Services > DHCP Relay > Relays). " +
			"An interface can have one relay per address family, and must not also run a DHCP server.",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable the relay. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Interface to listen on, e.g. `lan` or `opt1`.",
				Required:            true,
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "UUID of the `opnsense_dhcrelay_destination` to forward to.",
				Required:            true,
			},
			"agent_info": schema.BoolAttribute{
				MarkdownDescription: "Add relay agent information (option 82) to forwarded requests. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"circuit_id": schema.StringAttribute{
				MarkdownDescription: "Option 82 circuit ID; empty uses the interface name. Needs `agent_info` (OPNsense 26.7+). Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"remote_id": schema.StringAttribute{
				MarkdownDescription: "Option 82 remote ID. Needs `agent_info` (OPNsense 26.7+). Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"carp_depend_on": schema.StringAttribute{
				MarkdownDescription: "Only run while this CARP VIP is master (VIP UUID or address as listed by OPNsense); empty always runs. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the relay.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func relayDataSourceSchema() dschema.Schema {
	str := func(desc string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "DHCP relays (Services > DHCP Relay > Relays).",

		Attributes: map[string]dschema.Attribute{
			"id":             dschema.StringAttribute{MarkdownDescription: "UUID of the relay.", Required: true},
			"enabled":        dschema.BoolAttribute{MarkdownDescription: "Whether the relay is enabled.", Computed: true},
			"interface":      str("Interface listened on."),
			"destination":    str("Destination UUID."),
			"agent_info":     dschema.BoolAttribute{MarkdownDescription: "Whether option 82 is added.", Computed: true},
			"circuit_id":     str("Option 82 circuit ID."),
			"remote_id":      str("Option 82 remote ID."),
			"carp_depend_on": str("CARP VIP the relay depends on."),
		},
	}
}

func convertRelaySchemaToStruct(d *relayResourceModel) (*dhcrelay.Relay, error) {
	return &dhcrelay.Relay{
		Enabled:      tools.BoolToString(d.Enabled.ValueBool()),
		Interface:    api.SelectedMap(d.Interface.ValueString()),
		Destination:  api.SelectedMap(d.Destination.ValueString()),
		AgentInfo:    tools.BoolToString(d.AgentInfo.ValueBool()),
		CircuitId:    d.CircuitId.ValueString(),
		RemoteId:     d.RemoteId.ValueString(),
		CarpDependOn: api.SelectedMap(d.CarpDependOn.ValueString()),
	}, nil
}

func convertRelayStructToSchema(d *dhcrelay.Relay) (*relayResourceModel, error) {
	return &relayResourceModel{
		Enabled:      types.BoolValue(tools.StringToBool(d.Enabled)),
		Interface:    types.StringValue(d.Interface.String()),
		Destination:  types.StringValue(d.Destination.String()),
		AgentInfo:    types.BoolValue(tools.StringToBool(d.AgentInfo)),
		CircuitId:    types.StringValue(d.CircuitId),
		RemoteId:     types.StringValue(d.RemoteId),
		CarpDependOn: types.StringValue(d.CarpDependOn.String()),
	}, nil
}

func mergeRelayWriteOnly(prior, read *relayResourceModel) {}
