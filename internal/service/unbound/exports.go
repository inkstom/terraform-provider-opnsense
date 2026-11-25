package unbound

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newAclResource,
		newDnsblResource,
		newDomainOverrideResource,
		newForwardResource,
		newHostAliasResource,
		newHostOverrideResource,
		newSettingsResource,
	}
}

func DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newAclDataSource,
		newDnsblDataSource,
		newDomainOverrideDataSource,
		newForwardDataSource,
		newHostAliasDataSource,
		newHostOverrideDataSource,
		newSettingsDataSource,
	}
}
