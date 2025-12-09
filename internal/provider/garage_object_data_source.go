package provider

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &GarageObjectDataSource{}

type GarageObjectDataSource struct {
	s3Client *s3.Client
}

type GarageObjectDataSourceModel struct {
	Bucket        types.String `tfsdk:"bucket"`
	Key           types.String `tfsdk:"key"`
	Body          types.String `tfsdk:"body"`
	ContentType   types.String `tfsdk:"content_type"`
	ContentLength types.Int64  `tfsdk:"content_length"`
	ETag          types.String `tfsdk:"etag"`
	LastModified  types.String `tfsdk:"last_modified"`
	Metadata      types.Map    `tfsdk:"metadata"`
	VersionId     types.String `tfsdk:"version_id"`
	ID            types.String `tfsdk:"id"`
}

func NewGarageObjectDataSource() datasource.DataSource {
	return &GarageObjectDataSource{}
}

func (d *GarageObjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (d *GarageObjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an object from a Garage bucket",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket containing the object",
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "Key (name) of the object in the bucket",
			},
			"body": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Object content as a string (use for text files)",
			},
			"content_type": schema.StringAttribute{
				Computed:    true,
				Description: "MIME type of the object",
			},
			"content_length": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the object in bytes",
			},
			"etag": schema.StringAttribute{
				Computed:    true,
				Description: "ETag of the object",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Last modification time of the object",
			},
			"metadata": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "User-defined metadata for the object",
			},
			"version_id": schema.StringAttribute{
				Computed:    true,
				Description: "Version ID of the object (if versioning is enabled)",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (bucket/key)",
			},
		},
	}
}

func (d *GarageObjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*GarageProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *GarageProviderModel",
		)
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

	d.s3Client = s3.NewFromConfig(aws.Config{
		Region: "garage",
		Credentials: credentials.NewStaticCredentialsProvider(
			providerData.AccessKey.ValueString(),
			providerData.SecretKey.ValueString(),
			"",
		),
	}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3Endpoint)
		o.UsePathStyle = true
	})
}

func (d *GarageObjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config GarageObjectDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Download object from Garage
	getOutput, err := d.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.Bucket.ValueString()),
		Key:    aws.String(config.Key.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Retrieve Object",
			"Could not download object from Garage: "+err.Error(),
		)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(getOutput.Body)

	// Read object body
	bodyBytes, err := io.ReadAll(getOutput.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Object Body",
			"Could not read object content: "+err.Error(),
		)
		return
	}

	// Set computed attributes
	config.Body = types.StringValue(string(bodyBytes))
	config.ID = types.StringValue(config.Bucket.ValueString() + "/" + config.Key.ValueString())

	if getOutput.ContentType != nil {
		config.ContentType = types.StringValue(*getOutput.ContentType)
	} else {
		config.ContentType = types.StringValue("application/octet-stream")
	}

	if getOutput.ContentLength != nil {
		config.ContentLength = types.Int64Value(*getOutput.ContentLength)
	} else {
		config.ContentLength = types.Int64Value(int64(len(bodyBytes)))
	}

	if getOutput.ETag != nil {
		config.ETag = types.StringValue(*getOutput.ETag)
	}

	if getOutput.LastModified != nil {
		config.LastModified = types.StringValue(getOutput.LastModified.String())
	}

	if getOutput.VersionId != nil {
		config.VersionId = types.StringValue(*getOutput.VersionId)
	}

	// Convert metadata map
	if len(getOutput.Metadata) > 0 {
		metadataMap := make(map[string]types.String)
		for k, v := range getOutput.Metadata {
			metadataMap[k] = types.StringValue(v)
		}
		config.Metadata, diags = types.MapValueFrom(ctx, types.StringType, metadataMap)
		resp.Diagnostics.Append(diags...)
	} else {
		config.Metadata = types.MapNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
