package interfaces

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newAssignmentResource,
		newVipResource,
		newVlanResource,
		newLoopbackResource,
		newGifResource,
		newGreResource,
		newVxlanResource,
		newNeighborResource,
		newBridgeResource,
	}
}

func DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newAssignmentDataSource,
		newVipDataSource,
		newVlanDataSource,
		newLoopbackDataSource,
		newGifDataSource,
		newGreDataSource,
		newVxlanDataSource,
		newNeighborDataSource,
		newBridgeDataSource,
		newOverviewInterfaceDataSource,
		newOverviewAllDataSource,
	}
}
