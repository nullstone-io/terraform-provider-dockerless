package provider

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRemoteImageResource(t *testing.T) {
	// NOTE: This test requires ACC_DOCKER_USERNAME, ACC_DOCKER_PASSWORD env vars set with access to push to nullstone/tf-provider-test

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheckDockerHub(t) },
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

// TestAccRemoteImageResource_localRegistry exercises the full resource lifecycle against an
// in-process OCI registry, so it runs anywhere without external services or credentials.
func TestAccRemoteImageResource_localRegistry(t *testing.T) {
	srv := httptest.NewServer(registry.New())
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	// go-containerregistry only speaks http to a registry named localhost:<port>
	host := "localhost:" + u.Port()

	// Seed a source image in the local registry
	source := host + "/acc/source:v1"
	img, err := random.Image(1024, 3)
	if err != nil {
		t.Fatal(err)
	}
	srcRef, err := name.ParseReference(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Write(srcRef, img); err != nil {
		t.Fatal(err)
	}

	targetV1 := host + "/acc/target:v1"
	targetV2 := host + "/acc/target:v2"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRemoteImageResourceLocalConfig(source, targetV1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerless_remote_image.test", "source", source),
					resource.TestCheckResourceAttr("dockerless_remote_image.test", "target", targetV1),
					resource.TestCheckResourceAttrSet("dockerless_remote_image.test", "digest"),
				),
			},
			// Update and Read testing
			{
				Config: testAccRemoteImageResourceLocalConfig(source, targetV2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dockerless_remote_image.test", "target", targetV2),
					resource.TestCheckResourceAttrSet("dockerless_remote_image.test", "digest"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccRemoteImageResourceLocalConfig(source, target string) string {
	return fmt.Sprintf(`
resource "dockerless_remote_image" "test" {
  source = %[1]q
  target = %[2]q
}
`, source, target)
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
