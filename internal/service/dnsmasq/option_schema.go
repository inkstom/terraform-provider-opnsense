package dnsmasq

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/dnsmasq"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type optionResourceModel struct {
	Type        types.String `tfsdk:"type"`
	Option      types.String `tfsdk:"option"`
	Option6     types.String `tfsdk:"option6"`
	Interface   types.String `tfsdk:"interface"`
	Tags        types.Set    `tfsdk:"tags"`
	SetTag      types.String `tfsdk:"set_tag"`
	Value       types.String `tfsdk:"value"`
	Force       types.Bool   `tfsdk:"force"`
	Description types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func optionResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "A dnsmasq DHCP option (Services: Dnsmasq DNS & DHCP: DHCP options). " +
			"`type = \"set\"` sends `value` for `option`/`option6`; `type = \"match\"` tags clients that send it.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "`set` or `match`. Defaults to `\"set\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("set"),
				Validators:          []validator.String{stringvalidator.OneOf("set", "match")},
			},
			"option":      optionalString("DHCPv4 option number, e.g. `3` (router), `6` (DNS servers), `42` (NTP servers)."),
			"option6":     optionalString("DHCPv6 option number."),
			"interface":   optionalString("Only apply on this interface, e.g. `opt1`. Empty applies everywhere."),
			"tags":        optionalStringSet("UUIDs of dnsmasq tags a client must have for this option to apply (`set` type)."),
			"set_tag":     optionalString("UUID of the tag to set on matching clients (`match` type)."),
			"value":       optionalString("Option value, e.g. `172.30.30.1` or a comma-separated list."),
			"force":       optionalBool("Send the option even if the client didn't request it. Defaults to `false`.", false),
			"description": optionalString("Description."),
			"id":          idAttribute("DHCP option"),
		},
	}
}

func convertOptionSchemaToStruct(d *optionResourceModel) (*dnsmasq.Option, error) {
	return &dnsmasq.Option{
		Type:         api.SelectedMap(d.Type.ValueString()),
		OptionV4:     api.SelectedMap(d.Option.ValueString()),
		OptionV6:     api.SelectedMap(d.Option6.ValueString()),
		Interface:    api.SelectedMap(d.Interface.ValueString()),
		TypeSetTags:  tools.SetToStringSlice(d.Tags),
		TypeMatchTag: api.SelectedMap(d.SetTag.ValueString()),
		Value:        d.Value.ValueString(),
		Force:        tools.BoolToString(d.Force.ValueBool()),
		Description:  d.Description.ValueString(),
	}, nil
}

func convertOptionStructToSchema(d *dnsmasq.Option) (*optionResourceModel, error) {
	return &optionResourceModel{
		Type:        types.StringValue(d.Type.String()),
		Option:      types.StringValue(d.OptionV4.String()),
		Option6:     types.StringValue(d.OptionV6.String()),
		Interface:   types.StringValue(d.Interface.String()),
		Tags:        tools.StringSliceToSet(d.TypeSetTags),
		SetTag:      types.StringValue(d.TypeMatchTag.String()),
		Value:       types.StringValue(d.Value),
		Force:       types.BoolValue(tools.StringToBool(d.Force)),
		Description: types.StringValue(d.Description),
	}, nil
}
