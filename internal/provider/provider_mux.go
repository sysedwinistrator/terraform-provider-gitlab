package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/sdk"
)

func NewMuxedProviderServer(ctx context.Context, version string) (func() tfprotov6.ProviderServer, error) {
	sdkProvider, err := sdk.NewV6(ctx, version)
	if err != nil {
		return nil, err
	}

	providers := []func() tfprotov6.ProviderServer{
		// SDKv2 provider server
		func() tfprotov6.ProviderServer { return sdkProvider },
		// Framework provider server
		providerserver.NewProtocol6(New(version)()),
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		return nil, err
	}
	return muxServer.ProviderServer, nil
}
