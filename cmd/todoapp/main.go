package main

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/LisLisich/RESTAPI/internal/core/config"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_pgx_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/LisLisich/RESTAPI/internal/core/transport/http/middleware"
	core_http_server "github.com/LisLisich/RESTAPI/internal/core/transport/http/server"
	identity_google_provider "github.com/LisLisich/RESTAPI/internal/features/identity/provider/googleoidc"
	identity_jwt_provider "github.com/LisLisich/RESTAPI/internal/features/identity/provider/jwt"
	identity_password_provider "github.com/LisLisich/RESTAPI/internal/features/identity/provider/password"
	identity_token_provider "github.com/LisLisich/RESTAPI/internal/features/identity/provider/token"
	identity_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/identity/repository/postgres"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
	identity_transport_http "github.com/LisLisich/RESTAPI/internal/features/identity/transport/http"
	identity_http_middleware "github.com/LisLisich/RESTAPI/internal/features/identity/transport/http/middleware"
	notifications_smtp_provider "github.com/LisLisich/RESTAPI/internal/features/notifications/provider/smtp"
	notifications_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/notifications/repository/postgres"
	notifications_service "github.com/LisLisich/RESTAPI/internal/features/notifications/service"
	notifications_transport_http "github.com/LisLisich/RESTAPI/internal/features/notifications/transport/http"
	payments_yookassa_provider "github.com/LisLisich/RESTAPI/internal/features/payments/provider/yookassa"
	payments_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/payments/repository/postgres"
	payments_service "github.com/LisLisich/RESTAPI/internal/features/payments/service"
	payments_transport_http "github.com/LisLisich/RESTAPI/internal/features/payments/transport/http"
	statistics_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/statistics/repository/postgres"
	statistics_service "github.com/LisLisich/RESTAPI/internal/features/statistics/service"
	statistics_transport_http "github.com/LisLisich/RESTAPI/internal/features/statistics/transport/http"
	task_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/tasks/repository/postgres"
	task_service "github.com/LisLisich/RESTAPI/internal/features/tasks/service"
	tasks_transport_http "github.com/LisLisich/RESTAPI/internal/features/tasks/transport/http"
	tasks_transport_http_v2 "github.com/LisLisich/RESTAPI/internal/features/tasks/transport/http/v2"
	users_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/users/repository/postgres"
	users_service "github.com/LisLisich/RESTAPI/internal/features/users/service"
	users_transport_http "github.com/LisLisich/RESTAPI/internal/features/users/transport/http"
	wallet_postgres_repository "github.com/LisLisich/RESTAPI/internal/features/wallet/repository/postgres"
	wallet_service "github.com/LisLisich/RESTAPI/internal/features/wallet/service"
	wallet_transport_http "github.com/LisLisich/RESTAPI/internal/features/wallet/transport/http"
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
	jwtConfig := identity_jwt_provider.NewConfigMust()
	accessTokenIssuer := identity_jwt_provider.NewEd25519Provider(
		jwtConfig.PrivateKey,
		jwtConfig.PrivateKey.Public().(ed25519.PublicKey),
		"fintask",
		"fintask-cli",
		15*time.Minute,
		time.Now,
	)
	identityService := identity_service.NewIdentityService(
		identityRepository,
		passwordHasher,
		verificationTokenIssuer,
		time.Now,
		identity_service.WithAccessTokenIssuer(accessTokenIssuer),
	)
	authenticationMiddleware := identity_http_middleware.Authentication(
		identityService,
		accessTokenIssuer,
		identity_transport_http.SessionCookieName,
	)
	identityTransportHTTP := identity_transport_http.NewIdentityHTTPHandler(
		identityService,
		authenticationMiddleware,
	)
	googleConfig := identity_google_provider.NewConfigMust()
	if googleConfig.Enabled {
		googleClient, err := identity_google_provider.NewClient(
			ctx,
			googleConfig,
			&http.Client{Timeout: 15 * time.Second},
		)
		if err != nil {
			logger.Fatal("failed to initialize Google OIDC", zap.Error(err))
		}
		identityTransportHTTP.SetGoogleLoginService(
			identity_service.NewGoogleLoginService(
				identityRepository,
				googleClient,
				verificationTokenIssuer,
				time.Now,
			),
		)
	}

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUserRepository(pool)
	usersService := users_service.NewUserService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))
	tasksRepository := task_postgres_repository.NewTaskRepository(pool)
	tasksService := task_service.NewTasksService(tasksRepository)
	tasksTransportHTTP := tasks_transport_http.NewTasksHTTPHandler(tasksService)
	tasksTransportHTTPV2 := tasks_transport_http_v2.NewTasksHTTPHandler(
		tasksService,
		authenticationMiddleware,
	)

	logger.Debug("initializing feature", zap.String("feature", "wallet"))
	walletRepository := wallet_postgres_repository.NewWalletRepository(pool)
	walletService := wallet_service.NewWalletService(walletRepository)
	walletTransportHTTP := wallet_transport_http.NewWalletHTTPHandler(
		walletService,
		authenticationMiddleware,
	)

	logger.Debug("initializing feature", zap.String("feature", "notifications"))
	notificationRepository := notifications_postgres_repository.NewNotificationRepository(pool)
	smtpConfig := notifications_smtp_provider.NewConfigMust()
	var emailSender notifications_service.EmailSender = notifications_smtp_provider.DisabledSender{}
	if smtpConfig.Enabled {
		emailSender = notifications_smtp_provider.NewSender(smtpConfig)
	}
	notificationDispatcher := notifications_service.NewNotificationDispatcher(
		notificationRepository,
		emailSender,
	)
	outboxProcessor := notifications_service.NewProcessor(
		notificationRepository,
		notificationDispatcher,
		5,
		time.Minute,
		time.Now,
	)
	notificationsTransportHTTP := notifications_transport_http.NewNotificationHTTPHandler(
		notificationRepository,
		authenticationMiddleware,
	)
	go runOutboxProcessor(ctx, logger, outboxProcessor)

	yooKassaConfig := payments_yookassa_provider.NewConfigMust()
	var paymentsTransportHTTP *payments_transport_http.PaymentHTTPHandler
	if yooKassaConfig.Enabled {
		logger.Debug("initializing feature", zap.String("feature", "payments"))
		paymentProvider := payments_yookassa_provider.NewClient(
			yooKassaConfig.APIURL,
			yooKassaConfig.ShopID,
			yooKassaConfig.SecretKey,
			&http.Client{Timeout: 15 * time.Second},
		)
		paymentRepository := payments_postgres_repository.NewPaymentRepository(pool)
		paymentService := payments_service.NewPaymentService(
			paymentRepository,
			paymentProvider,
			yooKassaConfig.ReturnURL,
			time.Now,
		)
		paymentsTransportHTTP = payments_transport_http.NewPaymentHTTPHandler(
			paymentService,
			authenticationMiddleware,
		)
	}

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
	apiVersionRouterV2.RegisterRoutes(tasksTransportHTTPV2.Routes()...)
	apiVersionRouterV2.RegisterRoutes(walletTransportHTTP.Routes()...)
	apiVersionRouterV2.RegisterRoutes(notificationsTransportHTTP.Routes()...)
	if paymentsTransportHTTP != nil {
		apiVersionRouterV2.RegisterRoutes(paymentsTransportHTTP.Routes()...)
	}

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

func runOutboxProcessor(
	ctx context.Context,
	logger *core_logger.Logger,
	processor *notifications_service.Processor,
) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		processed, err := processor.ProcessOne(ctx)
		if err != nil && ctx.Err() == nil {
			logger.Warn("outbox event processing failed", zap.Error(err))
		}
		if processed && err == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
