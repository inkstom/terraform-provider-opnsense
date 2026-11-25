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
			"enabled":                     keepBool("Enable or disable monit."),
			"interval":                    keepInt("Polling interval in seconds. From `1` to `3600`."),
			"start_delay":                 keepInt("On system boot wait before Monit starts checking services. Seconds, from `0` to `86400`."),
			"mail_servers":                keepSet("List of SMTP servers for alert delivery. Hostnames or IP addresses."),
			"mail_port":                   keepInt("The port of the mail server. Typically `465` for SSL or `25` for TLS and non secure connections."),
			"mail_username":               keepString("The username for mail server authentication."),
			"mail_password":               keepSecret("The password for mail server authentication."),
			"mail_ssl":                    keepBool("Enable encryption for mail server communication."),
			"mail_ssl_version":            keepString("Set the SSL version to use. In `auto` mode, only TLS is used. One of `auto`, `tlsv1`, `tlsv11`, `tlsv12` or `tlsv13`.", stringvalidator.OneOf("auto", "tlsv1", "tlsv11", "tlsv12", "tlsv13")),
			"mail_ssl_verify":             keepBool("Enable or disable SSL server certificate verification."),
			"log_file":                    keepString("The log file of the Monit process. This can be the keyword `syslog` or the path to a file. Empty logs to syslog."),
			"state_file":                  keepString("The state file of the Monit process. An absolute path; empty uses Monit's default."),
			"event_queue_path":            keepString("The path to the eventqueue directory. The eventqueue holds alerts while the mail server is unreachable; empty disables it."),
			"event_queue_slots":           keepInt("The number of eventqueue slots. Only used when `event_queue_path` is set."),
			"httpd_enabled":               keepBool("Start the Monit httpd service. Needed for the Monit status page and M/Monit; requires `httpd_password`."),
			"httpd_username":              keepString("Username for the Monit httpd service."),
			"httpd_password":              keepSecret("Password for the Monit httpd service. Required when `httpd_enabled` is `true`."),
			"httpd_port":                  keepInt("Port of the Monit httpd service, from `1` to `65535`."),
			"httpd_allow":                 keepSet("The `username:password` or host/network etc. for accessing the Monit httpd service. See [Monit Authentication](https://mmonit.com/monit/documentation/monit.html#Authentication)."),
			"mmonit_url":                  keepString("The M/Monit URL, e.g. `https://user:pass@192.168.1.10:8443/collector`. If you want to control Monit services from your M/Monit instance you have to configure the Monit Port too and add corresponding firewall rules as well. Empty disables reporting."),
			"mmonit_timeout":              keepInt("M/Monit timeout in seconds, from `1` to `86400`."),
			"mmonit_register_credentials": keepBool("Automatically register in M/Monit by sending Monit credentials (see `httpd_allow`)."),
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
			"enabled":                     dschema.BoolAttribute{MarkdownDescription: "Enable or disable monit.", Computed: true},
			"interval":                    dschema.Int64Attribute{MarkdownDescription: "Polling interval in seconds.", Computed: true},
			"start_delay":                 dschema.Int64Attribute{MarkdownDescription: "On system boot wait before Monit starts checking services, in seconds.", Computed: true},
			"mail_servers":                dschema.SetAttribute{MarkdownDescription: "List of SMTP servers for alert delivery.", ElementType: types.StringType, Computed: true},
			"mail_port":                   dschema.Int64Attribute{MarkdownDescription: "The port of the mail server. Typically `465` for SSL or `25` for TLS and non secure connections.", Computed: true},
			"mail_username":               dschema.StringAttribute{MarkdownDescription: "The username for mail server authentication.", Computed: true},
			"mail_ssl":                    dschema.BoolAttribute{MarkdownDescription: "Enable encryption for mail server communication.", Computed: true},
			"mail_ssl_version":            dschema.StringAttribute{MarkdownDescription: "Set the SSL version to use. In `auto` mode, only TLS is used. One of `auto`, `tlsv1`, `tlsv11`, `tlsv12` or `tlsv13`.", Computed: true},
			"mail_ssl_verify":             dschema.BoolAttribute{MarkdownDescription: "Enable or disable SSL server certificate verification.", Computed: true},
			"log_file":                    dschema.StringAttribute{MarkdownDescription: "The log file of the Monit process. This can be the keyword `syslog` or the path to a file. Empty logs to syslog.", Computed: true},
			"state_file":                  dschema.StringAttribute{MarkdownDescription: "The state file of the Monit process.", Computed: true},
			"event_queue_path":            dschema.StringAttribute{MarkdownDescription: "The path to the eventqueue directory. The eventqueue holds alerts while the mail server is unreachable; empty disables it.", Computed: true},
			"event_queue_slots":           dschema.Int64Attribute{MarkdownDescription: "The number of eventqueue slots. Only used when `event_queue_path` is set.", Computed: true},
			"httpd_enabled":               dschema.BoolAttribute{MarkdownDescription: "Start the Monit httpd service. Needed for the Monit status page and M/Monit.", Computed: true},
			"httpd_username":              dschema.StringAttribute{MarkdownDescription: "Username for the Monit httpd service.", Computed: true},
			"httpd_port":                  dschema.Int64Attribute{MarkdownDescription: "Port of the Monit httpd service, from `1` to `65535`.", Computed: true},
			"httpd_allow":                 dschema.SetAttribute{MarkdownDescription: "The `username:password` or host/network etc. for accessing the Monit httpd service. See [Monit Authentication](https://mmonit.com/monit/documentation/monit.html#Authentication).", ElementType: types.StringType, Computed: true},
			"mmonit_url":                  dschema.StringAttribute{MarkdownDescription: "The M/Monit URL, e.g. `https://user:pass@192.168.1.10:8443/collector`. If you want to control Monit services from your M/Monit instance you have to configure the Monit Port too and add corresponding firewall rules as well. Empty disables reporting.", Computed: true},
			"mmonit_timeout":              dschema.Int64Attribute{MarkdownDescription: "M/Monit timeout in seconds, from `1` to `86400`.", Computed: true},
			"mmonit_register_credentials": dschema.BoolAttribute{MarkdownDescription: "Automatically register in M/Monit by sending Monit credentials (see `httpd_allow`).", Computed: true},
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
