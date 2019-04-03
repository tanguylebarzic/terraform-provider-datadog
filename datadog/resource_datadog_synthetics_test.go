package datadog

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
			"options": syntheticsTestOptions(),
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"setLive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
	return nil
}

func resourceDatadogSyntheticsTestRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogSyntheticsTestUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogSyntheticsTestDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogSyntheticsTestExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	return b, nil
}

func resourceDatadogSyntheticsTestImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogSyntheticsTestRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
