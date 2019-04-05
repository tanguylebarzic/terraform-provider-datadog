// For more info about writing custom provider: shttps://www.terraform.io/docs/extend/writing-custom-providers.html

package datadog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	datadog "github.com/zorkian/go-datadog-api"
)

// TODO: Add support for browser tests
var syntheticsTypes = []string{"api"}

func resourceDatadogSyntheticsTest() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogSyntheticsTestCreate,
		Read:   resourceDatadogSyntheticsTestRead,
		Update: resourceDatadogSyntheticsTestUpdate,
		Delete: resourceDatadogSyntheticsTestDelete,
		Exists: resourceDatadogSyntheticsTestExists,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(syntheticsTypes, false),
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"paused": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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

func resourceDatadogSyntheticsTestCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	syntheticsTest := newSyntheticsTestFromLocalState(d)
	createdSyntheticsTest, err := client.CreateSyntheticsTest(syntheticsTest)
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return fmt.Errorf("error creating synthetics test: %s", err.Error())
	}

	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsTest.GetPublicId())

	// Call resume/pause webservice, because it is a dedicated endpoint apart from classical CRUD operations
	updateSyntheticsTestLiveness(d, client)

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	syntheticsTest, err := client.GetSyntheticsTest(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
		}
		return err
	}

	updateSyntheticsTestLocalState(d, syntheticsTest)

	return nil
}

func resourceDatadogSyntheticsTestUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO: should we implement partial mode?
	client := meta.(*datadog.Client)

	syntheticsTest := newSyntheticsTestFromLocalState(d)
	if _, err := client.UpdateSyntheticsTest(d.Id(), syntheticsTest); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		return err
	}

	// Call resume/pause webservice, because it is a dedicated endpoint apart from classical CRUD operations
	updateSyntheticsTestLiveness(d, client)

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	if err := client.DeleteSyntheticsTests([]string{d.Id()}); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return err
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

// resourceDatadogSyntheticsTestExists is called to verify a resource still exists.
// It is called prior to Read, and lowers the burden of Read to be able to assume the resource exists.
func resourceDatadogSyntheticsTestExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*datadog.Client)

	if _, err := client.GetSyntheticsTest(d.Id()); err != nil {
		return false, err
	}

	return true, nil
}

func newSyntheticsTestFromLocalState(d *schema.ResourceData) *datadog.SyntheticsTest {
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

func updateSyntheticsTestLocalState(d *schema.ResourceData, syntheticsTest *datadog.SyntheticsTest) {
	// Note: there is no need to update `paused` since the source of truth is actually the value set in the terraform config file.
	d.Set("type", syntheticsTest.GetType())
	d.Set("request", syntheticsTest.GetConfig().Request)
	d.Set("assertions", syntheticsTest.GetConfig().Assertions)
	d.Set("locations", syntheticsTest.Locations)
	d.Set("options", syntheticsTest.GetOptions())
	d.Set("name", syntheticsTest.GetName())
	d.Set("message", syntheticsTest.GetMessage())
	d.Set("tags", syntheticsTest.Tags)
}

func updateSyntheticsTestLiveness(d *schema.ResourceData, client *datadog.Client) {
	if d.Get("paused").(bool) {
		client.ResumeSyntheticsTest(d.Id())
	} else {
		client.PauseSyntheticsTest(d.Id())
	}
}
