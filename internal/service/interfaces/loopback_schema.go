package interfaces

import (
	"github.com/browningluke/opnsense-go/pkg/interfaces"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type loopbackResourceModel struct {
	Description types.String `tfsdk:"description"`
	Device      types.String `tfsdk:"device"`

	Id types.String `tfsdk:"id"`
}

func loopbackResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Loopback interfaces (Interfaces > Devices > Loopback), e.g. for a router ID or a service address. " +
			"Assign the device with `opnsense_interfaces_assignment` to give it an address.",
		Attributes: map[string]schema.Attribute{
			"description": optStr("Description.", ""),
			"device":      computedStr("Device name assigned by OPNsense, e.g. `lo1`."),
			"id":          computedStr("UUID of the loopback."),
		},
	}
}

func loopbackDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Loopback interfaces (Interfaces > Devices > Loopback).",
		Attributes: map[string]dschema.Attribute{
			"id":          dsID("UUID of the loopback."),
			"description": dsStr("Description."),
			"device":      dsStr("Device name, e.g. `lo1`."),
		},
	}
}

func convertLoopbackSchemaToStruct(d *loopbackResourceModel) (*interfaces.Loopback, error) {
	return &interfaces.Loopback{Description: d.Description.ValueString()}, nil
}

func convertLoopbackStructToSchema(d *interfaces.Loopback) (*loopbackResourceModel, error) {
	return &loopbackResourceModel{
		Description: types.StringValue(d.Description),
		Device:      types.StringValue("lo" + d.DeviceId),
	}, nil
}

func mergeLoopbackWriteOnly(prior, read *loopbackResourceModel) {}
