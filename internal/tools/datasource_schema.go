package tools

import (
	"fmt"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// DataSourceSchemaFromResource derives a data source schema from a resource
// schema: every attribute becomes Computed, except the ones named in required,
// which become Required inputs (typically "id"). Plan modifiers, defaults and
// validators are dropped. It panics on attribute types it does not handle so
// that a new nested attribute fails loudly in tests rather than silently.
func DataSourceSchemaFromResource(rs schema.Schema, required ...string) dschema.Schema {
	req := map[string]bool{}
	for _, r := range required {
		req[r] = true
	}
	attrs := map[string]dschema.Attribute{}
	for name, a := range rs.Attributes {
		desc := a.GetMarkdownDescription()
		comp, reqd := !req[name], req[name]
		switch v := a.(type) {
		case schema.StringAttribute:
			attrs[name] = dschema.StringAttribute{MarkdownDescription: desc, Computed: comp, Required: reqd, Sensitive: v.Sensitive}
		case schema.BoolAttribute:
			attrs[name] = dschema.BoolAttribute{MarkdownDescription: desc, Computed: comp, Required: reqd, Sensitive: v.Sensitive}
		case schema.Int64Attribute:
			attrs[name] = dschema.Int64Attribute{MarkdownDescription: desc, Computed: comp, Required: reqd, Sensitive: v.Sensitive}
		case schema.Float64Attribute:
			attrs[name] = dschema.Float64Attribute{MarkdownDescription: desc, Computed: comp, Required: reqd, Sensitive: v.Sensitive}
		case schema.SetAttribute:
			attrs[name] = dschema.SetAttribute{MarkdownDescription: desc, ElementType: v.ElementType, Computed: comp, Required: reqd, Sensitive: v.Sensitive}
		case schema.ListAttribute:
			attrs[name] = dschema.ListAttribute{MarkdownDescription: desc, ElementType: v.ElementType, Computed: comp, Required: reqd, Sensitive: v.Sensitive}
		default:
			panic(fmt.Sprintf("DataSourceSchemaFromResource: attribute %q has unsupported type %T", name, a))
		}
	}
	return dschema.Schema{MarkdownDescription: rs.MarkdownDescription, Attributes: attrs}
}
