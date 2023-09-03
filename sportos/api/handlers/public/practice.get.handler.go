package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type PracticeGetHandler struct {
	PlayerId *string  `json:"playerId,omitempty"`
	CoachId  *string  `json:"coachId,omitempty"`
	Status   *string  `json:"status,omitempty"`
	Sports   []string `json:"sports"`
}

func (r PracticeGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r PracticeGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *PracticeGetHandler) Init(httpReq *http.Request) DA.Error {
	r.PlayerId = DA.GetParameterFromURLQuery(httpReq, "playerId")
	r.Sports = DA.ParseCommaSeparated(DA.GetParameterFromURLQuery(httpReq, "sports"))
	return nil
}

func (r *PracticeGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *PracticeGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sp := DR.PracticeSearchParams{
		PlayerId: r.PlayerId,
		Sports:   r.Sports,
		CoachId:  r.CoachId,
		Status:   r.Status,
		PracticeSortParams: DR.PracticeSortParams{
			EditInfoCUDSortParams: DR.EditInfoCUDSortParams{
				EditInfoCSortParams: DR.EditInfoCSortParams{
					CreatedAt: &DR.SortColumn{
						Prefix:    "pr",
						Column:    "created_at",
						Order:     0,
						Direction: 0,
					},
				},
			},
		},
	}
	practices, err := Repo.PracticeCrud.Search(ctx, sp, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	practices = r.formatPractices(ctx, Repo, practices)
	resMap := make(map[string]interface{})
	resMap["body"] = practices
	return resMap, nil
}

func (r *PracticeGetHandler) formatPractices(ctx context.Context, Repo *crud.Repo, practices []DR.Practice) []DR.Practice {
	for i := range practices {
		player, _ := Repo.PlayerCrud.GetById(ctx, practices[i].PlayerId, nil)
		practices[i].PlayerId = player.Name
		coach, _ := Repo.CoachCrud.GetById(ctx, practices[i].CoachId, nil)
		practices[i].CoachId = coach.Name
	}
	return practices
}
