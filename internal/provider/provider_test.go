// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"garage": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the garage provider.
var _ = map[string]func() (tfprotov6.ProviderServer, error){
	"garage": providerserver.NewProtocol6WithError(New("test")()),
	"echo":   echoprovider.NewProviderServer(),
}

func testAccPreCheck(t *testing.T) {
	// Admin API configuration
	if v := os.Getenv("GARAGE_ADMIN_ENDPOINT"); v == "" {
		t.Fatal("GARAGE_ADMIN_ENDPOINT must be set for acceptance tests")
	}
}

func testAccPreCheckS3(t *testing.T) {
	testAccPreCheck(t)

	// S3 API configuration
	if v := os.Getenv("GARAGE_S3_ENDPOINT"); v == "" {
		t.Fatal("GARAGE_S3_ENDPOINT must be set for object acceptance tests")
	}

	if v := os.Getenv("GARAGE_ACCESS_KEY"); v == "" {
		t.Fatal("GARAGE_ACCESS_KEY must be set for object acceptance tests")
	}

	if v := os.Getenv("GARAGE_SECRET_KEY"); v == "" {
		t.Fatal("GARAGE_SECRET_KEY must be set for object acceptance tests")
	}
}

// testAccProviderConfig returns the base provider configuration for tests.
func testAccProviderConfig() string {
	return `
provider "garage" {
  endpoints {
    admin = "` + os.Getenv("GARAGE_ADMIN_ENDPOINT") + `"
    s3    = "` + os.Getenv("GARAGE_S3_ENDPOINT") + `"
  }
  token      = "` + os.Getenv("GARAGE_TOKEN") + `"
  access_key = "` + os.Getenv("GARAGE_ACCESS_KEY") + `"
  secret_key = "` + os.Getenv("GARAGE_SECRET_KEY") + `"
}
`
}

func TestProviderSchema(t *testing.T) {
	p := New("test")()
	schemaReq := provider.SchemaRequest{}
	schemaResp := provider.SchemaResponse{}

	p.Schema(context.Background(), schemaReq, &schemaResp)

	t.Logf("Schema Attributes: %+v", schemaResp.Schema.Attributes)
	t.Logf("Schema Blocks: %+v", schemaResp.Schema.Blocks)

	if schemaResp.Schema.Blocks == nil {
		t.Fatal("Blocks is nil - endpoints block missing!")
	}

	if _, ok := schemaResp.Schema.Blocks["endpoints"]; !ok {
		t.Fatal("endpoints block not found in schema!")
	}

	t.Log("SUCCESS: endpoints block exists!")
}
