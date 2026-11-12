package monit

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/monit"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const settingsID = "monit_settings"

type settingsResourceModel struct {
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Interval                  types.Int64  `tfsdk:"interval"`
	StartDelay                types.Int64  `tfsdk:"start_delay"`
	MailServers               types.Set    `tfsdk:"mail_servers"`
	MailPort                  types.Int64  `tfsdk:"mail_port"`
	MailUsername              types.String `tfsdk:"mail_username"`
	MailPassword              types.String `tfsdk:"mail_password"`
	MailSsl                   types.Bool   `tfsdk:"mail_ssl"`
	MailSslVersion            types.String `tfsdk:"mail_ssl_version"`
	MailSslVerify             types.Bool   `tfsdk:"mail_ssl_verify"`
	LogFile                   types.String `tfsdk:"log_file"`
	StateFile                 types.String `tfsdk:"state_file"`
	EventQueuePath            types.String `tfsdk:"event_queue_path"`
	EventQueueSlots           types.Int64  `tfsdk:"event_queue_slots"`
	HttpdEnabled              types.Bool   `tfsdk:"httpd_enabled"`
	HttpdUsername             types.String `tfsdk:"httpd_username"`
	HttpdPassword             types.String `tfsdk:"httpd_password"`
	HttpdPort                 types.Int64  `tfsdk:"httpd_port"`
	HttpdAllow                types.Set    `tfsdk:"httpd_allow"`
	MmonitUrl                 types.String `tfsdk:"mmonit_url"`
	MmonitTimeout             types.Int64  `tfsdk:"mmonit_timeout"`
	MmonitRegisterCredentials types.Bool   `tfsdk:"mmonit_register_credentials"`

	Id types.String `tfsdk:"id"`
}

type settingsDataSourceModel struct {
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Interval                  types.Int64  `tfsdk:"interval"`
	StartDelay                types.Int64  `tfsdk:"start_delay"`
	MailServers               types.Set    `tfsdk:"mail_servers"`
	MailPort                  types.Int64  `tfsdk:"mail_port"`
	MailUsername              types.String `tfsdk:"mail_username"`
	MailSsl                   types.Bool   `tfsdk:"mail_ssl"`
	MailSslVersion            types.String `tfsdk:"mail_ssl_version"`
	MailSslVerify             types.Bool   `tfsdk:"mail_ssl_verify"`
	LogFile                   types.String `tfsdk:"log_file"`
	StateFile                 types.String `tfsdk:"state_file"`
	EventQueuePath            types.String `tfsdk:"event_queue_path"`
	EventQueueSlots           types.Int64  `tfsdk:"event_queue_slots"`
	HttpdEnabled              types.Bool   `tfsdk:"httpd_enabled"`
	HttpdUsername             types.String `tfsdk:"httpd_username"`
	HttpdPort                 types.Int64  `tfsdk:"httpd_port"`
	HttpdAllow                types.Set    `tfsdk:"httpd_allow"`
	MmonitUrl                 types.String `tfsdk:"mmonit_url"`
	MmonitTimeout             types.Int64  `tfsdk:"mmonit_timeout"`
	MmonitRegisterCredentials types.Bool   `tfsdk:"mmonit_register_credentials"`

	Id types.String `tfsdk:"id"`
}

// Every setting is optional: leave one unset and its current value on the
// firewall is kept, not managed.
const keepNote = " Unset keeps the current value."

func keepBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{MarkdownDescription: desc + keepNote, Optional: true, Computed: true,
		PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}}
}

func keepInt(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{MarkdownDescription: desc + keepNote, Optional: true, Computed: true,
		PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}}
}

func keepString(desc string, v ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: desc + keepNote, Optional: true, Computed: true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: v}
}

func keepSecret(desc string) schema.StringAttribute {
	a := keepString(desc)
	a.Sensitive = true
	return a
}

func keepSet(desc string) schema.SetAttribute {
	return schema.SetAttribute{MarkdownDescription: desc + keepNote, ElementType: types.StringType, Optional: true, Computed: true,
		PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()}}
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Monit general settings (Services > Monit > Settings): polling, the alert mail server, the local web interface and M/Monit. " +
			"This is a singleton that always exists, so it must be imported before Terraform manages it (`id = \"" + settingsID + "\"`); " +
			"destroying it only removes it from state.",

		Attributes: map[string]schema.Attribute{
			"enabled":                     keepBool("Run Monit."),
			"interval":                    keepInt("Seconds between checks."),
			"start_delay":                 keepInt("Seconds to wait after boot before the first check."),
			"mail_servers":                keepSet("Mail servers for alerts, tried in order of preference."),
			"mail_port":                   keepInt("Mail server port."),
			"mail_username":               keepString("Mail server user."),
			"mail_password":               keepSecret("Mail server password."),
			"mail_ssl":                    keepBool("Use TLS to the mail server."),
			"mail_ssl_version":            keepString("TLS version: `auto`, `tlsv1`, `tlsv11`, `tlsv12` or `tlsv13`.", stringvalidator.OneOf("auto", "tlsv1", "tlsv11", "tlsv12", "tlsv13")),
			"mail_ssl_verify":             keepBool("Verify the mail server's certificate."),
			"log_file":                    keepString("Log file; empty logs to syslog."),
			"state_file":                  keepString("State file; empty uses the default."),
			"event_queue_path":            keepString("Directory that queues alerts while the mail server is unreachable; empty disables the queue."),
			"event_queue_slots":           keepInt("Maximum queued alerts."),
			"httpd_enabled":               keepBool("Run Monit's local web interface (needed for the status page and M/Monit)."),
			"httpd_username":              keepString("Web interface user."),
			"httpd_password":              keepSecret("Web interface password."),
			"httpd_port":                  keepInt("Web interface port."),
			"httpd_allow":                 keepSet("Hosts or networks allowed to reach the web interface."),
			"mmonit_url":                  keepString("M/Monit server URL; empty disables reporting."),
			"mmonit_timeout":              keepInt("M/Monit timeout in seconds."),
			"mmonit_register_credentials": keepBool("Register the web interface credentials with M/Monit."),
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `" + settingsID + "`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func settingsDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Monit general settings (Services > Monit > Settings). Passwords are not exposed.",

		Attributes: map[string]dschema.Attribute{
			"enabled":                     dschema.BoolAttribute{MarkdownDescription: "Run Monit.", Computed: true},
			"interval":                    dschema.Int64Attribute{MarkdownDescription: "Seconds between checks.", Computed: true},
			"start_delay":                 dschema.Int64Attribute{MarkdownDescription: "Seconds to wait after boot before the first check.", Computed: true},
			"mail_servers":                dschema.SetAttribute{MarkdownDescription: "Mail servers for alerts, tried in order of preference.", ElementType: types.StringType, Computed: true},
			"mail_port":                   dschema.Int64Attribute{MarkdownDescription: "Mail server port.", Computed: true},
			"mail_username":               dschema.StringAttribute{MarkdownDescription: "Mail server user.", Computed: true},
			"mail_ssl":                    dschema.BoolAttribute{MarkdownDescription: "Use TLS to the mail server.", Computed: true},
			"mail_ssl_version":            dschema.StringAttribute{MarkdownDescription: "TLS version: `auto`, `tlsv1`, `tlsv11`, `tlsv12` or `tlsv13`.", Computed: true},
			"mail_ssl_verify":             dschema.BoolAttribute{MarkdownDescription: "Verify the mail server's certificate.", Computed: true},
			"log_file":                    dschema.StringAttribute{MarkdownDescription: "Log file.", Computed: true},
			"state_file":                  dschema.StringAttribute{MarkdownDescription: "State file.", Computed: true},
			"event_queue_path":            dschema.StringAttribute{MarkdownDescription: "Directory that queues alerts while the mail server is unreachable.", Computed: true},
			"event_queue_slots":           dschema.Int64Attribute{MarkdownDescription: "Maximum queued alerts.", Computed: true},
			"httpd_enabled":               dschema.BoolAttribute{MarkdownDescription: "Run Monit's local web interface (needed for the status page and M/Monit).", Computed: true},
			"httpd_username":              dschema.StringAttribute{MarkdownDescription: "Web interface user.", Computed: true},
			"httpd_port":                  dschema.Int64Attribute{MarkdownDescription: "Web interface port.", Computed: true},
			"httpd_allow":                 dschema.SetAttribute{MarkdownDescription: "Hosts or networks allowed to reach the web interface.", ElementType: types.StringType, Computed: true},
			"mmonit_url":                  dschema.StringAttribute{MarkdownDescription: "M/Monit server URL.", Computed: true},
			"mmonit_timeout":              dschema.Int64Attribute{MarkdownDescription: "M/Monit timeout in seconds.", Computed: true},
			"mmonit_register_credentials": dschema.BoolAttribute{MarkdownDescription: "Register the web interface credentials with M/Monit.", Computed: true},
			"id":                          dschema.StringAttribute{MarkdownDescription: "Always `" + settingsID + "`.", Computed: true},
		},
	}
}

func boolOrEmpty(v types.Bool) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.BoolToString(v.ValueBool())
}

func intOrEmpty(v types.Int64) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.Int64ToString(v.ValueInt64())
}

func stringOrEmpty(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func setOrNil(v types.Set) []string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return tools.SetToStringSlice(v)
}

func convertSettingsSchemaToStruct(d *settingsResourceModel) *monit.SettingsGeneral {
	return &monit.SettingsGeneral{General: monit.Settings{
		Enabled:                   boolOrEmpty(d.Enabled),
		Interval:                  intOrEmpty(d.Interval),
		Startdelay:                intOrEmpty(d.StartDelay),
		Mailserver:                api.SelectedMapList(setOrNil(d.MailServers)),
		Port:                      intOrEmpty(d.MailPort),
		Username:                  stringOrEmpty(d.MailUsername),
		Password:                  stringOrEmpty(d.MailPassword),
		Ssl:                       boolOrEmpty(d.MailSsl),
		SslVersion:                api.SelectedMap(stringOrEmpty(d.MailSslVersion)),
		SslVerify:                 boolOrEmpty(d.MailSslVerify),
		Logfile:                   stringOrEmpty(d.LogFile),
		Statefile:                 stringOrEmpty(d.StateFile),
		EventqueuePath:            stringOrEmpty(d.EventQueuePath),
		EventqueueSlots:           intOrEmpty(d.EventQueueSlots),
		HttpdEnabled:              boolOrEmpty(d.HttpdEnabled),
		HttpdUsername:             stringOrEmpty(d.HttpdUsername),
		HttpdPassword:             stringOrEmpty(d.HttpdPassword),
		HttpdPort:                 intOrEmpty(d.HttpdPort),
		HttpdAllow:                api.SelectedMapList(setOrNil(d.HttpdAllow)),
		MmonitUrl:                 stringOrEmpty(d.MmonitUrl),
		MmonitTimeout:             intOrEmpty(d.MmonitTimeout),
		MmonitRegisterCredentials: boolOrEmpty(d.MmonitRegisterCredentials),
	}}
}

func convertSettingsStructToSchema(d *monit.Settings) *settingsResourceModel {
	return &settingsResourceModel{
		Enabled:                   types.BoolValue(tools.StringToBool(d.Enabled)),
		Interval:                  tools.StringToInt64Null(d.Interval),
		StartDelay:                tools.StringToInt64Null(d.Startdelay),
		MailServers:               tools.StringSliceToSet(d.Mailserver),
		MailPort:                  tools.StringToInt64Null(d.Port),
		MailUsername:              types.StringValue(d.Username),
		MailPassword:              types.StringValue(d.Password),
		MailSsl:                   types.BoolValue(tools.StringToBool(d.Ssl)),
		MailSslVersion:            types.StringValue(d.SslVersion.String()),
		MailSslVerify:             types.BoolValue(tools.StringToBool(d.SslVerify)),
		LogFile:                   types.StringValue(d.Logfile),
		StateFile:                 types.StringValue(d.Statefile),
		EventQueuePath:            types.StringValue(d.EventqueuePath),
		EventQueueSlots:           tools.StringToInt64Null(d.EventqueueSlots),
		HttpdEnabled:              types.BoolValue(tools.StringToBool(d.HttpdEnabled)),
		HttpdUsername:             types.StringValue(d.HttpdUsername),
		HttpdPassword:             types.StringValue(d.HttpdPassword),
		HttpdPort:                 tools.StringToInt64Null(d.HttpdPort),
		HttpdAllow:                tools.StringSliceToSet(d.HttpdAllow),
		MmonitUrl:                 types.StringValue(d.MmonitUrl),
		MmonitTimeout:             tools.StringToInt64Null(d.MmonitTimeout),
		MmonitRegisterCredentials: types.BoolValue(tools.StringToBool(d.MmonitRegisterCredentials)),
		Id:                        types.StringValue(settingsID),
	}
}

func convertSettingsStructToDataSource(d *monit.Settings) *settingsDataSourceModel {
	return &settingsDataSourceModel{
		Enabled:                   types.BoolValue(tools.StringToBool(d.Enabled)),
		Interval:                  tools.StringToInt64Null(d.Interval),
		StartDelay:                tools.StringToInt64Null(d.Startdelay),
		MailServers:               tools.StringSliceToSet(d.Mailserver),
		MailPort:                  tools.StringToInt64Null(d.Port),
		MailUsername:              types.StringValue(d.Username),
		MailSsl:                   types.BoolValue(tools.StringToBool(d.Ssl)),
		MailSslVersion:            types.StringValue(d.SslVersion.String()),
		MailSslVerify:             types.BoolValue(tools.StringToBool(d.SslVerify)),
		LogFile:                   types.StringValue(d.Logfile),
		StateFile:                 types.StringValue(d.Statefile),
		EventQueuePath:            types.StringValue(d.EventqueuePath),
		EventQueueSlots:           tools.StringToInt64Null(d.EventqueueSlots),
		HttpdEnabled:              types.BoolValue(tools.StringToBool(d.HttpdEnabled)),
		HttpdUsername:             types.StringValue(d.HttpdUsername),
		HttpdPort:                 tools.StringToInt64Null(d.HttpdPort),
		HttpdAllow:                tools.StringSliceToSet(d.HttpdAllow),
		MmonitUrl:                 types.StringValue(d.MmonitUrl),
		MmonitTimeout:             tools.StringToInt64Null(d.MmonitTimeout),
		MmonitRegisterCredentials: types.BoolValue(tools.StringToBool(d.MmonitRegisterCredentials)),
		Id:                        types.StringValue(settingsID),
	}
}
