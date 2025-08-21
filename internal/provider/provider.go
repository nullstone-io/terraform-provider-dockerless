package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nullstone-io/terraform-provider-dockerless/dockerless"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ provider.Provider = &dockerlessProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dockerlessProvider{
			version: version,
		}
	}
}

// provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type dockerlessProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (d *dockerlessProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "dockerless"
	response.Version = d.version
}

func (d *dockerlessProvider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"registry_auth": schema.MapNestedAttribute{
				Optional:            true,
				MarkdownDescription: "A map of docker registries and their authentication credentials. Keys are registry endpoints (e.g., ECR proxy endpoint).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Optional:    true,
							Description: "Username for the registry",
						},
						"password": schema.StringAttribute{
							Optional:    true,
							Sensitive:   true,
							Description: "Password for the registry",
						},
					},
				},
			},
		},
		Blocks:              nil,
		Description:         "",
		MarkdownDescription: "",
		DeprecationMessage:  "",
	}
}

func (d *dockerlessProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config dockerlessProviderModel
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Example.Null { /* ... */ }
	registries := map[string]dockerless.RegistryAuth{}
	diags = config.RegistryAuths.ElementsAs(ctx, &registries, true)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	client := &dockerless.Client{
		Registries: registries,
	}
	response.DataSourceData = client
	response.ResourceData = client
}

func (d *dockerlessProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (d *dockerlessProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRemoteImageResource,
	}
}

type dockerlessProviderModel struct {
	RegistryAuths types.Map `tfsdk:"registry_auth"`
}
