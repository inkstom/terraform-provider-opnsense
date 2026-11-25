package monit

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/monit"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type alertResourceModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Recipient   types.String `tfsdk:"recipient"`
	NotOn       types.Bool   `tfsdk:"not_on"`
	Events      types.Set    `tfsdk:"events"`
	Format      types.String `tfsdk:"format"`
	Reminder    types.String `tfsdk:"reminder"`
	Description types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func alertResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Monit e-mail alert recipients (Services > Monit > Alert Settings). Monit's mail server is set in its general settings.",
		Attributes: map[string]schema.Attribute{
			"enabled":     boolDefault("Enable or disable alert.", true),
			"recipient":   schema.StringAttribute{MarkdownDescription: "The email address to send alerts to.", Required: true},
			"not_on":      boolDefault("Do not send alerts for the following events but on all others. The events are those listed in `events`.", false),
			"events":      setDefault("List with events. Leave it empty for all events. Each is one of `action`, `checksum`, `bytein`, `byteout`, `connection`, `content`, `data`, `exec`, `fsflags`, `gid`, `icmp`, `instance`, `invalid`, `link`, `nonexist`, `packetin`, `packetout`, `permission`, `pid`, `ppid`, `resource`, `saturation`, `size`, `speed`, `status`, `timeout`, `timestamp`, `uid` or `uptime`."),
			"format":      strDefault("The email format for alerts, e.g. `Subject: $SERVICE on $HOST failed`. `{` and `}` are not allowed; empty uses Monit's default format.", ""),
			"reminder":    strDefault("Send a reminder after some cycles. A number of cycles from `0` to `86400`; empty or `0` sends the alert only once.", ""),
			"description": strDefault("Description of the alert.", ""),
			"id":          idAttr("UUID of the alert."),
		},
	}
}

func alertDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Monit e-mail alert recipients (Services > Monit > Alert Settings).",
		Attributes: map[string]dschema.Attribute{
			"id":          dsID("UUID of the alert."),
			"enabled":     dsBool("Enable or disable alert."),
			"recipient":   dsStr("The email address to send alerts to."),
			"not_on":      dsBool("Do not send alerts for the following events but on all others. The events are those listed in `events`."),
			"events":      dsSet("List with events. Empty means all events."),
			"format":      dsStr("The email format for alerts."),
			"reminder":    dsStr("Send a reminder after some cycles. Empty or `0` sends the alert only once."),
			"description": dsStr("Description of the alert."),
		},
	}
}

func convertAlertSchemaToStruct(d *alertResourceModel) (*monit.Alert, error) {
	return &monit.Alert{
		Enabled:     tools.BoolToString(d.Enabled.ValueBool()),
		Recipient:   d.Recipient.ValueString(),
		Noton:       tools.BoolToString(d.NotOn.ValueBool()),
		Events:      api.SelectedMapList(tools.SetToStringSlice(d.Events)),
		Format:      d.Format.ValueString(),
		Reminder:    d.Reminder.ValueString(),
		Description: d.Description.ValueString(),
	}, nil
}

func convertAlertStructToSchema(d *monit.Alert) (*alertResourceModel, error) {
	return &alertResourceModel{
		Enabled:     types.BoolValue(tools.StringToBool(d.Enabled)),
		Recipient:   types.StringValue(d.Recipient),
		NotOn:       types.BoolValue(tools.StringToBool(d.Noton)),
		Events:      tools.StringSliceToSet(d.Events),
		Format:      types.StringValue(d.Format),
		Reminder:    types.StringValue(d.Reminder),
		Description: types.StringValue(d.Description),
	}, nil
}

func mergeAlertWriteOnly(prior, read *alertResourceModel) {}
