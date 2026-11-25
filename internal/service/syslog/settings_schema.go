package syslog

import (
	"github.com/browningluke/opnsense-go/pkg/syslog"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const settingsID = "syslog_settings"

type settingsResourceModel struct {
	Enabled     types.Bool  `tfsdk:"enabled"`
	LogLocal    types.Bool  `tfsdk:"log_local"`
	MaxPreserve types.Int64 `tfsdk:"max_preserve"`
	MaxFileSize types.Int64 `tfsdk:"max_file_size"`

	Id types.String `tfsdk:"id"`
}

func settingsResourceSchema() schema.Schema {
	keepBool := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{
			MarkdownDescription: desc + " Unset keeps the current value.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	}
	keepInt := func(desc string, v ...validator.Int64) schema.Int64Attribute {
		return schema.Int64Attribute{
			MarkdownDescription: desc + " Unset keeps the current value.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			Validators:          v,
		}
	}
	return schema.Schema{
		MarkdownDescription: "General logging settings (System > Settings > Logging). " +
			"This is a singleton that always exists, so it must be imported before Terraform manages it (`id = \"" + settingsID + "\"`); " +
			"destroying it only removes it from state. Remote targets are `opnsense_syslog_destination`.",

		Attributes: map[string]schema.Attribute{
			"enabled":       keepBool("Run the local syslog service. Disabling it stops all logging, including remote destinations."),
			"log_local":     keepBool("Enable writing log files to the local disk."),
			"max_preserve":  keepInt("Number of logs to preserve. When no maximum file size is offered or the logs are smaller than the size requested, this equals the number of days. OPNsense defaults to `31`.", int64validator.AtLeast(1)),
			"max_file_size": keepInt("Maximum file size per log file. When set and a log file exceeds the amount specified, it will be rotated. In megabytes (MB); `0` means no limit.", int64validator.AtLeast(0)),
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
		MarkdownDescription: "General logging settings (System > Settings > Logging).",

		Attributes: map[string]dschema.Attribute{
			"enabled":       dschema.BoolAttribute{MarkdownDescription: "Whether the syslog service runs.", Computed: true},
			"log_local":     dschema.BoolAttribute{MarkdownDescription: "Enable writing log files to the local disk.", Computed: true},
			"max_preserve":  dschema.Int64Attribute{MarkdownDescription: "Number of logs to preserve. When no maximum file size is offered or the logs are smaller than the size requested, this equals the number of days.", Computed: true},
			"max_file_size": dschema.Int64Attribute{MarkdownDescription: "Maximum file size per log file, in megabytes (MB). `0` means no limit.", Computed: true},
			"id":            dschema.StringAttribute{MarkdownDescription: "Always `" + settingsID + "`.", Computed: true},
		},
	}
}

func convertSettingsSchemaToStruct(d *settingsResourceModel) *syslog.GeneralSettingsBody {
	size := ""
	if v := d.MaxFileSize.ValueInt64(); v > 0 {
		size = tools.Int64ToString(v)
	}
	return &syslog.GeneralSettingsBody{General: syslog.GeneralSettings{
		Enabled:     tools.BoolToString(d.Enabled.ValueBool()),
		LogLocal:    tools.BoolToString(d.LogLocal.ValueBool()),
		MaxPreserve: tools.Int64ToString(d.MaxPreserve.ValueInt64()),
		MaxFileSize: size,
	}}
}

func convertSettingsStructToSchema(d *syslog.GeneralSettings) *settingsResourceModel {
	size := int64(0) // "" means no limit
	if d.MaxFileSize != "" {
		size = tools.StringToInt64(d.MaxFileSize)
	}
	return &settingsResourceModel{
		Enabled:     types.BoolValue(tools.StringToBool(d.Enabled)),
		LogLocal:    types.BoolValue(tools.StringToBool(d.LogLocal)),
		MaxPreserve: types.Int64Value(tools.StringToInt64(d.MaxPreserve)),
		MaxFileSize: types.Int64Value(size),
		Id:          types.StringValue(settingsID),
	}
}
