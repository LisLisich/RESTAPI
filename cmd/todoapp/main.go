package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/LisLisich/RESTAPI/internal/core/config"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_pgx_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/LisLisich/RESTAPI/internal/core/transport/http/middleware"
	core_http_server "github.com/LisLisich/RESTAPI/internal/core/transport/http/server"
	identity_password_provider "github.com/LisLisich/RESTAPI/internal/features/identity/provider/password"
	identity_token_provider "github.com/LisLisich/RESTAPI/internal/features/identity/provider/token"
	identity_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/identity/repository/postgres"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
	identity_transport_http "github.com/LisLisich/RESTAPI/internal/features/identity/transport/http"
	statistics_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/statistics/repository/postgres"
	statistics_service "github.com/LisLisich/RESTAPI/internal/features/statistics/service"
	statistics_transport_http "github.com/LisLisich/RESTAPI/internal/features/statistics/transport/http"
	task_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/tasks/repository/postgres"
	task_service "github.com/LisLisich/RESTAPI/internal/features/tasks/service"
	tasks_transport_http "github.com/LisLisich/RESTAPI/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/users/repository/postgres"
	users_service "github.com/LisLisich/RESTAPI/internal/features/users/service"
	users_transport_http "github.com/LisLisich/RESTAPI/internal/features/users/transport/http"
	web_fs_repository "github.com/LisLisich/RESTAPI/internal/features/web/repository/file_system"
	web_service "github.com/LisLisich/RESTAPI/internal/features/web/service"
	web_transport_http "github.com/LisLisich/RESTAPI/internal/features/web/transport/http"
	"go.uber.org/zap"

	_ "github.com/LisLisich/RESTAPI/docs"
)

// @title 		Golang Todo API
// @version 	1.0
// @description Todo Application REST-API scheme
// @host 		127.0.0.1:5050
// @BasePath 	/api/v1
func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()
	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger: ", err)
		os.Exit(1)
	}
	defer logger.Close()
	logger.Debug("application time zone", zap.Any("zone", time.Local))
	logger.Debug("initializing postgres connection pool")
	pool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)

	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()
	logger.Debug("initializing feature", zap.String("feature", "identity"))
	identityRepository := identity_postgres_repository.NewIdentityRepository(pool)
	passwordHasher := identity_password_provider.NewArgon2idHasher(
		identity_password_provider.DefaultArgon2idConfig(),
	)
	verificationTokenIssuer := identity_token_provider.NewRandomIssuer(32)
	identityService := identity_service.NewIdentityService(
		identityRepository,
		passwordHasher,
		verificationTokenIssuer,
		time.Now,
	)
	identityTransportHTTP := identity_transport_http.NewIdentityHTTPHandler(identityService)

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUserRepository(pool)
	usersService := users_service.NewUserService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))
	tasksRepository := task_postgres_repository.NewTaskRepository(pool)
	tasksService := task_service.NewTasksService(tasksRepository)
	tasksTransportHTTP := tasks_transport_http.NewTasksHTTPHandler(tasksService)

	logger.Debug("initializing feature", zap.String("initializing", "feature"))
	statisticsRepository := statistics_postgres_repository.NewStatisticsRepository(pool)
	statisticsService := statistics_service.NewStatisticsService(statisticsRepository)
	statisticsTransportHTTP := statistics_transport_http.NewStatisticsHTTPHandler(statisticsService)

	logger.Debug("initializing feature", zap.String("feature", "web"))
	webRepository := web_fs_repository.NewWebRepository()
	webService := web_service.NewWebService(webRepository)
	webTransportHTTP := web_transport_http.NewWebHTTPHandler(webService)

	logger.Debug("initializing HTTP server")
	httpConfig := core_http_server.NewConfigMust()
	httpServer := core_http_server.NewHTTPServer(
		httpConfig,
		logger,
		core_http_middleware.CORS(httpConfig.AllowedOrigins),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouterV1.RegisterRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouterV1.RegisterRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouterV1.RegisterRoutes(statisticsTransportHTTP.Routes()...)

	apiVersionRouterV2 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
	apiVersionRouterV2.RegisterRoutes(identityTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(
		apiVersionRouterV1,
		apiVersionRouterV2,
	)
	httpServer.RegisterRoutes(webTransportHTTP.Routes()...)
	httpServer.RegisterSwagger()

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
