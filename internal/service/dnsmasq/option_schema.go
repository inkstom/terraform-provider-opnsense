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
				MarkdownDescription: "`set` option to send it to a client in a DHCP offer or `match` option to dynamically tag clients that send it in the initial DHCP request. One of `set` or `match`. Defaults to `\"set\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("set"),
				Validators:          []validator.String{stringvalidator.OneOf("set", "match")},
			},
			"option":      optionalString("DHCPv4 option to offer to the client. Set as the option number, e.g. `3` (router), `6` (DNS servers), `42` (NTP servers)."),
			"option6":     optionalString("DHCPv6 option to offer to the client. Set as the option number."),
			"interface":   optionalString("This adds a single interface as tag so this DHCP option can match the interface of a DHCP range. Example `opt1`; empty matches any interface."),
			"tags":        optionalStringSet("If the optional tags are given then this option is only sent when all the tags are matched. Can be optionally combined with an interface tag. UUIDs of `opnsense_dnsmasq_tag` resources."),
			"set_tag":     optionalString("Tag to set for requests matching this range which can be used to selectively match DHCP options. UUID of an `opnsense_dnsmasq_tag`; used with `type = \"match\"`."),
			"value":       optionalString("Value (or values) to send to the client. The special address `0.0.0.0` or `[::]` is taken to mean \"the address of the machine running Dnsmasq\". When using `match`, leave empty to match on the option only. Separate multiple values with commas, e.g. `172.30.30.1,172.30.30.2`."),
			"force":       optionalBool("Always send the option, also when the client does not ask for it in the parameter request list. Defaults to `false`.", false),
			"description": optionalString("You may enter a description here for your reference (not parsed)."),
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
