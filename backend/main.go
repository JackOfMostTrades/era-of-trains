package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	_ "github.com/go-sql-driver/mysql"
	"log/slog"
	_ "modernc.org/sqlite"
	"net"
	"net/http"
	"net/http/cgi"
	"os"
	"os/signal"
	"sync"
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

	stmt, err := server.db.Prepare("SELECT id,nickname FROM users WHERE id=?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(sessionData.UserId)
	user := new(User)
	err = row.Scan(&user.Id, &user.Nickname)
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
	User *User `json:"user"`
}

func whoami(ctx *RequestContext, req *WhoAmIRequest) (resp *WhoAmIResponse, err error) {
	return &WhoAmIResponse{
		User: ctx.User,
	}, nil
}

func main() {
	configPath := "config-local.json"
	if envConfigPath := os.Getenv("EOT_CONFIG_PATH"); envConfigPath != "" {
		configPath = envConfigPath
	}
	config := new(Config)
	if f, err := os.Open(configPath); err != nil {
		panic(err)
	} else {
		err = json.NewDecoder(f).Decode(config)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
	if config.WorkingDirectory != "" {
		err := os.Chdir(config.WorkingDirectory)
		if err != nil {
			panic(fmt.Errorf("failed to chdir to %s: %v", config.WorkingDirectory, err))
		}
	}
	if config.Database == nil {
		panic(fmt.Errorf("database must be configured"))
	}

	var db *sql.DB
	if config.Database.SqlitePath != "" {
		var err error
		db, err = sql.Open("sqlite", config.Database.SqlitePath)
		if err != nil {
			panic(fmt.Errorf("failed to open sqlite db: %v", err))
		}
	} else if config.Database.MysqlHostname != "" {
		var err error
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s",
			config.Database.MysqlUsername, config.Database.MysqlPassword,
			config.Database.MysqlHostname, config.Database.MysqlDatabase))
		if err != nil {
			panic(fmt.Errorf("failed to open db: %v", err))
		}
	} else {
		panic(fmt.Errorf("no database configured"))
	}

	if config.Database.Bootstrap {
		bootstrapSql, err := os.ReadFile("bootstrap.sql")
		if err != nil {
			panic(fmt.Errorf("failed to load bootstrap.sql: %v", err))
		}
		_, err = db.Exec(string(bootstrapSql))
		if err != nil {
			panic(fmt.Errorf("failed to run bootstrap.sql: %v", err))
		}
	}

	gameMaps, err := maps.LoadMaps()
	if err != nil {
		panic(fmt.Errorf("failed to load maps: %v", err))
	}

	server := &GameServer{
		config:       config,
		db:           db,
		gameMaps:     gameMaps,
		randProvider: &common.CryptoRandProvider{},
	}
	defer server.Close()

	if len(os.Args) >= 2 && os.Args[1] == "run-task" {
		err = runTask(server, os.Args[2:])
		if err != nil {
			panic(fmt.Errorf("failed to run task: %v", err))
		}
	} else {
		err = server.runHttpServer()
		if err != nil {
			panic(fmt.Errorf("failed to run http server: %v", err))
		}
		slog.Info("Started HTTP server", "port", server.httpListenPort)

		// Wait for control-c
		wg := sync.WaitGroup{}
		wg.Add(1)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			select {
			case <-c:
				wg.Done()
			}
		}()
		wg.Wait()

		slog.Info("Begin graceful shutdown...")
		err = server.stopHttpServer()
		if err != nil {
			slog.Error("Error stopping http server", "error", err)
		}
	}
}

func (server *GameServer) runHttpServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", jsonHandlerUnAuthenticated(server.login))
	mux.HandleFunc("/api/register", jsonHandlerUnAuthenticated(server.register))
	mux.HandleFunc("/api/linkProfile", jsonHandler(server, server.linkProfile))
	mux.HandleFunc("/api/logout", jsonHandlerUnAuthenticated(server.logout))
	mux.HandleFunc("/api/whoami", jsonHandler(server, whoami))
	mux.HandleFunc("/api/createGame", jsonHandler(server, server.createGame))
	mux.HandleFunc("/api/joinGame", jsonHandler(server, server.joinGame))
	mux.HandleFunc("/api/leaveGame", jsonHandler(server, server.leaveGame))
	mux.HandleFunc("/api/startGame", jsonHandler(server, server.startGame))
	mux.HandleFunc("/api/listGames", jsonHandler(server, server.listGames))
	mux.HandleFunc("/api/confirmMove", jsonHandler(server, server.confirmMove))
	mux.HandleFunc("/api/viewGame", jsonHandler(server, server.viewGame))
	mux.HandleFunc("/api/getGameLogs", jsonHandler(server, server.getGameLogs))
	mux.HandleFunc("/api/getMyGames", jsonHandler(server, server.getMyGames))
	mux.HandleFunc("/api/getMyProfile", jsonHandler(server, server.getMyProfile))
	mux.HandleFunc("/api/setMyProfile", jsonHandler(server, server.setMyProfile))
	mux.HandleFunc("/api/getGameChat", jsonHandler(server, server.getGameChat))
	mux.HandleFunc("/api/sendGameChat", jsonHandler(server, server.sendGameChat))
	mux.HandleFunc("/api/pollGameStatus", jsonHandler(server, server.pollGameStatus))

	var err error
	if server.config.CgiMode {
		err = cgi.Serve(mux)
	} else {
		if server.httpServer != nil {
			return fmt.Errorf("http server is already running")
		}
		listenPort := server.config.HttpListenPort
		if listenPort == 0 {
			listenPort = 8080
		} else if listenPort < 0 {
			listenPort = 0
		}
		listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", listenPort))
		if err != nil {
			return err
		}
		server.httpListenPort = listener.Addr().(*net.TCPAddr).Port
		server.httpServer = &http.Server{Addr: "localhost:8080", Handler: mux}

		go server.httpServer.Serve(listener)
	}
	return err
}

func (server *GameServer) stopHttpServer() error {
	if server.httpServer != nil {
		err := server.httpServer.Shutdown(context.Background())
		server.httpServer = nil
		return err
	}
	return nil
}
