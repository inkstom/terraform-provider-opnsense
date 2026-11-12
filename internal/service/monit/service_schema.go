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
			"enabled":       boolDefault("Enable monitoring of this service.", true),
			"name":          schema.StringAttribute{MarkdownDescription: "Service name (letters, digits, `_` and `-`).", Required: true},
			"description":   strDefault("Description.", ""),
			"type":          schema.StringAttribute{MarkdownDescription: "`process`, `file`, `fifo`, `filesystem`, `directory`, `host`, `system`, `custom` or `network`.", Required: true},
			"pidfile":       strDefault("PID file, for `process`.", ""),
			"match":         strDefault("Process name pattern, for `process` without a PID file.", ""),
			"path":          strDefault("Path for `file`, `fifo`, `filesystem`, `directory` and `custom` (the script to run).", ""),
			"timeout":       strDefault("Program timeout in seconds, for `custom`.", "300"),
			"start_timeout": strDefault("Start timeout in seconds.", "30"),
			"address":       strDefault("Address for `host`.", ""),
			"interface":     strDefault("Interface for `network`.", ""),
			"start":         strDefault("Start command.", ""),
			"stop":          strDefault("Stop command.", ""),
			"tests":         setDefault("UUIDs of `opnsense_monit_test`s to run."),
			"depends":       setDefault("UUIDs of other `opnsense_monit_service`s this one depends on."),
			"poll_time":     strDefault("Polling schedule, e.g. `every 2 cycles` or a cron expression; empty polls every cycle.", ""),
			"id":            idAttr("UUID of the service."),
		},
	}
}

func serviceDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Services monitored by Monit (Services > Monit > Service Settings).",
		Attributes: map[string]dschema.Attribute{
			"id":            dsID("UUID of the service."),
			"enabled":       dsBool("Whether monitoring is enabled."),
			"name":          dsStr("Service name."),
			"description":   dsStr("Description."),
			"type":          dsStr("Service type."),
			"pidfile":       dsStr("PID file."),
			"match":         dsStr("Process name pattern."),
			"path":          dsStr("Path."),
			"timeout":       dsStr("Program timeout."),
			"start_timeout": dsStr("Start timeout."),
			"address":       dsStr("Host address."),
			"interface":     dsStr("Network interface."),
			"start":         dsStr("Start command."),
			"stop":          dsStr("Stop command."),
			"tests":         dsSet("Test UUIDs."),
			"depends":       dsSet("Dependency UUIDs."),
			"poll_time":     dsStr("Polling schedule."),
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
