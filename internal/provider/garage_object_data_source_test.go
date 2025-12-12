package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGarageObjectDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckS3(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGarageObjectDataSourceConfig("Hello from data source test!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.garage_object.test", "key", "test-data-object.txt"),
					resource.TestCheckResourceAttr("data.garage_object.test", "body", "Hello from data source test!"),
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

func testAccGarageObjectDataSourceConfig(content string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
       resource "garage_bucket" "test" {
          global_alias = "test-bucket-object-ds"
       }
       
       resource "garage_bucket_permission" "test" {
          bucket_id     = garage_bucket.test.id
          access_key_id = %[2]q
          
          read  = true
          write = true
          owner = false
       }
       
       resource "garage_object" "test" {
          depends_on = [garage_bucket_permission.test]
          
          bucket       = garage_bucket.test.id
          key          = "test-data-object.txt"
          content      = %[1]q
          content_type = "text/plain"
       }
       
       data "garage_object" "test" {
          bucket = garage_bucket.test.id
          key    = garage_object.test.key
       }
       `, content, os.Getenv("GARAGE_ACCESS_KEY"),
	)
}
