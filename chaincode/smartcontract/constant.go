package smartcontract

const (
	// BaseURL base url to call to server.
	BaseURL = "http://doan.vinaictgroup.com:8181/tourism-block/v1"
)

// list of categories of agreement.
const (
	AgreementCategoryView       = "view"
	AgreementCategoryService    = "service"
	AgreementCategoryInterior   = "interior"
	AgreementCategoryRoomDesign = "room_design"
	AgreementCategoryOutdoor    = "outdoor"
	AgreementCategoryBed        = "bed"
	AgreementCategoryFacility   = "facility" // TODO: deprecated
)

// list of penalty rule types.
const (
	PenaltyRuleTypeDiscount     = "discount"
	PenaltyRuleTypeUpgradeLevel = "upgrade_level"
)

// agreement item code.
const (
	AgreementItemCodeViewSea      = "V001"
	AgreementItemCodeViewRiver    = "V002"
	AgreementItemCodeViewPool     = "V003"
	AgreementItemCodeViewGarden   = "V004"
	AgreementItemCodeViewCity     = "V005"
	AgreementItemCodeViewMountain = "V006"

	AgreementItemCodeBedTwin  = "BE001"
	AgreementItemCodeBedQueen = "BE002"
	AgreementItemCodeBedKing  = "BE003"

	AgreementItemCodeInteriorBathtub      = "IBA001"
	AgreementItemCodeInteriorFlatScreenTV = "ITV001"

	AgreementItemCodeServiceSauna          = "SSA001"
	AgreementItemCodeServiceAirportShuttle = "SAS001"

	AgreementItemCodeOutdoorPatio   = "OPA001"
	AgreementItemCodeOutdoorBalcony = "OBA001"

	AgreementItemCodeRoomDesignSize = "RSI001"
)

// SupportedAgreementCategories supported agreement categories.
var SupportedAgreementCategories = []string{
	AgreementCategoryView,
	AgreementCategoryService,
	AgreementCategoryInterior,
	AgreementCategoryRoomDesign,
	AgreementCategoryOutdoor,
	AgreementCategoryBed,
}

// ViewLevelMapping maps a view to it's level.
var ViewLevelMapping = map[string]int{
	AgreementItemCodeViewMountain: 0,
	AgreementItemCodeViewCity:     1,
	AgreementItemCodeViewGarden:   2,
	AgreementItemCodeViewPool:     3,
	AgreementItemCodeViewRiver:    4,
	AgreementItemCodeViewSea:      5,
}

// BedPointMapping maps a kind of bed to it's points.
var BedPointMapping = map[string]int{
	AgreementItemCodeBedTwin:  1,
	AgreementItemCodeBedQueen: 10,
	AgreementItemCodeBedKing:  100,
}

// airport shuttle statuses.
const (
	AirportShuttleStatusConfirmed           = "confirmed"
	AirportShuttleStatusDriverWaiting       = "driver_waiting"
	AirportShuttleStatusInService           = "in_service"
	AirportShuttleStatusCompleted           = "completed"
	AirportShuttleStatusNotServed           = "not_served"
	AirportShuttleStatusWaitingTimeExceeded = "waiting_time_exceeded"
	AirportShuttleStatusCanceled            = "canceled"
)

// sauna request statuses.
const (
	SaunaRequestStatusSuccess = "success"
	SaunaRequestStatusFail    = "fail"
)
