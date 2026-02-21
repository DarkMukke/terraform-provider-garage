// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure GarageProvider satisfies various provider interfaces.
var _ provider.Provider = &GarageProvider{}
var _ provider.ProviderWithFunctions = &GarageProvider{}
var _ provider.ProviderWithEphemeralResources = &GarageProvider{}

// GarageProvider defines the provider implementation.
type GarageProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GarageProviderModel describes the provider data model.
type GarageProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"` // deprecated, for objects we can't use the admin endpoint
	Token    types.String `tfsdk:"token"`

	// new structure takes an object for both admin and s3 endpoints
	Endpoints *EndpointsModel `tfsdk:"endpoints"`
	//access keys are needed for s3
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
}

type EndpointsModel struct {
	Admin types.String `tfsdk:"admin"`
	S3    types.String `tfsdk:"s3"`
}

func (p *GarageProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "garage"
	resp.Version = p.version
}

func (p *GarageProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Garage S3 buckets and objects",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:           true,
				DeprecationMessage: "Use 'endpoints' block instead. This attribute will be removed in version 2.0.0",
				Description:        "(Deprecated) Admin API endpoint. Use 'endpoints.admin' instead.",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Admin API token for Garage cluster management",
			},
			"access_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "S3 access key for object operations. Can also be set via GARAGE_ACCESS_KEY environment variable",
			},
			"secret_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "S3 secret key for object operations. Can also be set via GARAGE_SECRET_KEY environment variable",
			},
			"endpoints": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Garage API endpoints configuration",
				Attributes: map[string]schema.Attribute{
					"admin": schema.StringAttribute{
						Optional:    true,
						Description: "Admin API endpoint (e.g., 'http://localhost:3903')",
					},
					"s3": schema.StringAttribute{
						Optional:    true,
						Description: "S3 API endpoint (e.g., 'http://localhost:3900')",
					},
				},
			},
		},
	}
}

func (p *GarageProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config GarageProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle backwards compatibility
	var adminEndpoint, s3Endpoint string

	if config.Endpoints != nil {
		// New endpoints block takes precedence
		if !config.Endpoints.Admin.IsNull() {
			adminEndpoint = config.Endpoints.Admin.ValueString()
		}
		if !config.Endpoints.S3.IsNull() {
			s3Endpoint = config.Endpoints.S3.ValueString()
		}
	}

	// Fall back to deprecated 'endpoint' attribute if endpoints block not used
	if adminEndpoint == "" && !config.Endpoint.IsNull() {
		adminEndpoint = config.Endpoint.ValueString()
		// If using old config, default S3 to port 3900 on same host
		if s3Endpoint == "" {
			// Simple heuristic: replace 3903 with 3900
			s3Endpoint = replacePort(adminEndpoint, "3903", "3900")
		}
	}

	// Environment variable fallback for S3 credentials
	accessKey := config.AccessKey.ValueString()
	if accessKey == "" {
		accessKey = os.Getenv("GARAGE_ACCESS_KEY")
	}

	secretKey := config.SecretKey.ValueString()
	if secretKey == "" {
		secretKey = os.Getenv("GARAGE_SECRET_KEY")
	}

	// Validation
	if adminEndpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Admin Endpoint",
			"Either 'endpoints.admin' or deprecated 'endpoint' must be configured",
		)
		return
	}

	// Store in provider data with both endpoints
	providerData := &GarageProviderModel{
		Endpoint:  types.StringValue(adminEndpoint),
		Token:     config.Token,
		AccessKey: types.StringValue(accessKey),
		SecretKey: types.StringValue(secretKey),
		Endpoints: &EndpointsModel{
			Admin: types.StringValue(adminEndpoint),
			S3:    types.StringValue(s3Endpoint),
		},
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *GarageProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBucketResource,
		NewBucketPermissionResource,
		NewKeyResource,
		NewGarageObjectResource,
	}
}

func (p *GarageProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *GarageProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewBucketDataSource,
		NewGarageObjectDataSource,
	}
}

func (p *GarageProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GarageProvider{
			version: version,
		}
	}
}

// Helper function to replace port in endpoint URL.
func replacePort(endpoint, oldPort, newPort string) string {
	// Simple string replacement - you may want more robust URL parsing
	return strings.Replace(endpoint, ":"+oldPort, ":"+newPort, 1)
}
