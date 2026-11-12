package core

import (
	"github.com/browningluke/opnsense-go/pkg/core"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tunableResourceModel struct {
	Tunable      types.String `tfsdk:"tunable"`
	Value        types.String `tfsdk:"value"`
	Description  types.String `tfsdk:"description"`
	DefaultValue types.String `tfsdk:"default_value"`
	Type         types.String `tfsdk:"type"`

	Id types.String `tfsdk:"id"`
}

func tunableResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "System tunables: sysctl values and boot-time loader settings (System > Settings > Tunables). " +
			"OPNsense ships many tunables already; import an existing one by UUID instead of creating a duplicate. " +
			"Destroying a tunable removes it from the configuration but does not reset the running value until reboot.",

		Attributes: map[string]schema.Attribute{
			"tunable": schema.StringAttribute{
				MarkdownDescription: "Tunable name, e.g. `net.inet.tcp.blackhole`.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value to set. Empty uses the default value. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"default_value": schema.StringAttribute{
				MarkdownDescription: "The system default for this tunable, if OPNsense knows it.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "`w` for a runtime sysctl, `t` for a boot-time tunable, empty when unknown.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the tunable.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func tunableDataSourceSchema() dschema.Schema {
	str := func(desc string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "System tunables (System > Settings > Tunables).",

		Attributes: map[string]dschema.Attribute{
			"id":            dschema.StringAttribute{MarkdownDescription: "UUID of the tunable.", Required: true},
			"tunable":       str("Tunable name."),
			"value":         str("Configured value; empty means the default."),
			"description":   str("Description."),
			"default_value": str("System default."),
			"type":          str("`w` runtime sysctl, `t` boot-time tunable."),
		},
	}
}

func convertTunableSchemaToStruct(d *tunableResourceModel) (*core.Tunable, error) {
	return &core.Tunable{
		Tunable:     d.Tunable.ValueString(),
		Value:       d.Value.ValueString(),
		Description: d.Description.ValueString(),
	}, nil
}

func convertTunableStructToSchema(d *core.Tunable) (*tunableResourceModel, error) {
	return &tunableResourceModel{
		Tunable:      types.StringValue(d.Tunable),
		Value:        types.StringValue(d.Value),
		Description:  types.StringValue(d.Description),
		DefaultValue: types.StringValue(d.DefaultValue),
		Type:         types.StringValue(d.Type),
	}, nil
}

func mergeTunableWriteOnly(prior, read *tunableResourceModel) {}
