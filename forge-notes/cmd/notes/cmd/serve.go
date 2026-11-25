// yours: scaffolded once, never touched again — forge will not overwrite this file
//
// THE PROCESS POLICY — the one every root runs under, and nothing else.
//
// What a process RUNS is one root of the composition graph
// (internal/app/graph_gen.go): `notes server` runs the server root — the
// API listener serving every user-audience service, plus every worker and
// operator — and every //forge:root you declare in internal/app gets a
// command of its own (roots_gen.go, generated). Each node is constructed
// once, from its own Deps, in internal/app/compose.go; what it is CONFIGURED
// with is the providers in internal/app/providers.go.
//
// runRoot is the policy around whichever root a command names: load and
// validate config, build the logger, hold the port with boot probes while
// dependencies come up, migrate (when the root reaches the database), install
// telemetry, then run the root until a signal — readiness flips to draining,
// the pre-stop pause runs, and every node drains and closes in reverse
// dependency order inside the shutdown budget, telemetry flushed last. Those
// are yours; change them here, for every root at once — or branch on
// root.Name for one.
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" sql driver for DATABASE_URL
	"github.com/reliant-labs/forge/pkg/lifecyclekit"
	"github.com/reliant-labs/forge/pkg/migratekit"
	"github.com/reliant-labs/forge/pkg/serverkit"

	forgedb "example.com/notes/db"
	"example.com/notes/internal/app"
	"example.com/notes/pkg/config"
)

// runRoot runs one root (roots_gen.go hands it the root a command names):
// every node it reaches, supervised until a signal or a failing node.
func runRoot(ctx context.Context, deps Deps, root rootSpec) error {
	// Load validates: a value the config's annotations reject fails here.
	// Only the fields the root reaches are demanded (root.Config, from
	// roots_gen.go): a worker that never opens the database starts without
	// DATABASE_URL, and its workload is never handed the credential.
	cfg, err := config.LoadReach(deps.Cmd, root.Config)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	skCfg, err := serverkitConfig(cfg)
	if err != nil {
		return err
	}

	logger := serverkit.NewLogger(skCfg)
	slog.SetDefault(logger)
	if skCfg.IsDevelopment() {
		logger.Warn("running in development mode — permissive edge defaults are enabled (CORS reflects any origin, HSTS off). NEVER set ENVIRONMENT=development in production.")
	}

	// SIGINT/SIGTERM begin the stop. The same context is the graph's boot
	// context, so a signal during boot also abandons a dependency still
	// being waited on.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// BOOT PROBES. Bind the API port NOW, before any dependency is touched:
	// until the API listener takes the address over, /healthz answers 200 and
	// /readyz 503 naming the dependency being waited on (the database connect
	// retries for up to DB_CONNECT_TIMEOUT instead of failing the process).
	ctx, bootProbes, err := serverkit.StartBootProbes(ctx, skCfg, logger)
	if err != nil {
		return err
	}
	defer bootProbes.Close()

	// Migrate BEFORE building the graph: a service that starts serving
	// against a schema its migrations have not reached answers requests
	// with errors that look like bugs. Only a root that reaches the database
	// migrates — a worker root that never opens it never runs a migration —
	// unless its //forge:root says otherwise: migrate=wait leaves the
	// migrations to another process and blocks (/readyz naming "schema")
	// until the schema is current; migrate=none leaves it alone.
	if root.Migrate {
		if merr := serverkit.AutoMigrateContext(ctx, skCfg, logger,
			migratekit.AutoMigrate(forgedb.Source(), forgedb.SourceDir)); merr != nil {
			return merr
		}
	} else if root.AwaitMigrations {
		if werr := serverkit.AwaitMigrationsContext(ctx, skCfg, logger,
			migratekit.AwaitMigrations(forgedb.Source(), forgedb.SourceDir, migratekit.AwaitOptions{})); werr != nil {
			return werr
		}
	}

	// TELEMETRY before the graph, so every instrumented client is created
	// against the installed providers; flushed LAST on the way down, so the
	// shutdown's own spans and metrics are exported.
	otelShutdown, metrics, oerr := serverkit.SetupTelemetry(ctx, skCfg)
	if oerr != nil {
		logger.Error("failed to initialize OpenTelemetry", "error", oerr)
	}

	// THE ROOT POLICY: the readiness gate and stop budget (PRE_STOP_DELAY,
	// SHUTDOWN_TIMEOUT), and the probes — served in front of the API
	// listener's edge, or, for a root without one, by a probe listener of its
	// own on the same address. Every node that reports Ready joins /readyz.
	drainer := serverkit.NewDrainer(skCfg)
	drainer.Logger = logger
	g := &app.Graph{
		Ctx:       ctx,
		Cfg:       cfg,
		Log:       logger,
		ServerCfg: skCfg,
		Probes:    &serverkit.Probes{Drainer: drainer, Metrics: metrics, Timeout: skCfg.ReadinessTimeout, Logger: logger, Paths: skCfg.Probes},
	}
	logger.Info("running root", "root", root.Name)
	return root.Run(g, ctx, lifecyclekit.Policy{
		Drainer:       drainer,
		Logger:        logger,
		FailurePolicy: skCfg.FailurePolicy,
		Flush:         otelShutdown,
	})
}

// serverkitConfig projects this project's typed *config.Config onto the
// vendor-neutral serverkit.Config, so serverkit needs no compile-time
// dependency on pkg/config.
func serverkitConfig(cfg *config.Config) (serverkit.Config, error) {
	skCfg := serverkit.Config{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		// ServiceName is APP IDENTITY (the project name baked into a
		// generated constant), NOT a config field. serverkit reports it on
		// OTel traces/metrics (semconv service.name).
		ServiceName:  ServiceName,
		OTLPEndpoint: cfg.OtlpEndpoint,
		// pprof, ON BY DEFAULT, on its own listener — never the app port.
		// Defaults to 127.0.0.1:6060 (proto/<project>/config/v1/config.proto), so it is
		// live in every environment and routable from none: no k8s Service,
		// no route, no published container port. Grab a profile with
		//
		//	go tool pprof http://localhost:6060/debug/pprof/heap
		//	kubectl port-forward <pod> 6060:6060   # then the line above
		//
		// Override with PPROF_ADDR / --pprof-addr; --pprof-addr= turns it off
		// (an empty PPROF_ADDR is unset, so it keeps the default). A failed
		// bind is logged and skipped, never fatal.
		PprofAddr:            cfg.PprofAddr,
		TLSCertPath:          cfg.TlsCertPath,
		TLSKeyPath:           cfg.TlsKeyPath,
		LogFormat:            cfg.LogFormat,
		Environment:          cfg.Environment,
		AutoMigrate:          cfg.AutoMigrate,
		DatabaseURL:          cfg.DatabaseUrl,
		ReadMaxBytes:         int(cfg.ReadMaxBytes),
		SendMaxBytes:         int(cfg.SendMaxBytes),
		SecurityHeaders:      cfg.SecurityHeadersEnabled,
		CORSAllowCredentials: cfg.CorsAllowCredentials,
	}

	if uerr := skCfg.LogLevel.UnmarshalText([]byte(cfg.LogLevel)); uerr != nil {
		return serverkit.Config{}, fmt.Errorf("invalid log level %q: %w", cfg.LogLevel, uerr)
	}

	if cfg.CorsOrigins != "" {
		for _, o := range strings.Split(cfg.CorsOrigins, ",") {
			if o = strings.TrimSpace(o); o != "" {
				skCfg.CORSOrigins = append(skCfg.CORSOrigins, o)
			}
		}
	}

	// The probe paths (empty is /healthz, /readyz). The KCL lowering reads
	// the same LIVENESS_PATHS / READINESS_PATHS for the probe it writes, so
	// the pod is probed where this process answers.
	skCfg.Probes.Live = serverkit.ProbePathList(cfg.LivenessPaths)
	skCfg.Probes.Ready = serverkit.ProbePathList(cfg.ReadinessPaths)

	skCfg.PreStopDelay = cfg.PreStopDelay.AsDuration()
	skCfg.ShutdownTimeout = cfg.ShutdownTimeout.AsDuration()

	skCfg.DBPoolTuning.MaxOpenConns = int(cfg.DbMaxOpenConns)
	skCfg.DBPoolTuning.MaxIdleConns = int(cfg.DbMaxIdleConns)
	skCfg.DBPoolTuning.ConnMaxIdleTime = cfg.DbConnMaxIdleTime.AsDuration()
	skCfg.DBPoolTuning.ConnMaxLifetime = cfg.DbConnMaxLifetime.AsDuration()
	skCfg.DBConnectTimeout = cfg.DbConnectTimeout.AsDuration()

	// Normalize LAST, so nothing downstream can observe a zero-valued
	// cap/timeout/driver. runRoot reads several of these fields — the Connect
	// payload caps, the auto-migrate sql.Open driver — BEFORE serverkit.Run
	// is entered, and Run's own normalization is far too late to help those
	// reads: a zero ReadMaxBytes/SendMaxBytes already reached connect as
	// UNLIMITED, and a zero DBDriver already reached sql.Open as `unknown
	// driver ""`. Normalizing here makes that whole class of ordering bug
	// unreachable.
	skCfg.Normalize()

	return skCfg, nil
}
