package provider

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = remoteImageResourceType{}
var _ tfsdk.Resource = remoteImageResource{}
var _ tfsdk.ResourceWithImportState = remoteImageResource{}

type remoteImageResourceType struct{}

func (t remoteImageResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `This resource pushes an image to a target docker image repository from a remote repository without using the docker daemon.`,

		Attributes: map[string]tfsdk.Attribute{
			"source": {
				MarkdownDescription: `The docker image name and tag to source for pushing to the target image repository. 
Currently, this docker image must be public or accessible using the same auth as the "target" image repository.`,
				Required: true,
				Type:     types.StringType,
			},
			"target": {
				MarkdownDescription: "The docker image name and tag to ensure exists in an image repository.",
				Required:            true,
				Type:                types.StringType,
			},
			"digest": {
				MarkdownDescription: "The digest of the target docker image.",
				Computed:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func (t remoteImageResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return remoteImageResource{
		provider: provider,
	}, diags
}

type exampleResourceData struct {
	Id     types.String `tfsdk:"id"`
	Source types.String `tfsdk:"source"`
	Target types.String `tfsdk:"target"`
	Digest types.String `tfsdk:"digest"`
}

type remoteImageResource struct {
	provider provider
}

func (r remoteImageResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data exampleResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.CreateExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	//data.Id = types.String{Value: "example-id"}

	// TODO: Implement

	// write logs using the tflog package
	// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	// for more information
	tflog.Trace(ctx, "created a resource")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r remoteImageResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data exampleResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	authenticator, err := r.getTargetCraneAuth(data.Target)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Target", err.Error())
		return
	}

	meta, err := crane.Head(data.Target.Value, authenticator)
	if err != nil {
		resp.Diagnostics.AddError("Docker Registry Error", fmt.Sprintf("Unable to retrieve image metadata: %s", err))
		return
	}
	data.Digest = types.String{Value: meta.Digest.String()}
	data.Id = types.String{Value: meta.Digest.String()}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.ReadExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r remoteImageResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data exampleResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.UpdateExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// TODO: Implement

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r remoteImageResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data exampleResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	authenticator, err := r.getTargetCraneAuth(data.Target)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Target", err.Error())
		return
	}

	if err := crane.Delete(data.Target.Value, authenticator); err != nil {
		resp.Diagnostics.AddError("Docker Registry Error", fmt.Sprintf("Unable to delete remote image: %s", err))
		return
	}
}

func (r remoteImageResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func (r remoteImageResource) getTargetCraneAuth(target types.String) (crane.Option, error) {
	ref, err := name.ParseReference(target.Value)
	if err != nil {
		return func(*crane.Options) {}, fmt.Errorf("target is an invalid docker reference: %s", err)
	}
	address := ref.Context().RegistryStr()
	registryAuth := r.provider.FindRegistryAuth(address)
	if registryAuth != nil {
		return crane.WithAuth(registryAuth), nil
	}
	return func(*crane.Options) {}, nil
}
