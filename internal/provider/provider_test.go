package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dockerless": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

// testAccPreCheckDockerHub gates tests that push to the real Docker Hub repository
// nullstone/tf-provider-test, which requires credentials with push access.
func testAccPreCheckDockerHub(t *testing.T) {
	if os.Getenv("ACC_DOCKER_USERNAME") == "" || os.Getenv("ACC_DOCKER_PASSWORD") == "" {
		t.Skip("set ACC_DOCKER_USERNAME and ACC_DOCKER_PASSWORD with push access to nullstone/tf-provider-test to run this test")
	}
}
