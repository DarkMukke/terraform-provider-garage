package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GarageObjectResource{}
var _ resource.ResourceWithImportState = &GarageObjectResource{}

type GarageObjectResource struct {
	s3Client *s3.Client
}

type GarageObjectResourceModel struct {
	Bucket      types.String `tfsdk:"bucket"`
	Key         types.String `tfsdk:"key"`
	Source      types.String `tfsdk:"source"`
	Content     types.String `tfsdk:"content"`
	ContentType types.String `tfsdk:"content_type"`
	ETag        types.String `tfsdk:"etag"`
	ID          types.String `tfsdk:"id"`
}

func NewGarageObjectResource() resource.Resource {
	return &GarageObjectResource{}
}

func (r *GarageObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (r *GarageObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an object in a Garage bucket",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket to store the object",
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "Name of the object in the bucket",
			},
			"source": schema.StringAttribute{
				Optional:    true,
				Description: "Path to a file that will be uploaded",
			},
			"content": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Literal string value to use as object content",
			},
			"content_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "MIME type of the object",
			},
			"etag": schema.StringAttribute{
				Computed:    true,
				Description: "ETag of the object",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (bucket/key)",
			},
		},
	}
}

func (r *GarageObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*GarageProviderModel)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *GarageProviderModel")
		return
	}

	s3Endpoint := providerData.Endpoints.S3.ValueString()
	if s3Endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing S3 Endpoint",
			"S3 endpoint must be configured in endpoints.s3 for object operations",
		)
		return
	}

	// Create S3 client with BaseEndpoint (new method)
	r.s3Client = s3.NewFromConfig(aws.Config{
		Region: "garage",
		Credentials: credentials.NewStaticCredentialsProvider(
			providerData.AccessKey.ValueString(),
			providerData.SecretKey.ValueString(),
			"",
		),
	}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3Endpoint)
		o.UsePathStyle = true // Important for S3-compatible storage like Garage
	})
}

func (r *GarageObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GarageObjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare object content
	var body io.Reader
	var contentType string

	if !plan.Source.IsNull() {
		file, err := os.Open(plan.Source.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("File Read Error", err.Error())
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)
		body = file

		if plan.ContentType.IsNull() {
			contentType = "application/octet-stream"
		} else {
			contentType = plan.ContentType.ValueString()
		}
	} else if !plan.Content.IsNull() {
		body = strings.NewReader(plan.Content.ValueString())
		if plan.ContentType.IsNull() {
			contentType = "text/plain"
		} else {
			contentType = plan.ContentType.ValueString()
		}
	} else {
		resp.Diagnostics.AddError("Missing Content", "Either source or content must be specified")
		return
	}

	// Upload object
	putOutput, err := r.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(plan.Bucket.ValueString()),
		Key:         aws.String(plan.Key.ValueString()),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		resp.Diagnostics.AddError("Object Upload Failed", err.Error())
		return
	}

	// Set computed values
	plan.ID = types.StringValue(plan.Bucket.ValueString() + "/" + plan.Key.ValueString())
	plan.ETag = types.StringValue(*putOutput.ETag)
	plan.ContentType = types.StringValue(contentType)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GarageObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GarageObjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if object exists
	headOutput, err := r.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(state.Bucket.ValueString()),
		Key:    aws.String(state.Key.ValueString()),
	})
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with current metadata
	state.ETag = types.StringValue(*headOutput.ETag)
	if headOutput.ContentType != nil {
		state.ContentType = types.StringValue(*headOutput.ContentType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GarageObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Treat update as delete + create
	r.Create(ctx, resource.CreateRequest{Plan: req.Plan}, (*resource.CreateResponse)(unsafe.Pointer(resp)))
}

func (r *GarageObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GarageObjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(state.Bucket.ValueString()),
		Key:    aws.String(state.Key.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Object Deletion Failed", err.Error())
		return
	}
}

func (r *GarageObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: bucket/key (same as AWS provider)
	// Supports keys with slashes by treating everything after first / as the key
	parts := strings.SplitN(req.ID, "/", 2)

	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'bucket/key', got: %s", req.ID),
		)
		return
	}

	bucket := parts[0]
	key := parts[1]

	// Set the state attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bucket"), bucket)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), key)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
