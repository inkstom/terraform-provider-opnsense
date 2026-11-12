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

func neighborResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Static ARP/NDP entries (Interfaces > Neighbors).",
		Attributes: map[string]schema.Attribute{
			"mac_address": reqStr("Hardware address, e.g. `00:11:22:33:44:55`."),
			"ip_address":  reqStr("IPv4 or IPv6 address."),
			"description": optStr("Description.", ""),
			"id":          computedStr("UUID of the entry."),
		},
	}
}

func neighborDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Static ARP/NDP entries (Interfaces > Neighbors).",
		Attributes: map[string]dschema.Attribute{
			"id":          dsID("UUID of the entry."),
			"mac_address": dsStr("Hardware address."),
			"ip_address":  dsStr("IP address."),
			"description": dsStr("Description."),
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
