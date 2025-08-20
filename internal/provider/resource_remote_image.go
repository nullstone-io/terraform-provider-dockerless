package provider

import (
	"context"
	"fmt"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nullstone-io/terraform-provider-dockerless/dockerless"
)

// Logging Reference:
// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &remoteImageResource{}
	_ resource.ResourceWithConfigure   = &remoteImageResource{}
	_ resource.ResourceWithImportState = &remoteImageResource{}
)

func NewRemoteImageResource() resource.Resource {
	return &remoteImageResource{}
}

type remoteImageResourceModel struct {
	Id     types.String `tfsdk:"id"`
	Source types.String `tfsdk:"source"`
	Target types.String `tfsdk:"target"`
	Digest types.String `tfsdk:"digest"`
}

type remoteImageResource struct {
	client *dockerless.Client
}

func (r *remoteImageResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*dockerless.Client)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *dockerless.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *remoteImageResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_remote_image"
}

func (r *remoteImageResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `This resource pushes an image to a target docker image repository from a remote repository without using the docker daemon.`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"source": schema.StringAttribute{
				MarkdownDescription: `The docker image name and tag to source for pushing to the target image repository. 
Currently, this docker image must be public or accessible using the same auth as the "target" image repository.`,
				Required: true,
			},
			"target": schema.StringAttribute{
				MarkdownDescription: "The docker image name and tag to ensure exists in an image repository.",
				Required:            true,
			},
			"digest": schema.StringAttribute{
				MarkdownDescription: "The digest of the target docker image.",
				Computed:            true,
			},
		},
	}
}

func (r *remoteImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan remoteImageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if digest, err := r.client.ForwardImage(plan.Source.ValueString(), plan.Target.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	} else {
		plan.Digest = types.StringValue(digest)
		plan.Id = types.StringValue(digest)
		tflog.Trace(ctx, fmt.Sprintf("Pushed Image: digest => %s", plan.Digest.ValueString()))
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *remoteImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state remoteImageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.client.GetImageMetadata(state.Target.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	state.Digest = types.StringValue(meta.Digest.String())
	state.Id = types.StringValue(meta.Digest.String())

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *remoteImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan remoteImageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if digest, err := r.client.ForwardImage(plan.Source.ValueString(), plan.Target.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	} else {
		plan.Digest = types.StringValue(digest)
		plan.Id = types.StringValue(digest)
		tflog.Trace(ctx, fmt.Sprintf("Pushed Image: digest => %s", plan.Digest.ValueString()))
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *remoteImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state remoteImageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteImageTag(state.Target.ValueString())
	if err != nil {
		// TODO: Deleting images in AWS doesn't work, returns 404
		//resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("Deleted Image: %s", state.Target.ValueString()))
}

func (r *remoteImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, tfpath.Root("id"), req, resp)
}
