package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zorkian/go-datadog-api"
)

func TestAccDatadogSyntheticsTest_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testSyntheticsTestConfig,
				Check: resource.ComposeTestCheckFunc(
					testSyntheticsTestExists(),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_test.foo", "name", "name for synthetics test foo"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_test.foo", "options.tick_every", "60"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_test.foo", "setLive", "false"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_test.foo", "tags.0", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_test.foo", "tags.1", "baz"),
				),
			},
		},
	})
}

const testSyntheticsTestConfig = `
resource "datadog_synthetics_test" "foo" {
  name = "name for synthetics test foo"
  type = "api"

  request {
	  method = "GET"
	  url = "https://datadoghq.com"
  }

  locations = [ "aws:eu-central-1", "aws:ap-northeast-1" ]

  assertions = [
    {
      type = "statusCode"
      operator = "is"
      target = "200"
  	},
    {
      type = "responseTime"
      operator = "lessThan"
      target = "2000"
  	}
  ]

  options {
	tick_every = 60
  }

  message = "Notify @datadog.user"
  set_live = false
  tags = ["foo:bar", "baz"]
}
`

func testSyntheticsTestExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)

		for _, r := range s.RootModule().Resources {
			if _, err := client.GetSyntheticsCheck(r.Primary.ID); err != nil {
				return fmt.Errorf("Received an error retrieving synthetics test %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsTestIsDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	for _, r := range s.RootModule().Resources {
		if _, err := client.GetSyntheticsCheck(r.Primary.ID); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("Received an error retrieving synthetics test %s", err)
		}
		return fmt.Errorf("Synthetics test still exists")
	}
	return nil
}
