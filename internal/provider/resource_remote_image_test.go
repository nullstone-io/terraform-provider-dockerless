package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRemoteImageResource(t *testing.T) {
	// NOTE: This test requires ACC_DOCKER_USERNAME, ACC_DOCKER_PASSWORD env vars set with access to push to nullstone/tf-provider-test

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRemoteImageResourceConfig("nullstone/tf-provider-test:v1", "nullstone/tf-provider-test:v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerless_remote_image.test", "source", "nullstone/tf-provider-test:v1"),
					resource.TestCheckResourceAttr("dockerless_remote_image.test", "target", "nullstone/tf-provider-test:v2"),
					resource.TestCheckResourceAttrSet("dockerless_remote_image.test", "digest"),
				),
			},
			// ImportState testing
			/*{
				ResourceName:            "dockerless_remote_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source"},
			},*/
			// Update and Read testing
			{
				Config: testAccRemoteImageResourceConfig("nullstone/tf-provider-test:v1", "nullstone/tf-provider-test:v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dockerless_remote_image.test", "digest"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccRemoteImageResourceConfig(source, target string) string {
	username := os.Getenv("ACC_DOCKER_USERNAME")
	password := os.Getenv("ACC_DOCKER_PASSWORD")

	return fmt.Sprintf(`
provider "dockerless" {
	registry_auth = {
      "index.docker.io" = {
        username = %[3]q
        password = %[4]q
      } 
	}
}

resource "dockerless_remote_image" "test" {
  source = %[1]q
  target = %[2]q
}
`, source, target, username, password)
}
