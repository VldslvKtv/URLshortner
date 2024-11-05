package save

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url_shortner/internal/lib/api_field/response"
	"url_shortner/internal/lib/logger/sl"
	"url_shortner/internal/lib/random"
	"url_shortner/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"` // omitempty - если пуст, то в json его не будет
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 8

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http_server.handlers.url.save.New"

		log = log.With( // добавить инфу к логу
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req) // распарсим запрос
		if err != nil {
			log.Error("failed to decode req-body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode req"))

			return // обзательно выйти тк render.JSON не прервет выполнение
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validatorErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validatorErr)) // делаем ответ более читаемым

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength) // TODO: Может сгенерироваться уже существующий alias
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLExist) {
				log.Info("url already exists", slog.String("url", req.URL))

				render.JSON(w, r, resp.Error("url already exists"))

				return
			}

			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, resp.OK())

	}
}
