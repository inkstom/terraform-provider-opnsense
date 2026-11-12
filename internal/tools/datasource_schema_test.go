package tools

import (
	"testing"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataSourceSchemaFromResource(t *testing.T) {
	rs := schema.Schema{
		MarkdownDescription: "thing",
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true, MarkdownDescription: "UUID."},
			"name":   schema.StringAttribute{Required: true, MarkdownDescription: "Name."},
			"secret": schema.StringAttribute{Optional: true, Sensitive: true},
			"on":     schema.BoolAttribute{Optional: true},
			"port":   schema.Int64Attribute{Optional: true},
			"tags":   schema.SetAttribute{Optional: true, ElementType: types.StringType},
		},
	}
	ds := DataSourceSchemaFromResource(rs, "id")

	if ds.MarkdownDescription != "thing" || len(ds.Attributes) != len(rs.Attributes) {
		t.Fatalf("got %d attributes, description %q", len(ds.Attributes), ds.MarkdownDescription)
	}
	id := ds.Attributes["id"].(dschema.StringAttribute)
	if !id.Required || id.Computed || id.MarkdownDescription != "UUID." {
		t.Errorf("id = %+v, want required input", id)
	}
	if n := ds.Attributes["name"].(dschema.StringAttribute); n.Required || !n.Computed {
		t.Errorf("name = %+v, want computed", n)
	}
	if s := ds.Attributes["secret"].(dschema.StringAttribute); !s.Sensitive {
		t.Error("secret lost Sensitive")
	}
	if s := ds.Attributes["tags"].(dschema.SetAttribute); !s.Computed || s.ElementType != types.StringType {
		t.Errorf("tags = %+v", s)
	}
	_ = ds.Attributes["on"].(dschema.BoolAttribute)
	_ = ds.Attributes["port"].(dschema.Int64Attribute)
}

func TestDataSourceSchemaFromResourcePanicsOnNested(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for a nested attribute")
		}
	}()
	DataSourceSchemaFromResource(schema.Schema{Attributes: map[string]schema.Attribute{
		"nested": schema.SingleNestedAttribute{Optional: true, Attributes: map[string]schema.Attribute{}},
	}})
}
