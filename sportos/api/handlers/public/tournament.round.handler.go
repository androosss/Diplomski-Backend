// swagger:operation POST /partner-payment-provider-profiles PartnerPaymentProviderProfileCreate
//
// Create a new payment provider merchant profile for partner
// ---
// summary: Create payment provider profile for partner
// operationId: PartnerPaymentProviderProfileCreate
// tags:
// - Backoffice API
// produces:
// - application/json
// parameters:
//   - name: X-Request-ID
//     in: header
//     description: Unique request ID
//     required: true
//     type: string
//   - name: User-Id
//     in: header
//     description: Id of backoffice user
//     required: true
//     type: string
//   - in: body
//     name: partnerPaymentProviderProfile
//     description: profile that should be created for partner
//     required: true
//     schema:
//     $ref: '#/definitions/PartnerPaymentProviderProfile'
//
// responses:
//
//	200:
//	  description: results
//	  schema:
//	    $ref: '#/definitions/PartnerPaymentProviderProfile'
//	400:
//	  description: Invalid request supplied
//	  schema:
//	    $ref: '#/definitions/APIError'
//	403:
//	  description: Request is forbidden
//	  schema:
//	    $ref: '#/definitions/APIError'
//	404:
//	  description: Predefined error occured
//	  schema:
//	    $ref: '#/definitions/APIError'
//	405:
//	  description: Method not allowed error
//	  schema:
//	    $ref: '#/definitions/APIError'
//	409:
//	  description: Conflict error, payment provider error
//	  schema:
//	    $ref: '#/definitions/APIError'
//	500:
//	  description: Internal server error
//	  schema:
//	    $ref: '#/definitions/APIError'

package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type TournamentRoundPostRequest struct {
	Id    string    `json:"id"`
	Round *DR.Round `json:"round,omitempty"`
}

type TournamentRoundPostHandler struct {
	TournamentRoundPostRequest
}

type TournamentRoundPostResponse struct {
	NextRound DR.Round      `json:"round,omitempty"`
	Standings []DR.Standing `json:"standings,omitempty"`
}

func (r TournamentRoundPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r TournamentRoundPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *TournamentRoundPostHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.TournamentRoundPostRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *TournamentRoundPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	_, err := Repo.EventCrud.GetById(ctx, r.Id, nil)
	if err != nil {
		return DA.ErrorBadRequest().WithMessage("Event doesn't exist")
	}
	if r.Round != nil {
		for _, elem := range r.Round.Pairing {
			if elem.Score == "" {
				return DA.ErrorBadRequest().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Not all results are submitted")
			}
			if len(strings.Split(elem.Score, ":")) != 2 {
				return DA.ErrorBadRequest().WithMessage("Result format isn't valid")
			}
			if _, err = strconv.Atoi(strings.Split(elem.Score, ":")[0]); err != nil {
				return DA.ErrorBadRequest().WithMessage("Result format isn't valid")
			}
			if _, err = strconv.Atoi(strings.Split(elem.Score, ":")[1]); err != nil {
				return DA.ErrorBadRequest().WithMessage("Result format isn't valid")
			}
		}
	}
	return nil
}

func (r *TournamentRoundPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	event, err := Repo.EventCrud.GetById(ctx, r.Id, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	err = updateTournament(ctx, Repo, &event, r.Round)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	nextRound(ctx, Repo, &event)
	up := DR.EventUpdateParams{
		Id:         r.Id,
		Tournament: event.Tournament,
	}
	if r.Round == nil && len(event.Tournament.Rounds) == 1 {
		active := DR.ES_ACTIVE
		up.Status = &active
	}
	event, err = Repo.EventCrud.Update(ctx, up, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = event
	return resMap, nil
}

// TODO update stats && DB
func updateTournament(ctx context.Context, Repo *crud.Repo, event *DR.Event, round *DR.Round) error {
	if event.Tournament == nil {
		event.Tournament = &DR.Tournament{}
	}
	if event.Tournament.Standings == nil {
		event.Tournament.Standings = make([]DR.Standing, 0)
		for i, elem := range event.Teams {
			standing := DR.Standing{
				TeamName: elem.Name,
				Points:   new(int),
				Ranking:  new(int),
			}
			*standing.Points = 0
			*standing.Ranking = i + 1
			event.Tournament.Standings = append(event.Tournament.Standings, standing)
		}
	}
	if len(event.Tournament.Rounds) > len(event.Teams)-1 {
		return nil
	}
	if round == nil {
		if event.Status == DR.ES_CREATED && len(event.Tournament.Rounds) == 0 {
			event.Status = DR.ES_ACTIVE
		}
		return nil
	}
	for _, elem := range round.Pairing {
		score1, _ := strconv.Atoi(strings.Split(elem.Score, ":")[0])
		score2, _ := strconv.Atoi(strings.Split(elem.Score, ":")[1])
		if score1 > score2 {
			for _, standing := range event.Tournament.Standings {
				if standing.TeamName == elem.TeamOne {
					*standing.Points = *standing.Points + 3
				}
			}
		}
		if score1 == score2 {
			for _, standing := range event.Tournament.Standings {
				if standing.TeamName == elem.TeamOne {
					*standing.Points = *standing.Points + 1
				}
				if standing.TeamName == elem.TeamTwo {
					*standing.Points = *standing.Points + 1
				}
			}
		}
		if score1 < score2 {
			for _, standing := range event.Tournament.Standings {
				if standing.TeamName == elem.TeamTwo {
					*standing.Points = *standing.Points + 3
				}
			}
		}
	}
	event.Tournament.Rounds[len(event.Tournament.Rounds)-1] = *round
	sort.Slice(event.Tournament.Standings, func(i, j int) bool {
		return *event.Tournament.Standings[i].Points > *event.Tournament.Standings[j].Points
	})
	for i := range event.Tournament.Standings {
		*event.Tournament.Standings[i].Ranking = i + 1
	}
	return nil
}

// TODO DB
func nextRound(ctx context.Context, Repo *crud.Repo, event *DR.Event) {
	if event.Tournament.Rounds == nil {
		event.Tournament.Rounds = make([]DR.Round, 0)
		event.Tournament.Rounds = append(event.Tournament.Rounds, DR.Round{})
		event.Tournament.Rounds[0].Pairing = make([]DR.Pairing, 0)
		if len(event.Teams)%2 == 0 {
			for i := 0; i < len(event.Teams)/2; i++ {
				event.Tournament.Rounds[0].Pairing = append(event.Tournament.Rounds[0].Pairing, DR.Pairing{
					TeamOne: event.Teams[i].Name,
					TeamTwo: event.Teams[len(event.Teams)-i-1].Name,
				})
			}
		} else {
			event.Tournament.Rounds[0].Pairing = append(event.Tournament.Rounds[0].Pairing, DR.Pairing{
				TeamOne: event.Teams[0].Name,
				TeamTwo: "Bye",
				Score:   "3:0",
			})
			for i := 1; i < (len(event.Teams)+1)/2; i++ {
				event.Tournament.Rounds[0].Pairing = append(event.Tournament.Rounds[0].Pairing, DR.Pairing{
					TeamOne: event.Teams[i].Name,
					TeamTwo: event.Teams[len(event.Teams)-i].Name,
				})
			}
		}
	} else {
		if len(event.Tournament.Rounds) >= len(event.Teams)-1 {
			return
		}
		nextRound := DR.Round{
			Pairing: make([]DR.Pairing, len(event.Tournament.Rounds[len(event.Tournament.Rounds)-1].Pairing)),
		}
		copy(nextRound.Pairing, event.Tournament.Rounds[len(event.Tournament.Rounds)-1].Pairing)
		save := nextRound.Pairing[0].TeamTwo
		for i := 0; i < len(nextRound.Pairing)-1; i++ {
			nextRound.Pairing[i].TeamTwo = nextRound.Pairing[i+1].TeamTwo
			nextRound.Pairing[i].Score = ""
		}
		nextRound.Pairing[len(nextRound.Pairing)-1].TeamTwo = nextRound.Pairing[len(nextRound.Pairing)-1].TeamOne
		for i := len(nextRound.Pairing) - 1; i > 1; i-- {
			nextRound.Pairing[i].TeamOne = nextRound.Pairing[i-1].TeamOne
		}
		nextRound.Pairing[len(nextRound.Pairing)-1].Score = ""
		nextRound.Pairing[1].TeamOne = save
		event.Tournament.Rounds = append(event.Tournament.Rounds, nextRound)
	}
}
