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
			"enabled":     boolDefault("Enable this alert.", true),
			"recipient":   schema.StringAttribute{MarkdownDescription: "E-mail address to notify.", Required: true},
			"not_on":      boolDefault("Invert `events`: alert on every event except those listed.", false),
			"events":      setDefault("Events to alert on, e.g. `nonexist`, `resource`, `status`, `connection`. Empty alerts on all."),
			"format":      strDefault("Mail format override, e.g. `subject: $SERVICE $EVENT at $DATE`.", ""),
			"reminder":    strDefault("Re-send the alert every N cycles while the problem persists; empty sends once.", ""),
			"description": strDefault("Description.", ""),
			"id":          idAttr("UUID of the alert."),
		},
	}
}

func alertDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Monit e-mail alert recipients (Services > Monit > Alert Settings).",
		Attributes: map[string]dschema.Attribute{
			"id":          dsID("UUID of the alert."),
			"enabled":     dsBool("Whether the alert is enabled."),
			"recipient":   dsStr("E-mail address."),
			"not_on":      dsBool("Whether `events` is inverted."),
			"events":      dsSet("Events alerted on."),
			"format":      dsStr("Mail format override."),
			"reminder":    dsStr("Reminder interval in cycles."),
			"description": dsStr("Description."),
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
