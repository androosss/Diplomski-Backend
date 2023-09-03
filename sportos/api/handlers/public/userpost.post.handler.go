package public

import (
	H "backend/internal/helpers"
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type UserpostPostHandler struct {
	userId   string
	userText string
	formData *multipart.Form
}

func (r UserpostPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r UserpostPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *UserpostPostHandler) Init(httpReq *http.Request) DA.Error {
	err := httpReq.ParseMultipartForm(32 << 20)
	if err != nil {
		return DA.InternalServerError(err)
	}
	r.userId = DA.GetUserIdFromContext(httpReq.Context())
	r.userText = httpReq.Form.Get("userText")
	r.formData = httpReq.MultipartForm
	return nil
}

func (r *UserpostPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.userId == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("User is mandatory")
	}
	if r.userText == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Text is mandatory")
	}
	return nil
}

func (r *UserpostPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	ret := DR.UserPost{
		UserId:   r.userId,
		UserText: r.userText,
	}

	for _, fh := range r.formData.File["image"] {

		file, err := fh.Open()
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
		defer file.Close()

		name, err := Repo.GetImageName(ctx)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}

		extension := H.Last(strings.Split(fh.Filename, "."))
		absPath, _ := filepath.Abs("../../assets/images")
		dst, err := os.Create(absPath + "/" + name + "." + extension)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return nil, DA.InternalServerError(err)
		}

		ret.ImageNames = append(ret.ImageNames, name+"."+extension)
	}

	ret, err := Repo.UserPostsCrud.Create(ctx, ret, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}

	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
