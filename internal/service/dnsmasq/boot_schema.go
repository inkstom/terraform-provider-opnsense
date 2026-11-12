package dnsmasq

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/dnsmasq"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type bootResourceModel struct {
	Interface     types.String `tfsdk:"interface"`
	Tags          types.Set    `tfsdk:"tags"`
	Filename      types.String `tfsdk:"filename"`
	ServerName    types.String `tfsdk:"server_name"`
	ServerAddress types.String `tfsdk:"server_address"`
	Description   types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func bootResourceSchema() schema.Schema {
	str := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc + " Defaults to `\"\"`.", Optional: true, Computed: true, Default: stringdefault.StaticString("")}
	}
	return schema.Schema{
		MarkdownDescription: "Dnsmasq network boot entries (Services > Dnsmasq DNS & DHCP > DHCP boot): the PXE/TFTP boot file and server handed to DHCP clients.",
		Attributes: map[string]schema.Attribute{
			"interface": str("Only answer on this interface; empty answers on all."),
			"tags": schema.SetAttribute{
				MarkdownDescription: "UUIDs of `opnsense_dnsmasq_tag`s a client must have for this entry to apply; empty applies to all. Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"filename":       schema.StringAttribute{MarkdownDescription: "Boot file, e.g. `pxelinux.0` or `ipxe.efi`.", Required: true},
			"server_name":    str("Boot server host name."),
			"server_address": str("Boot (TFTP) server address; empty uses this firewall."),
			"description":    str("Description."),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the entry.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func bootDataSourceSchema() dschema.Schema {
	s := func(d string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: d, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "Dnsmasq network boot entries (Services > Dnsmasq DNS & DHCP > DHCP boot).",
		Attributes: map[string]dschema.Attribute{
			"id":             dschema.StringAttribute{MarkdownDescription: "UUID of the entry.", Required: true},
			"interface":      s("Interface."),
			"tags":           dschema.SetAttribute{MarkdownDescription: "Tag UUIDs.", ElementType: types.StringType, Computed: true},
			"filename":       s("Boot file."),
			"server_name":    s("Boot server host name."),
			"server_address": s("Boot server address."),
			"description":    s("Description."),
		},
	}
}

func convertBootSchemaToStruct(d *bootResourceModel) (*dnsmasq.Boot, error) {
	return &dnsmasq.Boot{
		Interface:     api.SelectedMap(d.Interface.ValueString()),
		Tag:           api.SelectedMapList(tools.SetToStringSlice(d.Tags)),
		Filename:      d.Filename.ValueString(),
		Servername:    d.ServerName.ValueString(),
		ServerAddress: d.ServerAddress.ValueString(),
		Description:   d.Description.ValueString(),
	}, nil
}

func convertBootStructToSchema(d *dnsmasq.Boot) (*bootResourceModel, error) {
	return &bootResourceModel{
		Interface:     types.StringValue(d.Interface.String()),
		Tags:          tools.StringSliceToSet(d.Tag),
		Filename:      types.StringValue(d.Filename),
		ServerName:    types.StringValue(d.Servername),
		ServerAddress: types.StringValue(d.ServerAddress),
		Description:   types.StringValue(d.Description),
	}, nil
}

func mergeBootWriteOnly(prior, read *bootResourceModel) {}
