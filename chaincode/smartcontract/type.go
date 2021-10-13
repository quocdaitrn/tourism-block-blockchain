package smartcontract

import "time"

const (
	// RFC3339 format.
	RFC3339 = "2006-01-02T15:04:05.000Z"
)

// Service stores information of a service.
type Service struct {
	DocType             string       `json:"docType"` // docType is used to distinguish the various types of objects in state database.
	ServiceID           string       `json:"serviceId"`
	RuleAbidingRate     float32      `json:"ruleAbidingRate"`
	SatisfactionRate    float32      `json:"satisfactionRate"`
	NumberOfEvaluations uint64       `json:"numberOfEvaluations"`
	LastEvaluationAt    string       `json:"lastEvaluationAt"`
	Agreements          []*Agreement `json:"agreements"`
}

// Agreement stores information of a agreement of a service.
type Agreement struct {
	AgreementID string           `json:"agreementId"`
	Category    string           `json:"category"`
	Items       []*AgreementItem `json:"items"`

	TotalFeedbacks                         uint `json:"totalFeedbacks"`
	TotalUnsatisfied                       uint `json:"totalUnsatisfied"`
	TotalRuleViolations                    uint `json:"totalRuleViolations"`
	TotalRuleViolationWithoutCompensations uint `json:"totalRuleViolationWithoutCompensations"`

	HasPenaltyRule bool           `json:"hasPenaltyRule"`
	PenaltyRules   []*PenaltyRule `json:"penaltyRules"`

	LastEvaluationAt string `json:"lastEvaluationAt"`

	RuleAbidingRate  float32 `json:"ruleAbidingRate"`
	SatisfactionRate float32 `json:"satisfactionRate"`
}

// AgreementItem an item in a agreement.
type AgreementItem struct {
	Code string `json:"code"`

	// facilities
	Quantity int `json:"quantity,omitempty" metadata:"quantity,optional"`

	// airport shuttle
	DriverMaxWaitTime     int `json:"driverMaxWaitTime,omitempty" metadata:"driverMaxWaitTime,optional"`
	CustomerShortWaitTime int `json:"customerShortWaitTime,omitempty" metadata:"customerShortWaitTime,optional"`
	CustomerLongWaitTime  int `json:"customerLongWaitTime,omitempty" metadata:"customerLongWaitTime,optional"`

	// sauna
	MaxFailures             int `json:"maxFailures,omitempty" metadata:"maxFailures,optional"`
	MinTimeBetween2Failures int `json:"minTimeBetween2Failures,omitempty" metadata:"minTimeBetween2Failures,optional"`

	// room design
	Value interface{} `json:"value,omitempty" metadata:"value,optional"`
}

// PenaltyRule types of penalty rule.
type PenaltyRule struct {
	Type            string  `json:"type"`
	DiscountPercent float32 `json:"discountPercent,omitempty" metadata:"discountPercent,optional"`
}

// EvaluationData data for verifying agreement.
type EvaluationData struct {
	Code string `json:"code"`
	Name string `json:"name,omitempty" metadata:"name,optional"`

	// facility attributes
	Quantity int `json:"quantity,omitempty" metadata:"quantity,optional"`

	// airport shuttle service attributes
	PickUpTime                         time.Time `json:"pickUpTime,omitempty" metadata:"pickUpTime,optional"`
	DriverArriveAt                     time.Time `json:"driverArriveAt,omitempty" metadata:"driverArriveAt,optional"`
	CustomerCheckInAt                  time.Time `json:"customerCheckInAt,omitempty" metadata:"customerCheckInAt,optional"`
	LastUpdatedArrivalTimeByCustomerAt time.Time `json:"lastUpdatedArrivalTimeByCustomerAt,omitempty" metadata:"lastUpdatedArrivalTimeByCustomerAt,optional"`
	DriverNotifyCustomerDoNotShowUpAt  time.Time `json:"driverNotifyCustomerDoNotShowUpAt,omitempty" metadata:"driverNotifyCustomerDoNotShowUpAt,optional"`
	Status                             string    `json:"status,omitempty" metadata:"status,optional"`

	// sauna attributes
	SaunaRequests []*SaunaRequest `json:"saunaRequests,omitempty" metadata:"saunaRequests,optional"`

	// service attributes
	Fulfilled string `json:"fulfilled,omitempty" metadata:"fulfilled,optional"`

	// room design and other attributes
	Value interface{} `json:"value,omitempty" metadata:"value,optional"`
}

// SaunaRequest represents sauna request.
type SaunaRequest struct {
	RequestAt time.Time `json:"requestAt"`
	Status    string    `json:"status"`
}

// Evaluation stores information for auditing.
type Evaluation struct {
	DocType      string `json:"docType"` // docType is used to distinguish the various types of objects in state database.
	EvaluationID string `json:"evaluationId"`
	ServiceID    string `json:"serviceId"`
	AgreementID  string `json:"agreementId"`
	TxID         string `json:"txId"`
	Hash         string `json:"hash"`
}

// EvaluationResult represents for a evaluation result.
type EvaluationResult struct {
	Satisfied     bool         `json:"satisfied"`
	PenaltyRule   *PenaltyRule `json:"penaltyRule,omitempty" metadata:"penaltyRule,optional"`
	FailureReason string       `json:"failureReason,omitempty" metadata:"failureReason,optional"`
}

// AccessKey represents for a access key.
type AccessKey struct {
	Type  string `json:"type"`
	Token string `json:"key"`
}
