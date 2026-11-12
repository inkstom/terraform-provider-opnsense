package dhcrelay

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/dhcrelay"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type destinationResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Servers types.Set    `tfsdk:"servers"`

	Id types.String `tfsdk:"id"`
}

func destinationResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "DHCP relay destinations: the DHCP servers a relay forwards requests to (Services > DHCP Relay > Destinations).",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Destination name.",
				Required:            true,
			},
			"servers": schema.SetAttribute{
				MarkdownDescription: "DHCP server addresses (IPv4 or IPv6, not mixed).",
				ElementType:         types.StringType,
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the destination.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func destinationDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "DHCP relay destinations (Services > DHCP Relay > Destinations).",

		Attributes: map[string]dschema.Attribute{
			"id":      dschema.StringAttribute{MarkdownDescription: "UUID of the destination.", Required: true},
			"name":    dschema.StringAttribute{MarkdownDescription: "Destination name.", Computed: true},
			"servers": dschema.SetAttribute{MarkdownDescription: "DHCP server addresses.", ElementType: types.StringType, Computed: true},
		},
	}
}

func convertDestinationSchemaToStruct(d *destinationResourceModel) (*dhcrelay.Destination, error) {
	return &dhcrelay.Destination{
		Name:   d.Name.ValueString(),
		Server: api.SelectedMapList(tools.SetToStringSlice(d.Servers)),
	}, nil
}

func convertDestinationStructToSchema(d *dhcrelay.Destination) (*destinationResourceModel, error) {
	return &destinationResourceModel{
		Name:    types.StringValue(d.Name),
		Servers: tools.StringSliceToSet(d.Server),
	}, nil
}

func mergeDestinationWriteOnly(prior, read *destinationResourceModel) {}
