package ipsec

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newAuthLocalResource,
		newAuthRemoteResource,
		newChildResource,
		newConnectionResource,
		newPskResource,
		newVtiResource,
		newPoolResource,
		newKeyPairResource,
		newManualSpdResource,
		newSettingsResource,
	}
}

func DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newConnectionDataSource,
		newAuthLocalDataSource,
		newAuthRemoteDataSource,
		newChildDataSource,
		newPskDataSource,
		newVtiDataSource,
		newPoolDataSource,
		newKeyPairDataSource,
		newManualSpdDataSource,
		newSettingsDataSource,
	}
}
