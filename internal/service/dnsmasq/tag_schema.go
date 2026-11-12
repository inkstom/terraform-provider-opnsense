package dnsmasq

import (
	"github.com/browningluke/opnsense-go/pkg/dnsmasq"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagResourceModel struct {
	Tag types.String `tfsdk:"tag"`

	Id types.String `tfsdk:"id"`
}

func tagResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Dnsmasq DHCP tags (Services > Dnsmasq DNS & DHCP > DHCP tags). Hosts and ranges set tags on clients; options and boot entries apply only to clients with a matching tag. Reference a tag by `id`.",
		Attributes: map[string]schema.Attribute{
			"tag": schema.StringAttribute{MarkdownDescription: "Tag name.", Required: true},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the tag.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func tagDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Dnsmasq DHCP tags (Services > Dnsmasq DNS & DHCP > DHCP tags).",
		Attributes: map[string]dschema.Attribute{
			"id":  dschema.StringAttribute{MarkdownDescription: "UUID of the tag.", Required: true},
			"tag": dschema.StringAttribute{MarkdownDescription: "Tag name.", Computed: true},
		},
	}
}

func convertTagSchemaToStruct(d *tagResourceModel) (*dnsmasq.Tag, error) {
	return &dnsmasq.Tag{Tag: d.Tag.ValueString()}, nil
}

func convertTagStructToSchema(d *dnsmasq.Tag) (*tagResourceModel, error) {
	return &tagResourceModel{Tag: types.StringValue(d.Tag)}, nil
}

func mergeTagWriteOnly(prior, read *tagResourceModel) {}
