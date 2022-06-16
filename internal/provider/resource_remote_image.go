package provider

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io/ioutil"
	"os"
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

	if digest, err := r.forwardImage(data.Source, data.Target); err != nil {
		resp.Diagnostics.AddError("Docker Registry Error", err.Error())
		return
	} else {
		data.Digest = types.String{Value: digest.String()}
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

	_, opts, _, err := r.getCraneReference(data.Target)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Target", err.Error())
		return
	}

	meta, err := crane.Head(data.Target.Value, opts...)
	if err != nil {
		resp.Diagnostics.AddError("Docker Registry Error", fmt.Sprintf("Unable to retrieve image metadata: %s", err))
		return
	}
	data.Digest = types.String{Value: meta.Digest.String()}

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

	if digest, err := r.forwardImage(data.Source, data.Target); err != nil {
		resp.Diagnostics.AddError("Docker Registry Error", err.Error())
		return
	} else {
		data.Digest = types.String{Value: digest.String()}
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

	_, opts, _, err := r.getCraneReference(data.Target)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Target", err.Error())
		return
	}

	if err := crane.Delete(data.Target.Value, opts...); err != nil {
		resp.Diagnostics.AddError("Docker Registry Error", fmt.Sprintf("Unable to delete remote image: %s", err))
		return
	}
}

func (r remoteImageResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("digest"), req, resp)
}

func (r remoteImageResource) getCraneReference(dockerName types.String) (name.Reference, []crane.Option, []remote.Option, error) {
	craneOptions := []crane.Option{}
	remoteOptions := []remote.Option{}

	ref, err := name.ParseReference(dockerName.Value)
	if err != nil {
		return ref, craneOptions, remoteOptions, fmt.Errorf("%q is an invalid docker reference: %s", dockerName.Value, err)
	}
	address := ref.Context().RegistryStr()
	registryAuth := r.provider.FindRegistryAuth(address)
	if registryAuth != nil {
		craneOptions = append(craneOptions, crane.WithAuth(registryAuth))
		remoteOptions = append(remoteOptions, remote.WithAuth(registryAuth))
	}
	return ref, craneOptions, remoteOptions, nil
}

func (r remoteImageResource) forwardImage(src, target types.String) (v1.Hash, error) {
	srcRef, srcCraneOpts, srcRemoteOpts, err := r.getCraneReference(src)
	if err != nil {
		return v1.Hash{}, fmt.Errorf("source %q is an invalid docker reference: %s", src.Value, err)
	}
	targetRef, _, targetRemoteOpts, err := r.getCraneReference(target)
	if err != nil {
		return v1.Hash{}, fmt.Errorf("target %q is an invalid docker reference: %s", target.Value, err)
	}

	// Retrieve metadata about source image and build image map for pulling image
	rmt, err := remote.Get(srcRef, srcRemoteOpts...)
	if err != nil {
		return v1.Hash{}, fmt.Errorf("error retrieving metadata for source image: %s", err)
	}
	img, err := rmt.Image()
	if err != nil {
		return v1.Hash{}, fmt.Errorf("error preparing source image for pull: %s", err)
	}
	imgDigest, err := img.Digest()
	if err != nil {
		return v1.Hash{}, fmt.Errorf("source image is invalid: %s", err)
	}
	imageMap := map[string]v1.Image{src.Value: img}

	// Pull docker image using crane and save it as a tarball to 'path'

	file, err := ioutil.TempFile(".", "tmp_remote_image_*.tgz")
	if err != nil {
		return v1.Hash{}, fmt.Errorf("error creating temporary file for source image: %s", err)
	}
	file.Close() // close immediately to allow pull to work

	path := file.Name()
	if err := crane.MultiSave(imageMap, path, srcCraneOpts...); err != nil {
		return imgDigest, fmt.Errorf("error pulling source image: %s", err)
	}

	// Load image from tarball and push it
	if _, err := os.Stat(path); err != nil {
		return imgDigest, fmt.Errorf("error finding source tarball: %s", err)
	}
	pushImg, err := crane.Load(path)
	if err != nil {
		return imgDigest, fmt.Errorf("loading %s as tarball: %w", path, err)
	}
	if err := remote.Write(targetRef, pushImg, targetRemoteOpts...); err != nil {
		return imgDigest, fmt.Errorf("error pushing image: %w", err)
	}
	return imgDigest, nil
}
