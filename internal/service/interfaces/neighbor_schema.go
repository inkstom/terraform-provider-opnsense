package interfaces

import (
	"github.com/browningluke/opnsense-go/pkg/interfaces"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type neighborResourceModel struct {
	MacAddress  types.String `tfsdk:"mac_address"`
	IpAddress   types.String `tfsdk:"ip_address"`
	Description types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

const neighborIPHelp = "IP address to assign to the provided MAC address, which will either end up in the ARP (IPv4) or NDP (IPv6) table."

func neighborResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Static ARP/NDP entries (Interfaces > Neighbors).",
		Attributes: map[string]schema.Attribute{
			"mac_address": reqStr("Hardware MAC address of the client (format `xx:xx:xx:xx:xx:xx`)."),
			"ip_address":  reqStr(neighborIPHelp),
			"description": optStr(descrHelp, ""),
			"id":          computedStr("UUID of the entry."),
		},
	}
}

func neighborDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Static ARP/NDP entries (Interfaces > Neighbors).",
		Attributes: map[string]dschema.Attribute{
			"id":          dsID("UUID of the entry."),
			"mac_address": dsStr("Hardware MAC address of the client."),
			"ip_address":  dsStr(neighborIPHelp),
			"description": dsStr(descrHelp),
		},
	}
}

func convertNeighborSchemaToStruct(d *neighborResourceModel) (*interfaces.Neighbor, error) {
	return &interfaces.Neighbor{
		MacAddress:  d.MacAddress.ValueString(),
		IpAddress:   d.IpAddress.ValueString(),
		Description: d.Description.ValueString(),
	}, nil
}

func convertNeighborStructToSchema(d *interfaces.Neighbor) (*neighborResourceModel, error) {
	return &neighborResourceModel{
		MacAddress:  types.StringValue(d.MacAddress),
		IpAddress:   types.StringValue(d.IpAddress),
		Description: types.StringValue(d.Description),
	}, nil
}

func mergeNeighborWriteOnly(prior, read *neighborResourceModel) {}
