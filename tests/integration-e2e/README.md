# Combined platform-api + gateway end-to-end integration tests

This stack runs the **real platform-api (control plane)** and the **real gateway
(gateway-controller + gateway-runtime data plane)** against the **same database
engine**, so a single scenario exercises both products integrated end to end:
an API created in platform-api is deployed to a gateway and served by the data
plane.

It complements the per-component cross-database suites:
- `platform-api/it` — platform-api store on SQLite / PostgreSQL / SQL Server.
- `gateway/it` — gateway store on SQLite / PostgreSQL / SQL Server.

## Topology

```
        REST (9243)                 control-plane WS (9243)
client ───────────────► platform-api ◄────────────────── gateway-controller ──xDS──► gateway-runtime ──► sample-backend
                            │                                  │                          (8080 ingress)
                         platform_api DB                   gateway_test DB
                            └──────────── shared engine (postgres/sqlserver) ─────────────┘
```

platform-api and gateway-controller keep **separate databases** (their schemas
share table names like `artifacts`/`gateways`/`subscriptions`). The PostgreSQL
variant uses one server with two databases (`init-db.sql`).

## Integration contract (verified from source)

- gateway-controller dials platform-api at `…/api/internal/v1/ws/gateways/connect`
  with an `api-key` header (the gateway **registration token**); platform-api
  validates it via `GatewayService.VerifyToken` and replies `connection.ack`.
- platform-api pushes `subscription.created` / deployment events down that socket;
  the gateway-controller pulls subscription/plan data from
  `…/api/internal/v1/subscription-plans` and posts its manifest to
  `…/api/internal/v1/gateways/{id}/manifest`.

## Cucumber suite (godog)

The suite is written in Gherkin and run by [godog](https://github.com/cucumber/godog),
consistent with `gateway/it`:

| File | Purpose |
|------|---------|
| `features/api-deployment.feature` | The REST-API deployment scenarios (Gherkin). |
| `features/llm-deployment.feature` | The LLM-proxy deployment scenario (`@llm`). |
| `suite_test.go` | godog runner + `BeforeSuite`/`AfterSuite` (compose orchestration, bootstrap) + platform-api REST helpers. |
| `steps_test.go` | step definitions + ingress polling. |
| `mock-openai/openapi.yaml` | minimal OpenAI spec served by Prism (`mock-openapi`) as the LLM upstream. |
| `docker-compose*.yaml` | the stack per database engine. |

### How the harness works

`BeforeSuite` brings the whole stack up once and solves the registration-token
chicken-and-egg: it starts the control plane, authenticates (`admin`/`admin`),
creates a project and one (or two) gateway(s) with their registration tokens,
then starts the gateway controllers with those tokens. Each scenario then creates
its own API and deploys it to a pre-registered gateway, so scenarios are
independent. The `I deploy the API to the gateway` step bounces that gateway's
controller, because the controller runs its full deployment sync only once on
connect (`c.syncOnce` in `pkg/controlplane/client.go`); the restart re-runs that
sync so the new deployment is picked up.

## Running

Build the component images once (tagged `it-e2e`), then run the suite:

```bash
cd platform-api && docker build -t platform-api:it-e2e --build-context common=../common .
cd gateway      && make build VERSION=it-e2e   # gateway-controller / gateway-runtime :it-e2e

cd tests/integration-e2e
go test -run TestFeatures -v ./...                                  # PostgreSQL (default)
E2E_DB=sqlite go test -run TestFeatures -v ./...                    # SQLite
MSSQL_IMAGE=mcr.microsoft.com/azure-sql-edge:latest E2E_DB=sqlserver \
  go test -run TestFeatures -v ./...                                # SQL Server (azure-sql-edge on Apple Silicon)
```

Or via make (from `platform-api/`): `make e2e`, `make e2e-all-dbs`.

- `E2E_DB` = `postgres` (default) | `sqlite` | `sqlserver`.
- `E2E_KEEP=1` leaves the stack up after the run for inspection.
- `E2E_TAGS=@smoke` runs a tag subset. The `@multigateway` scenario runs only on
  the postgres stack (the only one wired with a second gateway) and is otherwise
  skipped automatically.
- `PA_HOST_PORT` / `GW_HTTP_PORT` / `GW2_HTTP_PORT` override the published host
  ports to avoid clashing with other local stacks (defaults 9243 / 18080 / 18081).

### Scenarios

1. **An API deployed to a gateway is served by the data plane** — deploy, then a
   request to the ingress returns 200 via Envoy; a path outside the API context
   returns 404.
2. **Undeploy / redeploy** — undeploying stops the data plane serving the API
   (404); redeploying restores it (200).
3. **Multi-gateway** (`@multigateway`, postgres) — the same API deployed to two
   gateways is served by both (fan-out), and undeploying from one leaves the
   other serving (per-gateway isolation).
4. **LLM proxy deployment** (`@llm`) — an LLM provider (built-in `openai`
   template, upstream pointed at the in-stack `mock-openapi` Prism service with a
   **dummy API key**) and a proxy referencing it are created in platform-api and
   deployed to the gateway; a `POST <proxy-context>/chat/completions` to the
   ingress then returns a `chat.completion` response proxied from the mock. No
   real OpenAI endpoint or credential is involved. The mock upstream is published
   on host port `4010` (override with `MOCK_OPENAI_PORT`).

## Status — passing on all three databases

The full live-traffic scenario passes on **SQLite, PostgreSQL and SQL Server**
(verified locally; SQL Server via `azure-sql-edge` on Apple Silicon).

Bugs this harness surfaced and that are fixed alongside it:
- platform-api image build (`go.sum` missing `go-mssqldb` → `go mod tidy`).
- platform-api SQL Server `LIMIT` and cascade-path/self-ref schema issues.
- gateway eventhub `INSERT … ON CONFLICT` (invalid on SQL Server) in the
  deployment event-publish path → made dialect-aware
  (`common/eventhub/sqlbackend.go`).
