package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGarageObjectDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckS3(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGarageObjectDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.garage_object.test", "key", "test-data-object.txt"),
					resource.TestCheckResourceAttr("data.garage_object.test", "body", "test data source content"),
					resource.TestCheckResourceAttr("data.garage_object.test", "content_type", "text/plain"),
					resource.TestCheckResourceAttrSet("data.garage_object.test", "etag"),
					resource.TestCheckResourceAttrSet("data.garage_object.test", "content_length"),
					resource.TestCheckResourceAttrSet("data.garage_object.test", "last_modified"),
					resource.TestCheckResourceAttrSet("data.garage_object.test", "id"),
				),
			},
		},
	})
}

func testAccGarageObjectDataSourceConfig() string {
	return testAccProviderConfig() + `
resource "garage_bucket" "test" {
  global_alias = "test-bucket-datasource"
}

resource "garage_object" "test" {
  bucket       = garage_bucket.test.id
  key          = "test-data-object.txt"
  content      = "test data source content"
  content_type = "text/plain"
}

data "garage_object" "test" {
  bucket = garage_bucket.test.id
  key    = garage_object.test.key
}
`
}
