package Route

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/stan.go"
	"io"
	"io/ioutil"
	"net/http"
	"pub/HttpProcessing"
	"pub/Models"
)

const secretKey = "secret_maui"

type NatsConn struct {
	Conn stan.Conn
}

func Router(stanConn NatsConn) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Post("/pub", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var user Models.User
		if errUnmarshalBody := UnmarshalBody(r.Body, &user); errUnmarshalBody != nil {
			HttpProcessing.HttpError(w, errUnmarshalBody, "failed unmarshal body",
				"server error", http.StatusInternalServerError)
			return
		}

		user.SecretKey = secretKey
		userB, errMarshal := json.Marshal(user)
		if errMarshal != nil {
			HttpProcessing.HttpError(w, errMarshal, "failed marshal user",
				"server error", http.StatusInternalServerError)
		}

		if errPub := stanConn.Conn.Publish("db_service", userB); errPub != nil {
			HttpProcessing.HttpError(w, errPub, "posting error",
				"server error", http.StatusInternalServerError)
			return
		}
	})

	return router
}

func UnmarshalBody(r io.Reader, v interface{}) error {
	resp, errResp := ioutil.ReadAll(r)
	if errResp != nil {
		//ErrorPorcessing.HttpError(w, errResp, "failed to get body", "Bad Request", http.StatusBadRequest)
		return fmt.Errorf("server error: %w", errResp)
	}

	if errUnmarshalJson := json.Unmarshal(resp, v); errUnmarshalJson != nil {
		//ErrorPorcessing.HttpError(w, errUnmarshalJson, "failed to get Json in Authorization", "Server Error", http.StatusInternalServerError)
		return fmt.Errorf("server error: %w", errUnmarshalJson)
	}

	return nil
}
