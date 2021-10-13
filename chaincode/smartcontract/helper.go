package smartcontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EnforcePenaltyRulesRequest represents for enforcing penalty rules request.
type EnforcePenaltyRulesRequest struct {
	EvaluationID  string         `json:"evaluationId"`
	ReservationID string         `json:"reservationId"`
	AgreementID   string         `json:"agreementId"`
	PenaltyRules  []*PenaltyRule `json:"penaltyRules"`
	Reason        string         `json:"reason"`
}

// EnforcePenaltyRules enforces penalty rules.
func EnforcePenaltyRules(req *EnforcePenaltyRulesRequest, token string) error {
	bodyReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url := BaseURL + "/rpc/agreements/enforce-penalty-rules"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyReq))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	timeout := 60 * time.Second
	client := &http.Client{Timeout: timeout}
	httpRes, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprintf("EnforcePenaltyRules with response :%d", httpRes.StatusCode))
	}

	return nil
}

// StringInSlice determines a string is in
// a slice of strings or not.
func StringInSlice(str string, a []string) bool {
	for _, s := range a {
		if s == str {
			return true
		}
	}
	return false
}

// ParseTime parses string to time.
func ParseTime(s string) (time.Time, error) {
	return time.Parse(RFC3339, s)
}

// MakeErrorAgreementItemCodeDoesNotSupport makes error that a agreement item code has not supported yet.
func MakeErrorAgreementItemCodeDoesNotSupport(code, cat string) error {
	return fmt.Errorf("agreement item code %s in category %s as not supported yet", code, cat)
}
