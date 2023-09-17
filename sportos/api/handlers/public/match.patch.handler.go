package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type MatchPatchHandler struct {
	MatchPatchRequest
	Players string
	sport   string
	status  DR.MatchStatus
}

type MatchPatchRequest struct {
	Id     string  `json:"id,omitempty"`
	Player *string `json:"player,omitempty"`
	Result *string `json:"result,omitempty"`
}

func (r MatchPatchHandler) SupportedMethod() string {
	return http.MethodPatch
}

func (r MatchPatchHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *MatchPatchHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.MatchPatchRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *MatchPatchHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if match, err := Repo.MatchCrud.GetById(ctx, r.Id, nil); err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("Match with id " + r.Id + " doesn't exist")
	} else {
		if r.Player != nil && strings.Contains(*match.Players, *r.Player) {
			return DA.ErrorBadRequest().WithMessage("Player is already in that match")
		}
		if r.Result != nil && match.Status != DR.MS_FULL {
			return DA.ErrorBadRequest().WithMessage("Can't submit result for match that isn't full")
		}
		if match.Status == DR.MS_FINISHED {
			return DA.ErrorBadRequest().WithMessage("Can't change match that is over")
		}
		if r.Player != nil {
			r.Players = *match.Players + "," + *r.Player
		} else {
			r.Players = *match.Players
		}
		r.sport = match.Sport
		r.status = match.Status
	}
	return nil
}

func (r *MatchPatchHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	up := DR.MatchUpdateParams{
		Id:      r.Id,
		Players: &r.Players,
		Result:  r.Result,
	}
	sport, err := DR.GetSportByName(r.sport)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	if r.Result != nil {
		fin := DR.MS_FINISHED
		up.Status = &fin
		updateStats(ctx, Repo, r.Id, *r.Result)
	}
	if len(strings.Split(r.Players, ",")) == 2*sport.TeamSize && r.status == DR.MS_CREATED {
		full := DR.MS_FULL
		up.Status = &full
		up.Teams = generateTeams(r.Players)
	}
	ret, err := Repo.MatchCrud.Update(ctx, up, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}

func generateTeams(players string) *DR.StrArr {
	retArr := DR.StrArr{"", ""}
	playersArr := strings.Split(players, ",")
	perm := rand.Perm(len(playersArr))
	for i, elem := range perm {
		if i < len(playersArr)/2 {
			retArr[0] += playersArr[elem] + ","
		} else {
			retArr[1] += playersArr[elem] + ","
		}
	}
	retArr[0] = retArr[0][0 : len(retArr[0])-1]
	retArr[1] = retArr[1][0 : len(retArr[1])-1]
	return &retArr
}

func updateStats(ctx context.Context, Repo *crud.Repo, id, result string) error {
	match, _ := Repo.MatchCrud.GetById(ctx, id, nil)
	pointsFirst, err := strconv.Atoi(strings.Split(result, ":")[0])
	if err != nil {
		return err
	}
	pointsSecond, err := strconv.Atoi(strings.Split(result, ":")[1])
	if err != nil {
		return err
	}
	winner := 1
	if pointsFirst > pointsSecond {
		winner = 0
	}
	for _, id := range strings.Split(match.Teams[winner], ",") {
		player, err := Repo.PlayerCrud.GetById(ctx, id, nil)
		if err != nil {
			return err
		}
		stats := player.Statistics
		if stats == nil {
			stats = DR.StatMap{}
		}
		teamResult := result
		if winner != 0 {
			loseResults := strings.Split(result, ":")
			teamResult = loseResults[1] + ":" + loseResults[0]
		}
		nrOfMatches := float64(len(stats[match.Sport].Matches))
		winRation := stats[match.Sport].WinRatio.Mul(decimal.NewFromFloat(nrOfMatches)).Add(decimal.NewFromFloat(1)).Div(decimal.NewFromFloat(nrOfMatches + 1))
		stats[match.Sport] = DR.Statistics{
			WinRatio: winRation,
			Matches: append(stats[match.Sport].Matches, DR.Statistic{
				Date:    *match.StartTime,
				Score:   teamResult,
				MyTeam:  strings.Split(match.Teams[winner], ","),
				OppTeam: strings.Split(match.Teams[1-winner], ","),
			}),
		}
		up := DR.PlayerUpdateParams{
			Id:         id,
			Statistics: &stats,
		}
		_, err = Repo.PlayerCrud.Update(ctx, up, nil, nil)
		if err != nil {
			return err
		}
	}
	for _, id := range strings.Split(match.Teams[1-winner], ",") {
		player, err := Repo.PlayerCrud.GetById(ctx, id, nil)
		if err != nil {
			return err
		}
		stats := player.Statistics
		if stats == nil {
			stats = DR.StatMap{}
		}
		teamResult := result
		if winner != 1 {
			loseResults := strings.Split(result, ":")
			teamResult = loseResults[1] + ":" + loseResults[0]
		}
		nrOfMatches := float64(len(stats[match.Sport].Matches))
		winRation := stats[match.Sport].WinRatio.Mul(decimal.NewFromFloat(nrOfMatches)).Div(decimal.NewFromFloat(nrOfMatches + 1))
		stats[match.Sport] = DR.Statistics{
			WinRatio: winRation,
			Matches: append(stats[match.Sport].Matches, DR.Statistic{
				Date:    *match.StartTime,
				Score:   teamResult,
				MyTeam:  strings.Split(match.Teams[1-winner], ","),
				OppTeam: strings.Split(match.Teams[winner], ","),
			}),
		}
		up := DR.PlayerUpdateParams{
			Id:         id,
			Statistics: &stats,
		}
		_, err = Repo.PlayerCrud.Update(ctx, up, nil, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
