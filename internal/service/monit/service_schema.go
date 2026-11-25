package monit

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/monit"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceResourceModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	Pidfile      types.String `tfsdk:"pidfile"`
	Match        types.String `tfsdk:"match"`
	Path         types.String `tfsdk:"path"`
	Timeout      types.String `tfsdk:"timeout"`
	StartTimeout types.String `tfsdk:"start_timeout"`
	Address      types.String `tfsdk:"address"`
	Interface    types.String `tfsdk:"interface"`
	Start        types.String `tfsdk:"start"`
	Stop         types.String `tfsdk:"stop"`
	Tests        types.Set    `tfsdk:"tests"`
	Depends      types.Set    `tfsdk:"depends"`
	PollTime     types.String `tfsdk:"poll_time"`

	Id types.String `tfsdk:"id"`
}

func serviceResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Services monitored by Monit (Services > Monit > Service Settings).",
		Attributes: map[string]schema.Attribute{
			"enabled":       boolDefault("Enable or disable service checks.", true),
			"name":          schema.StringAttribute{MarkdownDescription: "The name of the service. 1 to 255 letters, digits, `_`, `-`, `.` or `$`, unique across services.", Required: true},
			"description":   strDefault("Description of the service.", ""),
			"type":          schema.StringAttribute{MarkdownDescription: "The service check type. One of `process`, `file`, `fifo`, `filesystem`, `directory`, `host` (Remote Host), `system`, `custom` or `network`.", Required: true},
			"pidfile":       strDefault("The PID file of the process. An absolute path, used by `process` services.", ""),
			"match":         strDefault("Find the process by regular expression. Test your pattern with `monit procmatch PATTERN`. Used by `process` services without a PID file.", ""),
			"path":          strDefault("According to the service type path can be a file or a directory. An absolute path, used by `file`, `fifo`, `filesystem`, `directory` and `custom` (the script to run).", ""),
			"timeout":       strDefault("The timeout for custom program checks. Seconds, from `1` to `86400`.", "300"),
			"start_timeout": strDefault("The timeout for custom program to start a service. Seconds, from `0` to `86400`.", "30"),
			"address":       strDefault("The target IP address for 'Remote Host' (`host`) and 'Network' (`network`) checks.", ""),
			"interface":     strDefault("The Interface for 'Network' (`network`) checks.", ""),
			"start":         strDefault("The start script of the service. An absolute path to the executable with its arguments.", ""),
			"stop":          strDefault("The stop script of the service. An absolute path to the executable with its arguments.", ""),
			"tests":         setDefault("This is a list with service tests. UUIDs of `opnsense_monit_test` resources."),
			"depends":       setDefault("Optionally define a (list of) service(s) which are required before monitoring this one, if any of the dependencies are either stopped or unmonitored this service will stop/unmonitor too. UUIDs of other `opnsense_monit_service` resources."),
			"poll_time":     strDefault("Set the service poll time. Either as a number of cycles `NUMBER CYCLES` or Cron-style `* 8-19 * * 1-5`. Leave out the `every` keyword, OPNsense adds it; empty polls every cycle.", ""),
			"id":            idAttr("UUID of the service."),
		},
	}
}

func serviceDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Services monitored by Monit (Services > Monit > Service Settings).",
		Attributes: map[string]dschema.Attribute{
			"id":            dsID("UUID of the service."),
			"enabled":       dsBool("Enable or disable service checks."),
			"name":          dsStr("The name of the service."),
			"description":   dsStr("Description of the service."),
			"type":          dsStr("The service check type."),
			"pidfile":       dsStr("The PID file of the process."),
			"match":         dsStr("Find the process by regular expression."),
			"path":          dsStr("According to the service type path can be a file or a directory."),
			"timeout":       dsStr("The timeout for custom program checks."),
			"start_timeout": dsStr("The timeout for custom program to start a service."),
			"address":       dsStr("The target IP address for 'Remote Host' (`host`) and 'Network' (`network`) checks."),
			"interface":     dsStr("The Interface for 'Network' (`network`) checks."),
			"start":         dsStr("The start script of the service."),
			"stop":          dsStr("The stop script of the service."),
			"tests":         dsSet("This is a list with service tests. UUIDs of `opnsense_monit_test` resources."),
			"depends":       dsSet("Services which are required before monitoring this one. UUIDs of other `opnsense_monit_service` resources."),
			"poll_time":     dsStr("Set the service poll time. Either as a number of cycles `NUMBER CYCLES` or Cron-style `* 8-19 * * 1-5`."),
		},
	}
}

func convertServiceSchemaToStruct(d *serviceResourceModel) (*monit.Service, error) {
	return &monit.Service{
		Enabled:      tools.BoolToString(d.Enabled.ValueBool()),
		Name:         d.Name.ValueString(),
		Description:  d.Description.ValueString(),
		Type:         api.SelectedMap(d.Type.ValueString()),
		Pidfile:      d.Pidfile.ValueString(),
		Match:        d.Match.ValueString(),
		Path:         d.Path.ValueString(),
		Timeout:      d.Timeout.ValueString(),
		Starttimeout: d.StartTimeout.ValueString(),
		Address:      d.Address.ValueString(),
		Interface:    api.SelectedMap(d.Interface.ValueString()),
		Start:        d.Start.ValueString(),
		Stop:         d.Stop.ValueString(),
		Tests:        api.SelectedMapList(tools.SetToStringSlice(d.Tests)),
		Depends:      api.SelectedMapList(tools.SetToStringSlice(d.Depends)),
		Polltime:     d.PollTime.ValueString(),
	}, nil
}

func convertServiceStructToSchema(d *monit.Service) (*serviceResourceModel, error) {
	return &serviceResourceModel{
		Enabled:      types.BoolValue(tools.StringToBool(d.Enabled)),
		Name:         types.StringValue(d.Name),
		Description:  types.StringValue(d.Description),
		Type:         types.StringValue(d.Type.String()),
		Pidfile:      types.StringValue(d.Pidfile),
		Match:        types.StringValue(d.Match),
		Path:         types.StringValue(d.Path),
		Timeout:      types.StringValue(d.Timeout),
		StartTimeout: types.StringValue(d.Starttimeout),
		Address:      types.StringValue(d.Address),
		Interface:    types.StringValue(d.Interface.String()),
		Start:        types.StringValue(d.Start),
		Stop:         types.StringValue(d.Stop),
		Tests:        tools.StringSliceToSet(d.Tests),
		Depends:      tools.StringSliceToSet(d.Depends),
		PollTime:     types.StringValue(d.Polltime),
	}, nil
}

func mergeServiceWriteOnly(prior, read *serviceResourceModel) {}
