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
			createSyntheticsTestStep,
		},
	})
}

func TestAccDatadogSyntheticsTest_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed,
		Steps: []resource.TestStep{
			createSyntheticsTestStep,
			updateSyntheticsTestStep,
		},
	})
}

var createSyntheticsTestStep = resource.TestStep{
	Config: createSyntheticsTestConfig,
	Check: resource.ComposeTestCheckFunc(
		testSyntheticsTestExists(),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "type", "api"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.method", "GET"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.url", "https://www.datadoghq.com"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.#", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.type", "statusCode"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.operator", "is"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.target", "200"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.1.type", "responseTime"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.1.operator", "lessThan"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.1.target", "2000"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.#", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.0", "aws:eu-central-1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.1", "aws:ap-northeast-1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "options.tick_every", "60"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "name", "name for synthetics test foo"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "message", "Notify @datadog.user"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.#", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.0", "foo:bar"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.1", "baz"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "paused", "true"),
	),
}

const createSyntheticsTestConfig = `
resource "datadog_synthetics_test" "foo" {
  type = "api"

  request {
	  method = "GET"
	  url = "https://www.datadoghq.com"
  }
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

  locations = [ "aws:eu-central-1", "aws:ap-northeast-1" ]
  options {
	tick_every = 60
  }

  name = "name for synthetics test foo"
  message = "Notify @datadog.user"
  tags = ["foo:bar", "baz"]

  paused = true
}
`

var updateSyntheticsTestStep = resource.TestStep{
	Config: updateSyntheticsTestConfig,
	Check: resource.ComposeTestCheckFunc(
		testSyntheticsTestExists(),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "type", "api"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.method", "GET"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.url", "https://docs.datadoghq.com"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.#", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.type", "statusCode"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.operator", "isNot"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.target", "500"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.#", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.0", "aws:eu-central-1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "options.tick_every", "900"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "name", "updated name"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "message", "Notify @pagerduty"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.#", "3"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.0", "foo:bar"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.1", "foo"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.2", "env:test"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "paused", "false"),
	),
}

const updateSyntheticsTestConfig = `
resource "datadog_synthetics_test" "foo" {
  type = "api"

  request {
	  method = "GET"
	  url = "https://docs.datadoghq.com"
  }

  assertions = [
    {
      type = "statusCode"
      operator = "isNot"
      target = "500"
  	}
  ]

  locations = [ "aws:eu-central-1" ]

  options {
	tick_every = 900
  }

  name = "updated name"
  message = "Notify @pagerduty"
  tags = ["foo:bar", "foo", "env:test"]

  paused = false
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
