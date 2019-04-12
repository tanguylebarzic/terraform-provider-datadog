// For more info about writing custom provider: shttps://www.terraform.io/docs/extend/writing-custom-providers.html

package datadog

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	datadog "github.com/zorkian/go-datadog-api"
)

var syntheticsTypes = []string{"api", "browser"}

func resourceDatadogSyntheticsTest() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogSyntheticsTestCreate,
		Read:   resourceDatadogSyntheticsTestRead,
		Update: resourceDatadogSyntheticsTestUpdate,
		Delete: resourceDatadogSyntheticsTestDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(syntheticsTypes, false),
			},
			"request": syntheticsTestRequest(),
			"request_headers": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"assertions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"devices": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"locations": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
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
				"body": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"timeout": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  0,
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
				"min_failure_duration": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"min_location_failed": {
					Type:     schema.TypeInt,
					Optional: true,
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

func isTargetOfTypeInt(assertionType string) bool {
	for _, intTargetAssertionType := range []string{"responseTime", "statusCode"} {
		if assertionType == intTargetAssertionType {
			return true
		}
	}
	return false
}

func newSyntheticsTestFromLocalState(d *schema.ResourceData) *datadog.SyntheticsTest {
	request := datadog.SyntheticsRequest{}
	if attr, ok := d.GetOk("request.method"); ok {
		request.SetMethod(attr.(string))
	}
	if attr, ok := d.GetOk("request.url"); ok {
		request.SetUrl(attr.(string))
	}
	if attr, ok := d.GetOk("request.body"); ok {
		request.SetBody(attr.(string))
	}
	if attr, ok := d.GetOk("request.timeout"); ok {
		timeoutInt, _ := strconv.Atoi(attr.(string))
		request.SetTimeout(timeoutInt)
	}
	if attr, ok := d.GetOk("request_headers"); ok {
		headers := attr.(map[string]interface{})
		if len(headers) > 0 {
			request.Headers = make(map[string]string)
		}
		for k, v := range headers {
			request.Headers[k] = v.(string)
		}
	}

	config := datadog.SyntheticsConfig{
		Request:   &request,
		Variables: []interface{}{},
	}

	if attr, ok := d.GetOk("assertions"); ok {
		for _, attr := range attr.([]interface{}) {
			assertion := datadog.SyntheticsAssertion{}
			assertionMap := attr.(map[string]interface{})
			if v, ok := assertionMap["type"]; ok {
				assertionType := v.(string)
				assertion.Type = &assertionType
			}
			if v, ok := assertionMap["property"]; ok {
				assertionProperty := v.(string)
				assertion.Property = &assertionProperty
			}
			if v, ok := assertionMap["operator"]; ok {
				assertionOperator := v.(string)
				assertion.Operator = &assertionOperator
			}
			if v, ok := assertionMap["target"]; ok {
				if isTargetOfTypeInt(*assertion.Type) {
					assertionTargetInt, _ := strconv.Atoi(v.(string))
					assertion.Target = assertionTargetInt
				} else if *assertion.Operator == "validates" {
					assertion.Target = json.RawMessage(v.(string))
				} else {
					assertion.Target = v.(string)
				}
			}
			config.Assertions = append(config.Assertions, assertion)
		}
	}

	options := datadog.SyntheticsOptions{}
	if attr, ok := d.GetOk("options.tick_every"); ok {
		tickEvery, _ := strconv.Atoi(attr.(string))
		options.SetTickEvery(tickEvery)
	}
	if attr, ok := d.GetOk("options.min_failure_duration"); ok {
		minFailureDuration, _ := strconv.Atoi(attr.(string))
		options.SetMinFailureDuration(minFailureDuration)
	}
	if attr, ok := d.GetOk("options.min_location_failed"); ok {
		minLocationFailed, _ := strconv.Atoi(attr.(string))
		options.SetMinFailureDuration(minLocationFailed)
	}
	if attr, ok := d.GetOk("devices"); ok {
		for _, attr := range attr.([]interface{}) {
			device := datadog.SyntheticsDevice{}
			deviceMap := attr.(map[string]interface{})
			if v, ok := deviceMap["id"]; ok {
				device.SetId(v.(string))
			}
			if v, ok := deviceMap["name"]; ok {
				device.SetName(v.(string))
			}
			if v, ok := deviceMap["height"]; ok {
				deviceHeight := v.(string)
				deviceHeightInt, _ := strconv.Atoi(deviceHeight)
				device.SetHeight(deviceHeightInt)
			}
			if v, ok := deviceMap["width"]; ok {
				deviceWidth := v.(string)
				deviceWidthInt, _ := strconv.Atoi(deviceWidth)
				device.SetWidth(deviceWidthInt)
			}
			if v, ok := deviceMap["isMobile"]; ok {
				device.SetIsMobile(v.(string) == "true")
			}
			if v, ok := deviceMap["isLandscape"]; ok {
				device.SetIsLandscape(v.(string) == "true")
			}
			options.Devices = append(options.Devices, device)
		}
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
	d.Set("type", syntheticsTest.GetType())

	actualRequest := syntheticsTest.GetConfig().Request
	localRequest := newLocalMap(map[string]interface{}{
		"method":  actualRequest.GetMethod(),
		"url":     actualRequest.GetUrl(),
		"body":    actualRequest.GetBody(),
		"timeout": actualRequest.GetTimeout(),
	})
	d.Set("request", localRequest)
	d.Set("request_headers", actualRequest.Headers)

	actualAssertions := syntheticsTest.GetConfig().Assertions
	localAssertions := []map[string]string{}
	for _, assertion := range actualAssertions {
		localAssertion := newLocalMap(map[string]interface{}{
			"type":     assertion.GetType(),
			"property": assertion.GetProperty(),
			"operator": assertion.GetOperator(),
			"target":   assertion.Target,
		})
		localAssertions = append(localAssertions, localAssertion)
	}
	d.Set("assertions", localAssertions)

	actualDevices := syntheticsTest.GetOptions().Devices
	localDevices := []map[string]string{}
	for _, device := range actualDevices {
		localDevice := newLocalMap(map[string]interface{}{
			"id":       device.GetId(),
			"name":     device.GetName(),
			"height":   device.GetHeight(),
			"width":    device.GetWidth(),
			"isMobile": device.GetIsMobile(),
		})
		localDevices = append(localDevices, localDevice)
	}
	d.Set("devices", localDevices)

	d.Set("locations", syntheticsTest.Locations)

	actualOptions := syntheticsTest.GetOptions()
	localOptions := newLocalMap(map[string]interface{}{
		"tick_every":           actualOptions.GetTickEvery(),
		"min_failure_duration": actualOptions.GetMinFailureDuration(),
		"min_location_failed":  actualOptions.GetMinLocationFailed(),
	})
	d.Set("options", localOptions)

	d.Set("name", syntheticsTest.GetName())
	d.Set("message", syntheticsTest.GetMessage())
	d.Set("tags", syntheticsTest.Tags)
	d.Set("paused", *syntheticsTest.Status == "paused")
}

func updateSyntheticsTestLiveness(d *schema.ResourceData, client *datadog.Client) {
	paused, ok := d.GetOk("paused")
	if !ok {
		return
	}
	if paused.(bool) {
		client.PauseSyntheticsTest(d.Id())
	} else {
		client.ResumeSyntheticsTest(d.Id())
	}
}

func newLocalMap(actualMap map[string]interface{}) map[string]string {
	localMap := make(map[string]string)
	for k, i := range actualMap {
		var valStr string
		switch v := i.(type) {
		case bool:
			if v {
				valStr = "1"
			} else {
				valStr = "0"
			}
		case int:
			valStr = strconv.Itoa(v)
		case float64:
			valStr = strconv.Itoa(int(v))
		case string:
			valStr = v
		default:
			// Ignore value
			// TODO: manage target for JSON body assertions
		}
		if valStr != "" {
			localMap[k] = valStr
		}
	}
	return localMap
}
