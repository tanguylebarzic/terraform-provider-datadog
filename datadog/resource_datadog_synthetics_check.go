/**
For more info about writing custom provider: shttps://www.terraform.io/docs/extend/writing-custom-providers.html
**/
package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	datadog "github.com/zorkian/go-datadog-api"
	"strconv"
)

var syntheticsTypes = []string{"api", "browser"}

func resourceDatadogSyntheticsTest() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogSyntheticsTestCreate,
		Read:   resourceDatadogSyntheticsTestRead,
		Update: resourceDatadogSyntheticsTestUpdate,
		Delete: resourceDatadogSyntheticsTestDelete,
		Exists: resourceDatadogSyntheticsTestExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogSyntheticsTestImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(syntheticsTypes, false),
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"request": syntheticsTestRequest(),
			"assertions": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"locations": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					// TODO: validation with regexp
				},
			},
			"options": syntheticsTestOptions(),
			"set_live": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func syntheticsTestRequest() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"method": {
					Type:     schema.TypeString,
					Required: true,
				},
				"url": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func syntheticsTestOptions() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_every": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		},
	}
}

func buildSyntheticsTestStruct(d *schema.ResourceData) *datadog.SyntheticsTest {
	request := datadog.SyntheticsRequest{}
	if attr, ok := d.GetOk("request.method"); ok {
		request.SetMethod(attr.(string))
	}
	if attr, ok := d.GetOk("request.url"); ok {
		request.SetUrl(attr.(string))
	}

	assertions := []datadog.SyntheticsAssertion{}
	if attr, ok := d.GetOk("assertions"); ok {
		for _, attr := range attr.([]interface{}) {
			assertion := datadog.SyntheticsAssertion{}
			assertionMap := attr.(map[string]interface{})
			if v, ok := assertionMap["type"]; ok {
				assertionType := v.(string)
				assertion.Type = &assertionType
			}
			if v, ok := assertionMap["operator"]; ok {
				assertionOperator := v.(string)
				assertion.Operator = &assertionOperator
			}
			if v, ok := assertionMap["target"]; ok {
				assertionTarget := v.(string)
				t, _ := strconv.Atoi(assertionTarget)
				assertion.Target = t
			}
			assertions = append(assertions, assertion)
		}
	}

	config := datadog.SyntheticsConfig{
		Request:    &request,
		Assertions: assertions,
		Variables:  []interface{}{}, // TODO: what is it?
	}

	options := datadog.SyntheticsOptions{}
	if attr, ok := d.GetOk("options.tick_every"); ok {
		tickEvery, _ := strconv.Atoi(attr.(string))
		options.SetTickEvery(tickEvery)
	}

	syntheticsTest := datadog.SyntheticsTest{
		Name:    datadog.String(d.Get("name").(string)),
		Type:    datadog.String(d.Get("type").(string)),
		Config:  &config,
		Options: &options,
		Message: datadog.String(d.Get("message").(string)),
	}

	if attr, ok := d.GetOk("locations"); ok {
		locations := []string{}
		for _, s := range attr.([]interface{}) {
			locations = append(locations, s.(string))
		}
		syntheticsTest.Locations = locations
	}

	if attr, ok := d.GetOk("tags"); ok {
		tags := []string{}
		for _, s := range attr.([]interface{}) {
			tags = append(tags, s.(string))
		}
		syntheticsTest.Tags = tags
	}

	return &syntheticsTest
}

func resourceDatadogSyntheticsTestCreate(d *schema.ResourceData, meta interface{}) error {
	println("Creating")

	client := meta.(*datadog.Client)

	syntheticsTest := buildSyntheticsTestStruct(d)
	createdSyntheticsTest, err := client.CreateSyntheticsCheck(syntheticsTest)
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return fmt.Errorf("error creating synthetics test: %s", err.Error())
	}

	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsTest.GetPublicId())

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestRead(d *schema.ResourceData, meta interface{}) error {
	println("Reading")
	return nil
}

func resourceDatadogSyntheticsTestUpdate(d *schema.ResourceData, meta interface{}) error {
	println("Updating")

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestDelete(d *schema.ResourceData, meta interface{}) error {
	println("Deleting")
	client := meta.(*datadog.Client)

	if err := client.DeleteSyntheticsChecks([]string{d.Id()}); err != nil {
		return err
	}

	return nil
}

// resourceDatadogSyntheticsTestExists is called to verify a resource still exists.
// It is called prior to Read, and lowers the burden of Read to be able to assume the resource exists.
func resourceDatadogSyntheticsTestExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	println("Exists?")
	client := meta.(*datadog.Client)

	if _, err := client.GetSyntheticsCheck(d.Id()); err != nil {
		return false, err
	}

	return true, nil
}

func resourceDatadogSyntheticsTestImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	println("Synthetics test import")
	if err := resourceDatadogSyntheticsTestRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
