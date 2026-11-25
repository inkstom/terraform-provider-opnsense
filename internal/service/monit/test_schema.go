package monit

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/monit"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type testResourceModel struct {
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Condition types.String `tfsdk:"condition"`
	Action    types.String `tfsdk:"action"`
	Path      types.String `tfsdk:"path"`

	Id types.String `tfsdk:"id"`
}

func testResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Monit service tests: a condition and the action to take when it matches (Services > Monit > Service Tests Settings).",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{MarkdownDescription: "The name of the test. 1 to 255 letters, digits, spaces, `_`, `-` or `.`.", Required: true},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the test, e.g. `SystemResource`, `ProcessResource`, `Existence` or `Custom`. The OPNsense UI hides this field. OPNsense derives it from " +
					"`condition` when the condition is recognised, so it is usually best left unset. " +
					"A service only accepts tests of types that suit it (a `process` service takes `ProcessResource` tests, for example), " +
					"and OPNsense refuses to change the type of a test a service uses.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"condition": schema.StringAttribute{MarkdownDescription: "The test condition. E.g. `cpu is greater than 50%` or `failed host 127.0.0.1 port 22 protocol ssh`.", Required: true},
			"action":    schema.StringAttribute{MarkdownDescription: "Simply `alert` or `restart` the service or *execute* a program. One of `alert`, `restart`, `start`, `stop`, `exec` or `unmonitor`.", Required: true},
			"path":      strDefault("Make sure the script is executable by the Monit service. The absolute path of the script to run when `action` is `exec`.", ""),
			"id":        idAttr("UUID of the test."),
		},
	}
}

func testDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Monit service tests (Services > Monit > Service Tests Settings).",
		Attributes: map[string]dschema.Attribute{
			"id":        dsID("UUID of the test."),
			"name":      dsStr("The name of the test."),
			"type":      dsStr("The type of the test."),
			"condition": dsStr("The test condition."),
			"action":    dsStr("Simply `alert` or `restart` the service or *execute* a program."),
			"path":      dsStr("The script to run when `action` is `exec`."),
		},
	}
}

func convertTestSchemaToStruct(d *testResourceModel) (*monit.Test, error) {
	testType := "Custom"
	if !d.Type.IsNull() && !d.Type.IsUnknown() && d.Type.ValueString() != "" {
		testType = d.Type.ValueString()
	}
	return &monit.Test{
		Name:      d.Name.ValueString(),
		Type:      api.SelectedMap(testType),
		Condition: d.Condition.ValueString(),
		Action:    api.SelectedMap(d.Action.ValueString()),
		Path:      d.Path.ValueString(),
	}, nil
}

func convertTestStructToSchema(d *monit.Test) (*testResourceModel, error) {
	return &testResourceModel{
		Name:      types.StringValue(d.Name),
		Type:      types.StringValue(d.Type.String()),
		Condition: types.StringValue(d.Condition),
		Action:    types.StringValue(d.Action.String()),
		Path:      types.StringValue(d.Path),
	}, nil
}

func mergeTestWriteOnly(prior, read *testResourceModel) {}
