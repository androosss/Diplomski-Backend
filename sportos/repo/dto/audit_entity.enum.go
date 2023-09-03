package dto

// [swagger]

// SportosEntity
//
// Sportos database entity name. Possible values:
//   * 'currency'
//   * 'fraud_provider'
//   * 'pam'
//   * 'partner'
//   * 'partner_fraud_provider_profile"
//   * 'partner_payment_provider_profile'
//   * 'payment_instrument'
//   * 'payment_instrument_template'
//   * 'payment_method'
//   * 'payment_provider'
//   * 'payment_request'
//   * 'payment_route'
//   * 'player'
//   * 'schedule'
//   * 'user'
// swagger:model SportosEntity
type SportosEntity string

const (
	ENTITY_PLAYER = "player"
	ENTITY_USER   = "user"
	ENTITY_PLACE  = "place"
	ENTITY_COACH  = "coach"
	ENTITY_EVENT  = "event"
)

func (tpe SportosEntity) GetName() string {
	switch tpe {
	default:
		return KeyToName(string(tpe))
	}
}

func (tpe SportosEntity) IsValid() bool {
	switch tpe {
	case ENTITY_PLAYER, ENTITY_USER:
		return true
	}
	return false
}
