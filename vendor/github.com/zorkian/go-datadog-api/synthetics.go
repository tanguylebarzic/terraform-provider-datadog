package datadog

import (
	"fmt"
)

// SyntheticsTest represents a synthetics test, either api or browser
type SyntheticsTest struct {
	PublicId    *string            `json:"public_id,omitempty"`
	Name        *string            `json:"name,omitempty"`
	Type        *string            `json:"type,omitempty"`
	Tags        []string           `json:"tags"`
	CreatedAt   *string            `json:"created_at,omitempty"`
	ModifiedAt  *string            `json:"modified_at,omitempty"`
	DeletedAt   *string            `json:"deleted_at,omitempty"`
	Config      *SyntheticsConfig  `json:"config,omitempty"`
	Message     *string            `json:"message,omitempty"`
	Options     *SyntheticsOptions `json:"options,omitempty"`
	Locations   []string           `json:"locations,omitempty"`
	CreatedBy   *SyntheticsUser    `json:"created_by,omitempty"`
	ModifiedBy  *SyntheticsUser    `json:"modified_by,omitempty"`
	CheckStatus *string            `json:"check_status,omitempty"`
}

type SyntheticsConfig struct {
	Request    *SyntheticsRequest    `json:"request"`
	Assertions []SyntheticsAssertion `json:"assertions"`
	Variables  []interface{}         `json:"variables,omitempty"`
}

type SyntheticsRequest struct {
	Url     *string           `json:"url"`
	Method  *string           `json:"method"`
	Timeout *int              `json:"timeout,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    *string           `json:"body,omitempty"`
}

type SyntheticsAssertion struct {
	Operator *string `json:"operator,omitempty"`
	Property *string `json:"property,omitempty"`
	Type     *string `json:"type,omitempty"`
	// sometimes target is string ( like "text/html; charset=UTF-8" for header content-type )
	// and sometimes target is int ( like 1200 for responseTime, 200 for statusCode )
	// TODO: making it an interface{} doesn't work; it false back to string when serialized to JSON
	Target int `json:"target,omitempty"`
}

type SyntheticsOptions struct {
	TickEvery *int `json:"tick_every,omitempty"`
}

type SyntheticsUser struct {
	Id     *int    `json:"id,omitempty"`
	Name   *string `json:"name,omitempty"`
	Email  *string `json:"email,omitempty"`
	Handle *string `json:"handle,omitempty"`
}

type SyntheticsChecksList struct {
	Checks []SyntheticsTest `json:"checks,omitempty"`
}

type ToggleStatus struct {
	NewStatus *string `json:"new_status"`
}

// GetSyntheticsTests get all tests of type API
func (client *Client) GetSyntheticsTests() ([]SyntheticsTest, error) {
	var out SyntheticsChecksList
	if err := client.doJsonRequest("GET", "/v0/synthetics/checks?type=api", nil, &out); err != nil {
		return nil, err
	}
	return out.Checks, nil
}

// GetSyntheticsCheck get check by public id
func (client *Client) GetSyntheticsCheck(publicId string) (*SyntheticsTest, error) {
	var out SyntheticsTest
	if err := client.doJsonRequest("GET", "/v0/synthetics/checks/"+publicId, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateSyntheticsCheck creates a check
func (client *Client) CreateSyntheticsCheck(check *SyntheticsTest) (*SyntheticsTest, error) {
	var out SyntheticsTest
	if err := client.doJsonRequest("POST", "/v0/synthetics/checks", check, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSyntheticsCheck updates a check
func (client *Client) UpdateSyntheticsCheck(publicId string, check *SyntheticsTest) (*SyntheticsTest, error) {
	var out SyntheticsTest
	if err := client.doJsonRequest("PUT", fmt.Sprintf("/v0/synthetics/checks/%s", publicId), check, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PauseSyntheticsTest set a test status to live
func (client *Client) PauseSyntheticsTest(publicId string) (*bool, error) {
	payload := ToggleStatus{NewStatus: String("paused")}
	out := Bool(false)
	if err := client.doJsonRequest("PUT", fmt.Sprintf("/v0/synthetics/checks/%s/status", publicId), &payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ResumeSyntheticsTest set a test status to live
func (client *Client) ResumeSyntheticsTest(publicId string) (*bool, error) {
	payload := ToggleStatus{NewStatus: String("live")}
	out := Bool(false)
	if err := client.doJsonRequest("PUT", fmt.Sprintf("/v0/synthetics/checks/%s/status", publicId), &payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// string array of public_id
type DeleteSyntheticsChecksRequest struct {
	CheckIds []string `json:"check_ids,omitempty"`
}

// DeleteSyntheticsChecks deletes checks
func (client *Client) DeleteSyntheticsChecks(publicIds []string) error {
	req := DeleteSyntheticsChecksRequest{
		CheckIds: publicIds,
	}
	if err := client.doJsonRequest("POST", "/v0/synthetics/checks/delete", req, nil); err != nil {
		return err
	}
	return nil
}
