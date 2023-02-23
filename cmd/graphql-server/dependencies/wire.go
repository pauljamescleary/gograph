//go:build wireinject
// +build wireinject

package dependencies

import (
	"github.com/google/wire"
	"github.com/pauljamescleary/gograph/cmd/graphql-server/graph"
	"github.com/pauljamescleary/gograph/pkg/common/db"
	"github.com/pauljamescleary/gograph/pkg/services/notes"
)

var postgresDbConnectionSet = wire.NewSet(db.ProvidePgConnectionPool,
	db.ProvideNewPostgresTransactor, db.ProvideNewDatabaseConnection)

var notesApiService = wire.NewSet(notes.ProvideNewNotesRepository, notes.ProvideNewNotesService)

var graphQLServerDependencySet = wire.NewSet(
	postgresDbConnectionSet,
	notesApiService,
	graph.ProvideNewServerResolver,
)

func NewAppResolverService() (*graph.Resolver, error) {
	wire.Build(graphQLServerDependencySet)
	return nil, nil
}
