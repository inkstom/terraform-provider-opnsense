package syslog

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/syslog"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type destinationResourceModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Transport   types.String `tfsdk:"transport"`
	Hostname    types.String `tfsdk:"hostname"`
	Port        types.String `tfsdk:"port"`
	Levels      types.Set    `tfsdk:"levels"`
	Facilities  types.Set    `tfsdk:"facilities"`
	Programs    types.Set    `tfsdk:"programs"`
	Certificate types.String `tfsdk:"certificate"`
	Rfc5424     types.Bool   `tfsdk:"rfc5424"`
	Suppress    types.Bool   `tfsdk:"suppress"`
	Description types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func emptySet() schema.SetAttribute {
	return schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true,
		Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType))}
}

func destinationResourceSchema() schema.Schema {
	levels := emptySet()
	levels.MarkdownDescription = "Choose which levels to include, omit to select all. Values: `debug`, `info`, `notice`, `warn`, `err`, `crit`, `alert`, `emerg`. Defaults to `[]`."
	facilities := emptySet()
	facilities.MarkdownDescription = "Choose which facilities to include, omit to select all. Values include `kern`, `user`, `daemon`, `auth`, `authpriv`, `security`, `local0` to `local7`. Defaults to `[]`."
	programs := emptySet()
	programs.MarkdownDescription = "Choose which applications should be forwarded to the specified target, omit to select all. Values are application names, e.g. `filterlog`, `unbound`, `sshd`. Defaults to `[]`."
	return schema.Schema{
		MarkdownDescription: "Remote syslog destinations (System > Settings > Logging > Remote).",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Set this option to enable this destination. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"transport": schema.StringAttribute{
				MarkdownDescription: "Transport protocol. One of `udp4`, `tcp4`, `udp6`, `tcp6`, `tls4` or `tls6`.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf("udp4", "tcp4", "udp6", "tcp6", "tls4", "tls6")},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Syslog server host name or address.",
				Required:            true,
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Syslog server port. Defaults to `514`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("514"),
			},
			"levels":     levels,
			"facilities": facilities,
			"programs":   programs,
			"certificate": schema.StringAttribute{
				MarkdownDescription: "Transport certificate to use, please make sure to check the general system log when experiencing issues. Error messages can be a bit cryptic from time to time, in which case [this](https://support.oneidentity.com/kb/263658/common-issues-of-tls-encrypted-message-transfer) is a good resource for tracking common issues." +
					" Certificate reference, used with `tls4`/`tls6`. Defaults to `\"\"`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"rfc5424": schema.BoolAttribute{
				MarkdownDescription: "Use RFC5424 formatted messages for this destination. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"suppress": schema.BoolAttribute{
				MarkdownDescription: "Suppress consecutive identical messages for 30 seconds and send a summary with the number of repetitions. Needs OPNsense 26.7 or later; leave unset on older versions, which don't have the setting.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "You may enter a description here for your reference (not parsed). Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
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
	str := func(desc string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	set := func(desc string) dschema.SetAttribute {
		return dschema.SetAttribute{MarkdownDescription: desc, ElementType: types.StringType, Computed: true}
	}
	boolean := func(desc string) dschema.BoolAttribute {
		return dschema.BoolAttribute{MarkdownDescription: desc, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "Remote syslog destinations (System > Settings > Logging > Remote).",

		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{MarkdownDescription: "UUID of the destination.", Required: true},
			"enabled":     boolean("Set this option to enable this destination."),
			"transport":   str("Transport protocol, e.g. `udp4` or `tls6`."),
			"hostname":    str("Syslog server."),
			"port":        str("Syslog server port."),
			"levels":      set("Choose which levels to include, omit to select all."),
			"facilities":  set("Choose which facilities to include, omit to select all."),
			"programs":    set("Choose which applications should be forwarded to the specified target, omit to select all."),
			"certificate": str("Transport certificate to use. Certificate reference, used with `tls4`/`tls6`."),
			"rfc5424":     boolean("Use RFC5424 formatted messages for this destination."),
			"suppress":    boolean("Suppress consecutive identical messages for 30 seconds and send a summary with the number of repetitions."),
			"description": str("You may enter a description here for your reference (not parsed)."),
		},
	}
}

func convertDestinationSchemaToStruct(d *destinationResourceModel) (*syslog.Destination, error) {
	return &syslog.Destination{
		Enabled:     tools.BoolToString(d.Enabled.ValueBool()),
		Transport:   api.SelectedMap(d.Transport.ValueString()),
		Hostname:    d.Hostname.ValueString(),
		Port:        d.Port.ValueString(),
		Level:       api.SelectedMapList(tools.SetToStringSlice(d.Levels)),
		Facility:    api.SelectedMapList(tools.SetToStringSlice(d.Facilities)),
		Program:     api.SelectedMapList(tools.SetToStringSlice(d.Programs)),
		Certificate: api.SelectedMap(d.Certificate.ValueString()),
		Rfc5424:     tools.BoolToString(d.Rfc5424.ValueBool()),
		Suppress:    boolOrEmpty(d.Suppress),
		Description: d.Description.ValueString(),
	}, nil
}

func convertDestinationStructToSchema(d *syslog.Destination) (*destinationResourceModel, error) {
	return &destinationResourceModel{
		Enabled:     types.BoolValue(tools.StringToBool(d.Enabled)),
		Transport:   types.StringValue(d.Transport.String()),
		Hostname:    types.StringValue(d.Hostname),
		Port:        types.StringValue(d.Port),
		Levels:      tools.StringSliceToSet(d.Level),
		Facilities:  tools.StringSliceToSet(d.Facility),
		Programs:    tools.StringSliceToSet(d.Program),
		Certificate: types.StringValue(d.Certificate.String()),
		Rfc5424:     types.BoolValue(tools.StringToBool(d.Rfc5424)),
		Suppress:    emptyOrBool(d.Suppress),
		Description: types.StringValue(d.Description),
	}, nil
}

// boolOrEmpty sends "" (omitted by the client) for an unset optional bool, so
// OPNsense versions without the field are not sent it.
func boolOrEmpty(v types.Bool) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.BoolToString(v.ValueBool())
}

// emptyOrBool maps a field the OPNsense version doesn't have ("") to null.
func emptyOrBool(s string) types.Bool {
	if s == "" {
		return types.BoolNull()
	}
	return types.BoolValue(tools.StringToBool(s))
}

// mergeDestinationWriteOnly keeps suppress unmanaged (null) when it is not
// configured, even on versions that report it.
func mergeDestinationWriteOnly(prior, read *destinationResourceModel) {
	if prior != nil && !prior.Hostname.IsNull() && prior.Suppress.IsNull() {
		read.Suppress = types.BoolNull()
	}
}
