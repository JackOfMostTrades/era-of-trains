package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
)

type RequestContext struct {
	HttpRequest  *http.Request
	HttpResponse http.ResponseWriter
	User         *User
}

type HttpError struct {
	description string
	code        int
}

func (e *HttpError) Error() string {
	return e.description
}

func (server *GameServer) assertAuthentication(ctx *RequestContext) error {
	cookie, err := ctx.HttpRequest.Cookie("eot-session")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return &HttpError{description: "Unauthenticated", code: http.StatusUnauthorized}
		}
		return &HttpError{description: "Failed to parse cookies", code: http.StatusInternalServerError}
	}
	session := cookie.Value
	sessionData, err := server.validateSession(session)
	if err != nil {
		return &HttpError{description: fmt.Sprintf("Failed to parse session: %v", err), code: http.StatusBadRequest}
	}

	stmt, err := server.db.Prepare("SELECT id,nickname,email FROM users WHERE id=?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(sessionData.UserId)
	user := new(User)
	err = row.Scan(&user.Id, &user.Nickname, &user.Email)
	if err != nil {
		return fmt.Errorf("failed to resolve user %s: %v", sessionData.UserId, err)
	}

	ctx.User = user
	return nil
}

func handleJsonRequest[ReqT any, ResT any](ctx *RequestContext, handler func(ctx *RequestContext, r *ReqT) (s *ResT, err error)) {
	if ctx.HttpRequest.Header.Get("Content-Type") != "application/json" {
		http.Error(ctx.HttpResponse, "Bad Content-Type", http.StatusBadRequest)
		return
	}
	if ctx.HttpRequest.Body == nil {
		http.Error(ctx.HttpResponse, "Missing content body", http.StatusBadRequest)
		return
	}

	req := new(ReqT)
	err := json.NewDecoder(ctx.HttpRequest.Body).Decode(req)
	if err != nil {
		http.Error(ctx.HttpResponse, err.Error(), http.StatusBadRequest)
		return
	}
	s, err := handler(ctx, req)
	if err != nil {
		if httpErr, ok := err.(*HttpError); ok {
			http.Error(ctx.HttpResponse, httpErr.Error(), httpErr.code)
			return
		} else {
			http.Error(ctx.HttpResponse, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	ctx.HttpResponse.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(ctx.HttpResponse).Encode(s)
}

func jsonHandlerUnAuthenticated[ReqT any, ResT any](handler func(ctx *RequestContext, r *ReqT) (s *ResT, err error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := &RequestContext{
			HttpRequest:  r,
			HttpResponse: w,
		}
		handleJsonRequest(ctx, handler)
	}
}

func jsonHandler[ReqT any, ResT any](server *GameServer, handler func(ctx *RequestContext, r *ReqT) (s *ResT, err error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := &RequestContext{
			HttpRequest:  r,
			HttpResponse: w,
		}
		err := server.assertAuthentication(ctx)
		if err != nil {
			if httpError, ok := err.(*HttpError); ok {
				http.Error(w, httpError.Error(), httpError.code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		handleJsonRequest(ctx, handler)
	}
}

type WhoAmIRequest struct {
}

type WhoAmIResponse struct {
	User string
}

func whoami(ctx *RequestContext, req *WhoAmIRequest) (resp *WhoAmIResponse, err error) {
	return &WhoAmIResponse{
		User: ctx.User.Nickname,
	}, nil
}

func main() {
	config := new(Config)
	if f, err := os.Open("config-local.json"); err != nil {
		panic(err)
	} else {
		err = json.NewDecoder(f).Decode(config)
		if err != nil {
			panic(err)
		}
		f.Close()
	}

	db, err := sql.Open("sqlite", "file:eot.db")
	if err != nil {
		panic(fmt.Errorf("failed to open sqlite db: %v", err))
	}
	bootstrapSql, err := os.ReadFile("bootstrap.sql")
	if err != nil {
		panic(fmt.Errorf("failed to load bootstrap.sql: %v", err))
	}
	_, err = db.Exec(string(bootstrapSql))
	if err != nil {
		panic(fmt.Errorf("failed to run bootstrap.sql: %v", err))
	}

	maps, err := loadMaps()
	if err != nil {
		panic(fmt.Errorf("failed to load maps: %v", err))
	}

	server := &GameServer{
		config: config,
		db:     db,
		maps:   maps,
	}
	defer server.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", jsonHandlerUnAuthenticated(server.login))
	mux.HandleFunc("/api/logout", jsonHandlerUnAuthenticated(server.logout))
	mux.HandleFunc("/api/whoami", jsonHandler(server, whoami))
	mux.HandleFunc("/api/createGame", jsonHandler(server, server.createGame))
	mux.HandleFunc("/api/joinGame", jsonHandler(server, server.joinGame))
	mux.HandleFunc("/api/leaveGame", jsonHandler(server, server.leaveGame))
	mux.HandleFunc("/api/startGame", jsonHandler(server, server.startGame))
	mux.HandleFunc("/api/listGames", jsonHandler(server, server.listGames))
	mux.HandleFunc("/api/confirmMove", jsonHandler(server, server.confirmMove))
	mux.HandleFunc("/api/viewGame", jsonHandler(server, server.viewGame))

	err = http.ListenAndServe("localhost:8080", mux)
	if err != http.ErrServerClosed {
		panic(err)
	}
}
