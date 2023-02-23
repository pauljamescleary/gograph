package main

import (
	"fmt"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pauljamescleary/gograph/cmd/graphql-server/graph"
	"github.com/pauljamescleary/gograph/cmd/graphql-server/graph/generated"
	"github.com/pauljamescleary/gograph/pkg/common/config"
	"github.com/pauljamescleary/gograph/pkg/common/db"
	"github.com/pauljamescleary/gograph/pkg/services/notes"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const defaultPort = "8080"

func main() {
	port := GetPortFromEnv()
	// get config file from go args -configpath
	configPath := config.MustGetConfigPathFromFlags("configpath")
	srv, err := SetupServer(configPath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to start server")
		return
	}
	err = srv.Start(fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to start server on port %s", port)
	}
}

const allowedOriginsConfigKey = "server.cors.allowed_origins"

// Defining the Graphql handler
func NewGraphqlHandler(resolver *graph.Resolver) echo.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	c := GetConfigForGrqphQLServer(resolver)
	h := handler.NewDefaultServer(generated.NewExecutableSchema(c))

	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func GetPortFromEnv() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return port
}

func GetConfigForGrqphQLServer(resolver *graph.Resolver) generated.Config {
	c := generated.Config{Resolvers: resolver}
	return c
}

// Defining the Playground handler
func playgroundHandler() echo.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func setCORSMiddleware(r *echo.Echo) {
	allowedOrigins := config.MustGetStringSet(allowedOriginsConfigKey)
	r.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowMethods: []string{"PUT", "POST", "GET", "DELETE"},
		AllowHeaders: []string{"*"},
		MaxAge:       int((12 * time.Hour).Seconds()),
		AllowOriginFunc: func(origin string) (bool, error) {
			return lo.Contains(allowedOrigins, origin), nil
		},
	}))

}

func NewAppResolverService() (*graph.Resolver, error) {
	pool, err := db.ProvidePgConnectionPool()
	if err != nil {
		return nil, err
	}
	database := db.ProvideNewDatabaseConnection(pool)
	repository, err := notes.ProvideNewNotesRepository(database)
	if err != nil {
		return nil, err
	}
	transactionManager, err := db.ProvideNewPostgresTransactor(database)
	if err != nil {
		return nil, err
	}
	service, err := notes.ProvideNewNotesService(repository, transactionManager)
	if err != nil {
		return nil, err
	}
	resolver := graph.ProvideNewServerResolver(service)
	return resolver, nil
}

func SetupServer(configPath string) (*echo.Echo, error) {
	err := config.MustLoadConfigAtPath(configPath)
	r := echo.New()

	if err != nil {
		return nil, err
	}
	resolver, err := NewAppResolverService()
	if err != nil {
		return nil, err
	}
	// Add cors middleware.
	setCORSMiddleware(r)

	r.POST("/query", NewGraphqlHandler(resolver))
	// TODO only in dev mode.
	r.GET("/", playgroundHandler())

	return r, nil
}
