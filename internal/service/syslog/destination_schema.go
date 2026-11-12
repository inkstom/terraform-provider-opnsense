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
	levels.MarkdownDescription = "Levels to send: `debug`, `info`, `notice`, `warn`, `err`, `crit`, `alert`, `emerg`. Empty sends all. Defaults to `[]`."
	facilities := emptySet()
	facilities.MarkdownDescription = "Facilities to send, e.g. `kern`, `auth`, `daemon`, `local0`. Empty sends all. Defaults to `[]`."
	programs := emptySet()
	programs.MarkdownDescription = "Applications to send, e.g. `filterlog`, `unbound`, `sshd`. Empty sends all. Defaults to `[]`."
	return schema.Schema{
		MarkdownDescription: "Remote syslog destinations (System > Settings > Logging > Remote).",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable this destination. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"transport": schema.StringAttribute{
				MarkdownDescription: "Protocol and address family: `udp4`, `tcp4`, `udp6`, `tcp6`, `tls4` or `tls6`.",
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
				MarkdownDescription: "Client certificate reference for `tls4`/`tls6`. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"rfc5424": schema.BoolAttribute{
				MarkdownDescription: "Send RFC 5424 formatted messages instead of BSD (RFC 3164). Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"suppress": schema.BoolAttribute{
				MarkdownDescription: "Suppress repeated messages. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description. Defaults to `\"\"`.",
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
			"enabled":     boolean("Whether the destination is enabled."),
			"transport":   str("Protocol and address family."),
			"hostname":    str("Syslog server."),
			"port":        str("Syslog server port."),
			"levels":      set("Levels sent."),
			"facilities":  set("Facilities sent."),
			"programs":    set("Applications sent."),
			"certificate": str("Client certificate for TLS."),
			"rfc5424":     boolean("Whether RFC 5424 format is used."),
			"suppress":    boolean("Whether repeated messages are suppressed."),
			"description": str("Description."),
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
		Suppress:    tools.BoolToString(d.Suppress.ValueBool()),
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
		Suppress:    types.BoolValue(tools.StringToBool(d.Suppress)),
		Description: types.StringValue(d.Description),
	}, nil
}

func mergeDestinationWriteOnly(prior, read *destinationResourceModel) {}
