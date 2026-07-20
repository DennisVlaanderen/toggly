# Copilot instructions

Guidance for GitHub Copilot when working with code in this repository.

## Project overview

Toggly is a distributed feature-flagging toolkit — **a work in progress, not
yet a working product**. Core premise: a Raft-replicated cluster of servers
stores flags and serves them over both a custom binary TCP protocol (FQDP)
and an HTTP API, so applications get near-real-time flag updates without
polling. See [README.md](../README.md), [IDEA.md](../IDEA.md),
[docs/premise.md](../docs/premise.md) (design intent) and
[docs/fqdp.md](../docs/fqdp.md) (wire protocol spec) for the full picture.
[docs/user-stories.md](../docs/user-stories.md) is the working backlog and is
useful for understanding *why* a piece of the API/UI looks the way it does,
including places where the current code is intentionally behind the spec.

The repo has three parts, each with its own toolchain:

- `backend/` — Go server. Module root is the repo root (module `toggly`), entry
  point `backend/cmd/server/main.go`.
- `ui/` — SvelteKit 5 admin webapp (TypeScript, Tailwind, pnpm).
- `proxy/` — nginx config fronting backend + UI (see `docker-compose.yml`).

## Commands

Go backend (run from repo root):

```sh
go build -v ./...
go test -v ./...
go test -v ./backend/internal/store/...   # single package
go test -v -run TestName ./backend/internal/auth/...  # single test
```

UI (run from `ui/`):

```sh
pnpm install
pnpm run build
pnpm run check   # svelte-kit sync + svelte-check (types)
pnpm run lint    # prettier --check + eslint
pnpm run format  # prettier --write
pnpm test        # vitest --run, includes Playwright browser-mode tests
pnpm run test:unit -- path/to/file.spec.ts   # single test file, watch mode
```

CI mirrors these exactly and is path-filtered per side —
[.github/workflows/go.yml](./workflows/go.yml) only triggers on `backend/**`,
[.github/workflows/node.js.yml](./workflows/node.js.yml) only on `ui/**`.
**Only run the suite for the side you actually touched** unless the change is
cross-cutting.

Full local stack: `docker-compose.yml` at the repo root (backend + UI + nginx
proxy). Backend env vars are documented in
[.env.example](../.env.example) (Raft node identity/storage, JWT secret,
bootstrap admin credentials) — copy to `.env` only to override the
docker-compose defaults; never commit real secrets.

## Backend architecture (`backend/internal/`)

`main.go` wires three independent pieces together and does not otherwise
contain logic: it opens the `store.Store`, starts the FQDP TCP listener in a
goroutine, and starts the HTTP API — all sharing the same underlying flag
store.

- **`store/`** — the flag data layer, built directly on `hashicorp/raft`:
  - `cluster.go` (`newRaft`) sets up a Raft node backed by BoltDB
    (`raft-boltdb`) for the log/stable store and a file snapshot store. On
    first run with `Bootstrap: true` it bootstraps a single-member cluster;
    on restart it rejoins existing on-disk state instead of re-bootstrapping.
  - `fsm.go` is the Raft FSM: an in-memory `map[string]Flag` guarded by a
    `sync.RWMutex`. There is deliberately no separate embedded DB for
    application data — Raft's replicated log plus FSM snapshot/restore is the
    durability mechanism.
  - `store.go` (`Store`) is the public API: `Get`/`List` read the FSM
    directly; `Set` only succeeds when the local node is Raft leader (it
    encodes a `command` and calls `raft.Apply`) — with today's single-node
    bootstrap that's always true, but multi-node behavior depends on this.
  - `flag.go` defines `Flag{Key, Enabled, Value, Version}`; `Version` is set
    from the Raft log index on apply, not a separately tracked counter.

- **`api/`** — `net/http` (stdlib `ServeMux` with Go 1.22+ method patterns,
  e.g. `"GET /api/flags"`), no framework. `RegisterRoutes` wires
  `/api/health`, `/api/flags` (GET for admin+user, POST for admin only),
  `/api/auth/login`, `/api/auth/me`. `middleware.go`'s `requireRoles` wraps a
  handler with bearer-token parsing + role check via the `auth` package.
  Route handlers currently do minimal validation — check `api.go` directly
  rather than assuming REST conventions beyond what's there.

- **`auth/`** — `Service` in `service.go` issues/parses HS256 JWTs
  (`golang-jwt/jwt`) and hashes passwords with bcrypt. There are exactly two
  accounts today: a configurable admin (`TOGGLY_ADMIN_USERNAME` /
  `TOGGLY_ADMIN_PASSWORD`, bcrypt-hashed at startup) and a hardcoded
  `user`/`user123` — this is not yet a real user store (see Epic 7 in
  `docs/user-stories.md` for where this is headed). Roles are the plain
  strings `RoleAdmin` ("admin") / `RoleUser` ("user").

- **`fqdp/`** — the FQDP binary protocol server described in
  `docs/fqdp.md`. **Currently a stub**: `server.go` implements the
  length-prefixed framing (`readFrame`) and accepts connections, but only the
  Handshake message type is even parsed, and it always responds with an
  "not implemented" `Error` frame (`ErrHandshakeNotImplemented`) — no session,
  Query, Subscribe, or Update handling exists yet. `message_types.go` /
  `errors.go` hold the wire constants. Treat `docs/fqdp.md` as the target
  spec, not a description of current behavior, when working in this package.

## UI architecture (`ui/src/`)

SvelteKit with server-side rendering; the app never talks to FQDP (browsers
can't hold raw TCP sockets) — it only calls the backend's HTTP API, and only
from server-side code (`+page.server.ts` / `lib/server/`), never directly
from the browser.

- **`lib/server/auth.ts`** is the only bridge to the backend API
  (`TOGGLY_API_ORIGIN` env var, default `http://127.0.0.1:8080`). `login()`
  POSTs to `/api/auth/login`; `getSession()` reads the `toggly.auth` httpOnly
  cookie and validates it against `/api/auth/me` on every call (no local
  session cache) — the JWT itself lives only in that cookie.
  `setAuthCookie`/`clearAuthCookie` manage it (`secure` in non-dev).
- **`routes/login/+page.server.ts`** performs the login action;
  **`routes/dashboard/+page.server.ts`** shows the pattern for protecting a
  route: call `getSession`, `redirect(303, '/login')` if absent. Follow this
  same pattern for any new authenticated route rather than inventing a
  different guard.
- **`lib/paraglide/`** is generated i18n code (from `project.inlang` /
  `@inlang/paraglide-js`) — don't hand-edit files under here.

## Conventions (from AGENTS.md — keep in sync if that file changes)

- Backend Go code stays inside the existing `backend/internal/*` package
  layout; don't add new top-level packages unless a new concern truly
  warrants one.
- UI code is Prettier-formatted and ESLint-linted; match existing Svelte 5 +
  TypeScript idioms already in `ui/src`.
- `docker-compose.yml` is the reference way to run the full stack locally.
