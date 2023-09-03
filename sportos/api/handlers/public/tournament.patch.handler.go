package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type TournamentPatchHandler struct {
	TournamentPatchRequest
}

type TournamentPatchRequest struct {
	Id     string    `json:"id"`
	Team   *string   `json:"team,omitempty"`
	Cancel *bool     `json:"cancel"`
	Finish *bool     `json:"finish"`
	Round  *DR.Round `json:"round"`
}

func (r TournamentPatchRequest) SupportedMethod() string {
	return http.MethodPatch
}

func (r TournamentPatchRequest) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *TournamentPatchHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.TournamentPatchRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *TournamentPatchHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	event, err := Repo.EventCrud.GetById(ctx, r.Id, nil)
	if err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_UNIQUE_CONSTRAINT).WithMessage("Tournament doesn't exist")
	}
	if r.Team != nil {
		for _, team := range event.Teams {
			if team.TeamId == *r.Team {
				return DA.ErrorBadRequest().WithMessage("Team is already applied for event")
			}
		}
		_, err = Repo.TeamCrud.GetById(ctx, *r.Team, nil)
		if err != nil {
			return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_UNIQUE_CONSTRAINT).WithMessage("Team doesn't exist")
		}
	}
	return nil
}

func (r *TournamentPatchHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	tx, err := Repo.DB.Begin()
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	defer tx.Rollback()
	event, err := Repo.EventCrud.GetById(ctx, r.Id, tx)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	up := DR.EventUpdateParams{
		Id: event.EventId,
	}
	if r.Team != nil {
		team, err := Repo.TeamCrud.GetById(ctx, *r.Team, tx)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
		teams := event.Teams
		teams = append(teams, DR.TeamRef{TeamId: *r.Team, Name: team.Name})
		up.Teams = &teams
	}
	if r.Cancel != nil && *r.Cancel {
		cancelled := DR.ES_CANCELLED
		up.Status = &cancelled
	}
	if r.Finish != nil && *r.Finish {
		finished := DR.ES_FINISHED
		up.Status = &finished
		if r.Round != nil {
			updateTournament(ctx, Repo, &event, r.Round)
			up.Tournament = event.Tournament
		}
		for _, teamRef := range event.Teams {
			team, _ := Repo.TeamCrud.GetById(ctx, teamRef.TeamId, tx)
			for _, standing := range event.Tournament.Standings {
				if standing.TeamName == teamRef.Name {
					for _, playerId := range strings.Split(team.Players, ",") {
						player, _ := Repo.PlayerCrud.GetById(ctx, playerId, tx)
						if player.Statistics == nil {
							player.Statistics = DR.StatMap{}
						}
						stat := player.Statistics[event.Sport]
						stat.Tournaments = append(stat.Tournaments, DR.TournamentFinish{
							MyTeam:     strings.Split(team.Players, ","),
							Tournament: event.Name,
							Ranking:    *standing.Ranking,
						})
						player.Statistics[event.Sport] = stat
						up := DR.PlayerUpdateParams{
							Id:         player.Username,
							Statistics: &player.Statistics,
						}
						Repo.PlayerCrud.Update(ctx, up, tx, nil)
					}
				}
			}
		}
	}
	ret, err := Repo.EventCrud.Update(ctx, up, tx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	tx.Commit()
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
