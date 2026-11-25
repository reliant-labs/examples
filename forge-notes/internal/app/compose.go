// forge-scaffold (USER-OWNED, RECONCILED): forge writes this file once and then
// keeps only its projection in sync. It is the composition graph's wiring: one
// constructor method per node, holding the value expression for every field of
// that node's Deps. Those expressions are yours — change what fills a field
// here, and forge never touches it again.
//
// What forge reconciles on every `forge generate`:
//   - a node with no new<X> method (a service, worker, package or provider you
//     just added) gets one, each Deps field resolved BY TYPE;
//   - a Deps field a node gained gets its key added; a key whose field the
//     Deps struct no longer has is dropped;
//   - a key two nodes could fill is a generate error until you write it here.
//
// The memoized accessors these feed (g.<X>()), the roots and their lifecycle
// are generated in graph_gen.go. `forge project disown internal/app/compose.go`
// stops the reconciliation.
package app

import (
	"net/http"

	"github.com/reliant-labs/forge/pkg/orm"
	"github.com/reliant-labs/forge/pkg/serverkit"

	"example.com/notes/internal/handlers/notes"
)

// newNotes constructs the notes handler (internal/handlers/notes).
func (g *Graph) newNotes() (*notes.Service, error) {
	return notes.New(notes.Deps{
		Logger: g.Log,
		Config: g.Cfg,
		DB:     g.ORM(),
	})
}

// newDB constructs the DB provider (NewDB, internal/app/providers.go).
func (g *Graph) newDB() (*DB, error) {
	return NewDB(DBDeps{
		Ctx:       g.Ctx,
		ServerCfg: g.ServerCfg,
		Logger:    g.Log,
	})
}

// newORM constructs the ORM provider (NewORM, internal/app/providers.go).
func (g *Graph) newORM() (*orm.Client, error) {
	return NewORM(ORMDeps{
		DB: g.DB(),
	})
}

// newHTTPClient constructs the HTTPClient provider (NewHTTPClient, internal/app/providers.go).
func (g *Graph) newHTTPClient() (*http.Client, error) {
	return NewHTTPClient(HTTPClientDeps{})
}

// newSessions constructs the Sessions provider (NewSessions, internal/app/providers.go).
func (g *Graph) newSessions() (*Sessions, error) {
	return NewSessions(SessionsDeps{
		DB:     g.DB(),
		Config: g.Cfg,
		Logger: g.Log,
	})
}

// newUserAuth constructs the UserAuth provider (NewUserAuth, internal/app/providers.go).
func (g *Graph) newUserAuth() (*UserAuth, error) {
	return NewUserAuth(UserAuthDeps{
		Config:   g.Cfg,
		DB:       g.DB(),
		Enrich:   nil, // optional — no provider
		Sessions: g.Sessions(),
	})
}

// newUserChain constructs the UserChain provider (NewUserChain, internal/app/providers.go).
func (g *Graph) newUserChain() (*UserChain, error) {
	return NewUserChain(UserChainDeps{
		Auth:      g.UserAuth(),
		Config:    g.Cfg,
		ServerCfg: g.ServerCfg,
		Logger:    g.Log,
	})
}

// newAPIListener constructs the APIListener provider (NewAPIListener, internal/app/providers.go).
func (g *Graph) newAPIListener() (*APIListener, error) {
	return NewAPIListener(APIListenerDeps{
		Config:    g.Cfg,
		ServerCfg: g.ServerCfg,
		Logger:    g.Log,
		Chain:     g.UserChain(),
		Routes:    g.UserRoutes(),
		Probes:    g.Probes,
		Sessions:  g.Sessions(),
	})
}

// newPprof constructs the Pprof provider (NewPprof, internal/app/providers.go).
func (g *Graph) newPprof() (*serverkit.Pprof, error) {
	return NewPprof(PprofDeps{
		ServerCfg: g.ServerCfg,
		Logger:    g.Log,
	})
}

// newUserRoutes constructs the user-audience route set (routes_gen.go).
func (g *Graph) newUserRoutes() (*UserRoutes, error) {
	return NewUserRoutes(UserRoutesDeps{
		Notes:   g.Notes(),
		Logger:  g.Log,
		Options: MountOptions{}, // per-service mountkit options on top of the listener's chain — yours
	})
}
