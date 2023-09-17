package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"fmt"
	"net/http"
)

type StatisticsGetHandler struct {
	userId string
	sport  string
}

func (r StatisticsGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r StatisticsGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *StatisticsGetHandler) Init(httpReq *http.Request) DA.Error {
	r.userId = DA.GetUserIdFromContext(httpReq.Context())
	r.sport = *DA.GetParameterFromURLQuery(httpReq, "sport")
	return nil
}

func (r *StatisticsGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *StatisticsGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	player, err := Repo.PlayerCrud.GetById(ctx, r.userId, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	ret := DA.Statistics{}
	if stat, ok := player.Statistics[r.sport]; ok {
		ret = DA.Statistics{
			WinRatio: fmt.Sprintf("%.2f%%", stat.WinRatio.InexactFloat64()*100),
		}
		for _, tournament := range stat.Tournaments {
			singleApiTournament := DA.Tournament{
				Tournament: tournament.Tournament,
				Ranking:    tournament.Ranking,
			}
			for _, playerId := range tournament.MyTeam {
				player, _ := Repo.PlayerCrud.GetById(ctx, playerId, nil)
				singleApiTournament.MyTeam = append(singleApiTournament.MyTeam, player.Name)
			}
			ret.Tournaments = append(ret.Tournaments, singleApiTournament)
		}
		for _, singleStat := range stat.Matches {
			singleApiStat := DA.Statistic{
				Date:  singleStat.Date,
				Score: singleStat.Score,
			}
			for _, playerId := range singleStat.MyTeam {
				player, _ := Repo.PlayerCrud.GetById(ctx, playerId, nil)
				singleApiStat.MyTeam = append(singleApiStat.MyTeam, player.Name)
			}
			for _, playerId := range singleStat.OppTeam {
				player, _ := Repo.PlayerCrud.GetById(ctx, playerId, nil)
				singleApiStat.OppTeam = append(singleApiStat.OppTeam, player.Name)
			}
			ret.Matches = append(ret.Matches, singleApiStat)
		}
	}
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
