package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Logging Reference:
// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog

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
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
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

	if digest, err := r.provider.client.ForwardImage(data.Source.Value, data.Target.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	} else {
		data.Digest = types.String{Value: digest}
		data.Id = types.String{Value: digest}
		tflog.Trace(ctx, fmt.Sprintf("Pushed Image: digest => %s", data.Digest.Value))
	}

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

	meta, err := r.provider.client.GetImageMetadata(data.Target.Value)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	data.Digest = types.String{Value: meta.Digest.String()}
	data.Id = types.String{Value: meta.Digest.String()}

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

	if digest, err := r.provider.client.ForwardImage(data.Source.Value, data.Target.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	} else {
		data.Digest = types.String{Value: digest}
		data.Id = types.String{Value: digest}
		tflog.Trace(ctx, fmt.Sprintf("Pushed Image: digest => %s", data.Digest.Value))
	}

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

	err := r.provider.client.DeleteImageTag(data.Target.Value)
	if err != nil {
		// TODO: Deleting images in AWS doesn't work, returns 404
		//resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("Deleted Image: %s", data.Target.Value))
}

func (r remoteImageResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
