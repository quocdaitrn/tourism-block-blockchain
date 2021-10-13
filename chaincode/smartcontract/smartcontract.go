package smartcontract

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Index keys
const (
	serviceIndex    = "doc~service"
	evaluationIndex = "doc~evaluation"
)

const (
	JWTInternalServiceAccessKey = "jwt_internal_service_access_key"
)

// SmartContract provides functions for managing service rating.
type SmartContract struct {
	contractapi.Contract
}

// CreateOrUpdateInternalServiceAccessKey creates or updates internal service access key.
func (s *SmartContract) CreateOrUpdateInternalServiceAccessKey(ctx contractapi.TransactionContextInterface, token string) error {
	accessKey := AccessKey{
		Type:  "Bearer",
		Token: token,
	}

	jAccessKey, err := json.Marshal(accessKey)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(JWTInternalServiceAccessKey, jAccessKey)
}

// ReadInternalServiceAccessKey returns the internal service access key.
func (s *SmartContract) ReadInternalServiceAccessKey(ctx contractapi.TransactionContextInterface) (*AccessKey, error) {
	jAccessKey, err := ctx.GetStub().GetState(JWTInternalServiceAccessKey)
	if err != nil {
		return nil, err
	}
	if jAccessKey == nil {
		return nil, fmt.Errorf("the internal service access key does not exist")
	}

	var accessKey AccessKey
	err = json.Unmarshal(jAccessKey, &accessKey)
	if err != nil {
		return nil, err
	}

	return &accessKey, nil
}

// CreateService issues a new service to the world state with given details.
func (s *SmartContract) CreateService(ctx contractapi.TransactionContextInterface, id string) error {
	exist, err := s.ServiceExists(ctx, id)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the service %s already exists", id)
	}

	service := Service{
		DocType:          "Service",
		ServiceID:        id,
		SatisfactionRate: 1,
		RuleAbidingRate:  1,
		Agreements:       []*Agreement{},
	}

	jService, err := json.Marshal(service)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, jService)
	if err != nil {
		return fmt.Errorf("can not create service: %s", id)
	}

	docServiceIndexKey, err := ctx.GetStub().CreateCompositeKey(serviceIndex, []string{service.DocType, service.ServiceID})
	if err != nil {
		return err
	}
	//  Save serviceIndex entry to world state. Only the key name is needed, no need to store a duplicate copy of the service.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	return ctx.GetStub().PutState(docServiceIndexKey, value)
}

// ReadService returns the service stored in the world state with given id.
func (s *SmartContract) ReadService(ctx contractapi.TransactionContextInterface, id string) (*Service, error) {
	jService, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if jService == nil {
		return nil, fmt.Errorf("the service %s does not exist", id)
	}

	var service Service
	err = json.Unmarshal(jService, &service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

// DeleteService deletes an given provider from the world state.
func (s *SmartContract) DeleteService(ctx contractapi.TransactionContextInterface, id string) error {
	service, err := s.ReadService(ctx, id)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(id)
	if err != nil {
		return fmt.Errorf("can not delete the service %s", id)
	}

	docServiceIndexKey, err := ctx.GetStub().CreateCompositeKey(serviceIndex, []string{service.DocType, service.ServiceID})
	if err != nil {
		return err
	}

	// Delete serviceIndex entry
	return ctx.GetStub().DelState(docServiceIndexKey)
}

// GetAllServices returns all services found in world state.
func (s *SmartContract) GetAllServices(ctx contractapi.TransactionContextInterface) ([]*Service, error) {
	serviceResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(serviceIndex, []string{"Service"})
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	defer serviceResultsIterator.Close()

	var services []*Service
	for serviceResultsIterator.HasNext() {
		rangeResponse, err := serviceResultsIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(rangeResponse.Key)
		if err != nil {
			return nil, err
		}

		if len(compositeKeyParts) > 1 {
			returnedServiceID := compositeKeyParts[1]
			service, err := s.ReadService(ctx, returnedServiceID)
			if err != nil {
				return nil, err
			}
			services = append(services, service)
		}
	}

	return services, nil
}

// ServiceExists returns true when service with given id exists in world state.
func (s *SmartContract) ServiceExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	exist, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read service from world state: %v", err)
	}

	return exist != nil, nil
}

// AddAgreement adds a agreement to a service.
func (s *SmartContract) AddAgreement(ctx contractapi.TransactionContextInterface, sid, aid, cat string, hasPenalty bool, items, penaltyRules string) (*Service, error) {
	exist, err := s.ServiceExists(ctx, sid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("the service %s does not exist", sid)
	}

	service, err := s.ReadService(ctx, sid)
	if err != nil {
		return nil, err
	}

	for _, a := range service.Agreements {
		if a.AgreementID == aid {
			return nil, fmt.Errorf("the agreement %s already exist", aid)
		}
	}

	dItems, err := b64.StdEncoding.DecodeString(items)
	if err != nil {
		return nil, fmt.Errorf("can not decode agreement items from base64: %v", err)
	}
	var aItems []*AgreementItem
	err = json.Unmarshal(dItems, &aItems)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal agreement items: %v", err)
	}

	dPenaltyRules, err := b64.StdEncoding.DecodeString(penaltyRules)
	if err != nil {
		return nil, fmt.Errorf("can not decode penalty rules from base64: %v", err)
	}
	var aPenaltyRules []*PenaltyRule
	err = json.Unmarshal(dPenaltyRules, &aPenaltyRules)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal penalty rules: %v", err)
	}

	agreement := &Agreement{
		AgreementID:                            aid,
		Category:                               cat,
		Items:                                  aItems,
		TotalFeedbacks:                         0,
		TotalUnsatisfied:                       0,
		TotalRuleViolations:                    0,
		TotalRuleViolationWithoutCompensations: 0,
		HasPenaltyRule:                         hasPenalty,
		PenaltyRules:                           aPenaltyRules,
		RuleAbidingRate:                        1.0,
		SatisfactionRate:                       1.0,
	}
	service.Agreements = append(service.Agreements, agreement)

	jService, err := json.Marshal(service)
	err = ctx.GetStub().PutState(sid, jService)
	if err != nil {
		return nil, fmt.Errorf("fail to add agreement %s to service %s", aid, sid)
	}

	return service, nil
}

// UpdateAgreement updates a agreement in a service.
func (s *SmartContract) UpdateAgreement(ctx contractapi.TransactionContextInterface, sid, aid, cat string, hasPenalty bool, items, penaltyRules string) (*Service, error) {
	exist, err := s.ServiceExists(ctx, sid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("the service %s does not exist", sid)
	}

	service, err := s.ReadService(ctx, sid)
	if err != nil {
		return nil, err
	}

	aIndex := -1
	for i, a := range service.Agreements {
		if a.AgreementID == aid {
			aIndex = i
			break
		}
	}
	if aIndex == -1 {
		return nil, fmt.Errorf("the agreement %s does not exist", aid)
	}

	dItems, err := b64.StdEncoding.DecodeString(items)
	if err != nil {
		return nil, fmt.Errorf("can not decode agreement items from base64: %v", err)
	}
	var aItems []*AgreementItem
	err = json.Unmarshal(dItems, &aItems)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal agreement items: %v", err)
	}

	dPenaltyRules, err := b64.StdEncoding.DecodeString(penaltyRules)
	if err != nil {
		return nil, fmt.Errorf("can not decode penalty rules from base64: %v", err)
	}
	var aPenaltyRules []*PenaltyRule
	err = json.Unmarshal(dPenaltyRules, &aPenaltyRules)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal penalty rules: %v", err)
	}

	service.Agreements[aIndex].Category = cat
	service.Agreements[aIndex].Items = aItems
	service.Agreements[aIndex].HasPenaltyRule = hasPenalty
	service.Agreements[aIndex].PenaltyRules = aPenaltyRules

	jService, err := json.Marshal(service)
	err = ctx.GetStub().PutState(sid, jService)
	if err != nil {
		return nil, fmt.Errorf("fail to update agreement %s in service %s", aid, sid)
	}

	return service, nil
}

// RemoveAgreement removes a agreement from a service.
func (s *SmartContract) RemoveAgreement(ctx contractapi.TransactionContextInterface, sid, aid string) (*Service, error) {
	exist, err := s.ServiceExists(ctx, sid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("the service %s does not exist", sid)
	}

	service, err := s.ReadService(ctx, sid)
	if err != nil {
		return nil, err
	}

	aIndex := -1
	for i, a := range service.Agreements {
		if a.AgreementID == aid {
			aIndex = i
			break
		}
	}
	if aIndex == -1 {
		return nil, fmt.Errorf("the agreement %s does not exist", aid)
	}

	agreements := append(service.Agreements[:aIndex], service.Agreements[aIndex+1:]...)
	service.Agreements = agreements

	minSatisfactionRate := float32(1)
	minRuleAbidingRate := float32(1)
	if len(agreements) > 0 {
		minSatisfactionRate = agreements[0].SatisfactionRate
		minRuleAbidingRate = agreements[0].RuleAbidingRate
	}

	for _, a := range agreements {
		if minSatisfactionRate > a.SatisfactionRate {
			minSatisfactionRate = a.SatisfactionRate
		}
		if minRuleAbidingRate > a.RuleAbidingRate {
			minRuleAbidingRate = a.RuleAbidingRate
		}
	}

	service.SatisfactionRate = minSatisfactionRate
	service.RuleAbidingRate = minRuleAbidingRate

	jService, err := json.Marshal(service)
	err = ctx.GetStub().PutState(sid, jService)
	if err != nil {
		return nil, fmt.Errorf("fail to remove agreement %s from service %s", aid, sid)
	}

	return service, nil
}

// EvaluateSLA handles evaluating SLA request.
func (s *SmartContract) EvaluateSLA(ctx contractapi.TransactionContextInterface, sid, aid, eid, eData, hash, at string) (*EvaluationResult, error) {
	exist, err := s.ServiceExists(ctx, sid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("the service %s does not exist", sid)
	}

	service, err := s.ReadService(ctx, sid)
	if err != nil {
		return nil, err
	}

	aIndex := -1
	for i, a := range service.Agreements {
		if a.AgreementID == aid {
			aIndex = i
			break
		}
	}
	if aIndex == -1 {
		return nil, fmt.Errorf("the agreement %s does not exist", aid)
	}

	dEvaData, err := b64.StdEncoding.DecodeString(eData)
	if err != nil {
		return nil, fmt.Errorf("can not decode evaluation data from base64: %v", err)
	}
	var evaData []*EvaluationData
	err = json.Unmarshal(dEvaData, &evaData)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal evaluation data: %v", err)
	}

	agreement := service.Agreements[aIndex]
	eResult, err := s.VerifySLA(ctx, agreement, evaData)
	if err != nil {
		return nil, err
	}

	agreement.TotalFeedbacks++
	if !eResult.Satisfied {
		agreement.TotalUnsatisfied++
	}
	agreement.SatisfactionRate = float32(agreement.TotalFeedbacks-
		agreement.TotalUnsatisfied) / float32(agreement.TotalFeedbacks)

	agreement.LastEvaluationAt = at
	minSatisfactionRate := service.Agreements[0].SatisfactionRate
	for _, a := range service.Agreements {
		if a.SatisfactionRate < minSatisfactionRate {
			minSatisfactionRate = a.SatisfactionRate
		}
	}
	service.SatisfactionRate = minSatisfactionRate
	service.NumberOfEvaluations++
	service.LastEvaluationAt = at

	jService, err := json.Marshal(service)
	err = ctx.GetStub().PutState(sid, jService)
	if err != nil {
		return nil, fmt.Errorf("fail to update satisfaction rate for service %s", sid)
	}

	evaluation := &Evaluation{
		DocType:      "Evaluation",
		EvaluationID: eid,
		ServiceID:    sid,
		AgreementID:  aid,
		TxID:         ctx.GetStub().GetTxID(),
		Hash:         hash,
	}
	jEvaluation, _ := json.Marshal(evaluation)
	_ = ctx.GetStub().PutState(eid, jEvaluation)

	docEvaluationIndexKey, err := ctx.GetStub().CreateCompositeKey(evaluationIndex, []string{evaluation.DocType, evaluation.EvaluationID})
	if err != nil {
		return nil, err
	}
	//  Save serviceIndex entry to world state. Only the key name is needed, no need to store a duplicate copy of the evaluation.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	err = ctx.GetStub().PutState(docEvaluationIndexKey, value)
	if err != nil {
		return nil, err
	}

	return eResult, nil
}

// UpdateRuleAbidingRate handles updating SLA rule-abiding rate request.
func (s *SmartContract) UpdateRuleAbidingRate(ctx contractapi.TransactionContextInterface, sid, aid, eid, hash, at string, compensated bool) (*Service, error) {
	exist, err := s.ServiceExists(ctx, sid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("the service %s does not exist", sid)
	}

	service, err := s.ReadService(ctx, sid)
	if err != nil {
		return nil, err
	}

	aIndex := -1
	for i, a := range service.Agreements {
		if a.AgreementID == aid {
			aIndex = i
			break
		}
	}
	if aIndex == -1 {
		return nil, fmt.Errorf("the agreement %s does not exist", aid)
	}

	agreement := service.Agreements[aIndex]
	agreement.TotalRuleViolations++
	if !compensated {
		agreement.TotalRuleViolationWithoutCompensations++
	}

	agreement.RuleAbidingRate = float32(agreement.TotalRuleViolations-
		agreement.TotalRuleViolationWithoutCompensations) / float32(agreement.TotalRuleViolations)

	agreement.LastEvaluationAt = at
	minRuleAbidingRate := service.Agreements[0].RuleAbidingRate
	for _, a := range service.Agreements {
		if a.RuleAbidingRate < minRuleAbidingRate {
			minRuleAbidingRate = a.RuleAbidingRate
		}
	}
	service.RuleAbidingRate = minRuleAbidingRate
	service.NumberOfEvaluations++
	service.LastEvaluationAt = at

	jService, err := json.Marshal(service)
	err = ctx.GetStub().PutState(sid, jService)
	if err != nil {
		return nil, fmt.Errorf("fail to update rule-abiding rate for service %s", sid)
	}

	evaluation := &Evaluation{
		DocType:      "Evaluation",
		EvaluationID: eid,
		ServiceID:    sid,
		AgreementID:  aid,
		TxID:         ctx.GetStub().GetTxID(),
		Hash:         hash,
	}
	jEvaluation, _ := json.Marshal(evaluation)
	_ = ctx.GetStub().PutState(eid, jEvaluation)

	docEvaluationIndexKey, err := ctx.GetStub().CreateCompositeKey(evaluationIndex, []string{evaluation.DocType, evaluation.EvaluationID})
	if err != nil {
		return nil, err
	}
	//  Save serviceIndex entry to world state. Only the key name is needed, no need to store a duplicate copy of the evalution.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	err = ctx.GetStub().PutState(docEvaluationIndexKey, value)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// VerifySLA verifies SLA agreement.
func (s *SmartContract) VerifySLA(ctx contractapi.TransactionContextInterface, a *Agreement, eData []*EvaluationData) (*EvaluationResult, error) {
	if !StringInSlice(a.Category, SupportedAgreementCategories) {
		return nil, fmt.Errorf("agreement category %s has not supported yet", a.Category)
	}

	var penaltyRule *PenaltyRule
	if a.HasPenaltyRule {
		penaltyRule = a.PenaltyRules[0]
	}

	if len(eData) == 0 {
		return &EvaluationResult{
			Satisfied:     false,
			PenaltyRule:   penaltyRule,
			FailureReason: "no data to evaluate",
		}, nil
	}

	switch a.Category {
	case AgreementCategoryService:
		switch a.Items[0].Code {
		case AgreementItemCodeServiceAirportShuttle:
			return s.VerifyAirportShuttleAgreement(ctx, a, eData[0])
		case AgreementItemCodeServiceSauna:
			return s.VerifySaunaAgreement(ctx, a, eData[0])
		default:
			return nil, MakeErrorAgreementItemCodeDoesNotSupport(a.Items[0].Code, a.Category)
		}
	case AgreementCategoryRoomDesign:
		switch a.Items[0].Code {
		case AgreementItemCodeRoomDesignSize:
			return s.VerifyRoomSizeAgreement(ctx, a, eData[0])
		default:
			return nil, MakeErrorAgreementItemCodeDoesNotSupport(a.Items[0].Code, a.Category)
		}
	case AgreementCategoryBed:
		return s.VerifyBedAgreement(ctx, a, eData)
	default:
		satisfied := true
		for _, agreementItem := range a.Items {
			satisfied = false
			for _, item := range eData {
				if a.Category == AgreementCategoryView {
					if ViewLevelMapping[item.Code] >= ViewLevelMapping[agreementItem.Code] {
						satisfied = true
						break
					}
				} else {
					if item.Code == agreementItem.Code && (agreementItem.Quantity == 0 || item.Quantity >= agreementItem.Quantity) {
						satisfied = true
						break
					}
				}
			}
			if !satisfied {
				break
			}
		}

		return &EvaluationResult{
			Satisfied:   satisfied,
			PenaltyRule: penaltyRule,
		}, nil
	}
}

// VerifyAirportShuttleAgreement verify airport shuttle agreement.
func (s *SmartContract) VerifyAirportShuttleAgreement(_ contractapi.TransactionContextInterface, a *Agreement, data *EvaluationData) (*EvaluationResult, error) {
	if data.Code != AgreementItemCodeServiceAirportShuttle {
		return nil, fmt.Errorf("evaluation data code %s is not airport shuttle item code", data.Code)
	}

	if data.Status == AirportShuttleStatusDriverWaiting || data.Status == AirportShuttleStatusInService {
		return nil, fmt.Errorf("can not verify airport shuttle agreement for status: driver_waiting or in_service")
	}

	// customer cancels service
	if data.Status == AirportShuttleStatusCanceled {
		return &EvaluationResult{
			Satisfied: true,
		}, nil
	}

	// driver doesn't show up
	if data.Status == AirportShuttleStatusNotServed {
		return &EvaluationResult{
			Satisfied:     false,
			PenaltyRule:   a.PenaltyRules[0],
			FailureReason: "driver did not come to pick up the passenger",
		}, nil
	}

	if data.Status == AirportShuttleStatusWaitingTimeExceeded {
		if data.DriverNotifyCustomerDoNotShowUpAt.Sub(data.PickUpTime).Minutes() > float64(a.Items[0].DriverMaxWaitTime) &&
			data.DriverArriveAt.Sub(data.PickUpTime).Minutes() <= float64(a.Items[0].CustomerShortWaitTime) {
			return &EvaluationResult{
				Satisfied: true,
			}, nil
		} else {
			return &EvaluationResult{
				Satisfied:     false,
				PenaltyRule:   a.PenaltyRules[0],
				FailureReason: "the driver came to pick up the passenger late",
			}, nil
		}
	}

	// case 'completed'
	driverDelayTime := data.DriverArriveAt.Sub(data.PickUpTime).Minutes()
	switch {
	case driverDelayTime <= float64(a.Items[0].CustomerShortWaitTime):
		return &EvaluationResult{
			Satisfied: true,
		}, nil
	case driverDelayTime <= float64(a.Items[0].CustomerLongWaitTime):
		return &EvaluationResult{
			Satisfied:     false,
			PenaltyRule:   a.PenaltyRules[1],
			FailureReason: "the driver came to pick up the passenger late",
		}, nil
	default:
		return &EvaluationResult{
			Satisfied:     false,
			PenaltyRule:   a.PenaltyRules[2],
			FailureReason: "the driver came to pick up the passenger late",
		}, nil
	}
}

// VerifySaunaAgreement verify sauna agreement.
func (s *SmartContract) VerifySaunaAgreement(_ contractapi.TransactionContextInterface, a *Agreement, data *EvaluationData) (*EvaluationResult, error) {
	if data.Code != AgreementItemCodeServiceSauna {
		return nil, fmt.Errorf("evaluation data code %s is not sauna service item code", data.Code)
	}

	requests := data.SaunaRequests

	// sort requests by request time
	sort.SliceStable(requests, func(i, j int) bool {
		return requests[i].RequestAt.Before(requests[j].RequestAt)
	})

	idx := -1
	numFailures := 0
	for i, req := range requests {
		if req.Status == SaunaRequestStatusFail {
			if idx >= 0 && req.RequestAt.Sub(requests[idx].RequestAt).Minutes() >= float64(a.Items[0].MinTimeBetween2Failures) {
				numFailures++
			}
			idx = i
		}

		if numFailures >= a.Items[0].MaxFailures {
			var penaltyRule *PenaltyRule
			if a.HasPenaltyRule {
				penaltyRule = a.PenaltyRules[0]
			}
			return &EvaluationResult{
				Satisfied:   false,
				PenaltyRule: penaltyRule,
			}, nil
		}
	}

	return &EvaluationResult{
		Satisfied: true,
	}, nil
}

// VerifySaunaAgreement verify sauna agreement.
func (s *SmartContract) VerifyRoomSizeAgreement(_ contractapi.TransactionContextInterface, a *Agreement, data *EvaluationData) (*EvaluationResult, error) {
	if data.Code != AgreementItemCodeRoomDesignSize {
		return nil, fmt.Errorf("evaluation data code %s is not room size code", data.Code)
	}

	var penaltyRule *PenaltyRule
	if a.HasPenaltyRule {
		penaltyRule = a.PenaltyRules[0]
	}
	return &EvaluationResult{
		Satisfied:   data.Value.(float64) >= a.Items[0].Value.(float64),
		PenaltyRule: penaltyRule,
	}, nil
}

// VerifyBedAgreement verify bed agreement.
func (s *SmartContract) VerifyBedAgreement(_ contractapi.TransactionContextInterface, a *Agreement, data []*EvaluationData) (*EvaluationResult, error) {
	aNumOfBeds := 0
	rNumOfBeds := 0
	aTotalPoints := 0
	rTotalPoints := 0

	for _, agreementItem := range a.Items {
		aTotalPoints += BedPointMapping[agreementItem.Code]
		aNumOfBeds += agreementItem.Quantity
	}
	for _, item := range data {
		rTotalPoints += BedPointMapping[item.Code]
		rNumOfBeds += item.Quantity
	}

	var penaltyRule *PenaltyRule
	if a.HasPenaltyRule {
		penaltyRule = a.PenaltyRules[0]
	}

	return &EvaluationResult{
		Satisfied:   rNumOfBeds >= aNumOfBeds && rTotalPoints >= aTotalPoints,
		PenaltyRule: penaltyRule,
	}, nil
}

// EnforcePenaltyRuleFromBlockChain enforces penalty rule from blockchain.
func (s *SmartContract) EnforcePenaltyRuleFromBlockChain(_ contractapi.TransactionContextInterface, rid, aid string, eResult *EvaluationResult, token string) error {
	enforcePenaltyRulesRequest := &EnforcePenaltyRulesRequest{
		ReservationID: rid,
		AgreementID:   aid,
	}

	if eResult != nil {
		enforcePenaltyRulesRequest.PenaltyRules = []*PenaltyRule{eResult.PenaltyRule}
		enforcePenaltyRulesRequest.Reason = eResult.FailureReason
	}

	err := EnforcePenaltyRules(enforcePenaltyRulesRequest, token)
	if err != nil {
		return fmt.Errorf("fail to enforce penalty rule %v", err)
	}

	return nil
}

// HandleSatisfactionEvaluationEvent calculate satisfaction rate when
// receiving satisfaction evaluation for a agreement.
func (s *SmartContract) HandleSatisfactionEvaluationEvent(ctx contractapi.TransactionContextInterface, sid, aid, eid, rid, hash, at string, satisfied, enforcePenaltyRule bool) (*Service, error) {
	exist, err := s.ServiceExists(ctx, sid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("the service %s does not exist", sid)
	}

	service, err := s.ReadService(ctx, sid)
	if err != nil {
		return nil, err
	}

	aIndex := -1
	for i, a := range service.Agreements {
		if a.AgreementID == aid {
			aIndex = i
			break
		}
	}
	if aIndex == -1 {
		return nil, fmt.Errorf("the agreement %s does not exist", aid)
	}

	agreement := service.Agreements[aIndex]
	agreement.TotalFeedbacks++
	if !satisfied {
		agreement.TotalUnsatisfied++
		if agreement.HasPenaltyRule && enforcePenaltyRule {
			accessKey, err := s.ReadInternalServiceAccessKey(ctx)
			if err == nil {
				_ = s.EnforcePenaltyRuleFromBlockChain(ctx, rid, aid, nil, accessKey.Token)
			}
		}
	}
	agreement.SatisfactionRate = float32(agreement.TotalFeedbacks-
		agreement.TotalUnsatisfied) / float32(agreement.TotalFeedbacks)

	agreement.LastEvaluationAt = at
	minSatisfactionRate := service.Agreements[0].SatisfactionRate
	for _, a := range service.Agreements {
		if a.SatisfactionRate < minSatisfactionRate {
			minSatisfactionRate = a.SatisfactionRate
		}
	}
	service.SatisfactionRate = minSatisfactionRate
	service.NumberOfEvaluations++
	service.LastEvaluationAt = at

	jService, err := json.Marshal(service)
	err = ctx.GetStub().PutState(sid, jService)
	if err != nil {
		return nil, fmt.Errorf("fail to update satisfaction rate for service %s", sid)
	}

	evaluation := &Evaluation{
		DocType:      "Evaluation",
		EvaluationID: eid,
		ServiceID:    sid,
		AgreementID:  aid,
		TxID:         ctx.GetStub().GetTxID(),
		Hash:         hash,
	}
	jEvaluation, _ := json.Marshal(evaluation)
	_ = ctx.GetStub().PutState(eid, jEvaluation)

	docEvaluationIndexKey, err := ctx.GetStub().CreateCompositeKey(evaluationIndex, []string{evaluation.DocType, evaluation.EvaluationID})
	if err != nil {
		return nil, err
	}
	//  Save serviceIndex entry to world state. Only the key name is needed, no need to store a duplicate copy of the evaluation.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	err = ctx.GetStub().PutState(docEvaluationIndexKey, value)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// HandlePenaltyRuleEvaluationEvent calculate rule-abiding rate.
func (s *SmartContract) HandlePenaltyRuleEvaluationEvent(ctx contractapi.TransactionContextInterface, sid, aid, eid, hash, at string, compensated bool) (*Service, error) {
	exist, err := s.ServiceExists(ctx, sid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("the service %s does not exist", sid)
	}

	service, err := s.ReadService(ctx, sid)
	if err != nil {
		return nil, err
	}

	aIndex := -1
	for i, a := range service.Agreements {
		if a.AgreementID == aid {
			aIndex = i
			break
		}
	}
	if aIndex == -1 {
		return nil, fmt.Errorf("the agreement %s does not exist", aid)
	}

	agreement := service.Agreements[aIndex]
	agreement.TotalRuleViolations++
	if !compensated {
		agreement.TotalRuleViolationWithoutCompensations++
	}

	agreement.RuleAbidingRate = float32(agreement.TotalRuleViolations-
		agreement.TotalRuleViolationWithoutCompensations) / float32(agreement.TotalRuleViolations)

	agreement.LastEvaluationAt = at
	minRuleAbidingRate := service.Agreements[0].RuleAbidingRate
	for _, a := range service.Agreements {
		if a.RuleAbidingRate < minRuleAbidingRate {
			minRuleAbidingRate = a.RuleAbidingRate
		}
	}
	service.RuleAbidingRate = minRuleAbidingRate
	service.NumberOfEvaluations++
	service.LastEvaluationAt = at

	jService, err := json.Marshal(service)
	err = ctx.GetStub().PutState(sid, jService)
	if err != nil {
		return nil, fmt.Errorf("fail to update rule-abiding rate for service %s", sid)
	}

	evaluation := &Evaluation{
		DocType:      "Evaluation",
		EvaluationID: eid,
		ServiceID:    sid,
		AgreementID:  aid,
		TxID:         ctx.GetStub().GetTxID(),
		Hash:         hash,
	}
	jEvaluation, _ := json.Marshal(evaluation)
	_ = ctx.GetStub().PutState(eid, jEvaluation)

	docEvaluationIndexKey, err := ctx.GetStub().CreateCompositeKey(evaluationIndex, []string{evaluation.DocType, evaluation.EvaluationID})
	if err != nil {
		return nil, err
	}
	//  Save serviceIndex entry to world state. Only the key name is needed, no need to store a duplicate copy of the evalution.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	err = ctx.GetStub().PutState(docEvaluationIndexKey, value)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// CountAllEvaluations returns number of evaluations.
func (s *SmartContract) CountAllEvaluations(ctx contractapi.TransactionContextInterface, pageSize int) (int32, error) {
	evaluationResultsIterator, responseMetadata, err := ctx.GetStub().GetQueryResultWithPagination(`{"selector":{"docType":"Evaluation"}}`, int32(pageSize), "")
	if err != nil {
		return -1, fmt.Errorf("failed to count number of evaluations: %v", err)
	}

	defer evaluationResultsIterator.Close()

	return responseMetadata.FetchedRecordsCount, nil
}

// PostPenaltyRule posts penalty rule.
func (s *SmartContract) PostPenaltyRule(_ contractapi.TransactionContextInterface) (string, error) {
	enforcePenaltyRulesRequest := &EnforcePenaltyRulesRequest{
		ReservationID: "rid",
		AgreementID:   "aid",
	}

	err := EnforcePenaltyRules(enforcePenaltyRulesRequest, "token")
	if err != nil {
		return "", fmt.Errorf("fail to enforce penalty rule %v", err)
	}

	return "success", nil
}

// PostUsers posts users.
func (s *SmartContract) PostUsers(_ contractapi.TransactionContextInterface) error {
	reqBody, err := json.Marshal(map[string]interface{}{
		"title":  "foo",
		"body":   "bar",
		"userId": 1,
	})

	if err != nil {
		return err
	}

	url := "https://jsonplaceholder.typicode.com/posts"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf(fmt.Sprintf("PostUsers with response :%d", resp.StatusCode))
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// TestTime tests time.
func (s *SmartContract) TestTime(_ contractapi.TransactionContextInterface, data string) (*SaunaRequest, error) {
	dEvaData, err := b64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("can not decode evaluation data from base64: %v", err)
	}
	var evaData []*EvaluationData
	err = json.Unmarshal(dEvaData, &evaData)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal evaluation data: %v", err)
	}

	requests := evaData[0].SaunaRequests

	// sort requests by request time
	sort.SliceStable(requests, func(i, j int) bool {
		return requests[i].RequestAt.After(requests[j].RequestAt)
	})

	if requests[0].RequestAt.Before(requests[1].RequestAt) {
		return requests[0], nil
	}

	return requests[1], nil
}
