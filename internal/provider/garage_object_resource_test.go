package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGarageObjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckS3(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGarageObjectResourceConfig("test-content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("garage_object.test", "key", "test-object.txt"),
					resource.TestCheckResourceAttr("garage_object.test", "content", "test-content"),
					resource.TestCheckResourceAttr("garage_object.test", "content_type", "text/plain"),
					resource.TestCheckResourceAttrSet("garage_object.test", "etag"),
					resource.TestCheckResourceAttrSet("garage_object.test", "id"),
				),
			},
			{
				ResourceName:            "garage_object.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"content", "source"},
			},
			{
				Config: testAccGarageObjectResourceConfig("updated-content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("garage_object.test", "content", "updated-content"),
				),
			},
		},
	})
}

func testAccGarageObjectResourceConfig(content string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
		resource "garage_bucket" "test" {
			global_alias = "test-bucket-object"
		}
		
		resource "garage_bucket_permission" "test" {
			bucket_id = garage_bucket.test.id
			access_key_id = %[2]q
			
			read  = true
			write = true
			owner = false
		}
		
		resource "garage_object" "test" {
			depends_on = [garage_bucket_permission.test]

			bucket       = garage_bucket.test.id
			key          = "test-object.txt"
			content      = %[1]q
			content_type = "text/plain"
			
			
		}
		`, content, os.Getenv("GARAGE_ACCESS_KEY"),
	)
}
