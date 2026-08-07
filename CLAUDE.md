# FindFore Development Guidelines for Claude

**Last Updated:** August 07, 2026
**Project Name:** FindFore
**Repository:** https://github.com/errabun/FindFore
**Version:** 1.0

## 1. Project Overview & Vision

> **See [`VISION.md`](./VISION.md) for the full north star, roadmap, and strategic context.**  
> **See [`ARCHITECTURE.md`](./ARCHITECTURE.md) for system diagrams and architecture principles.**

- **One-Line Pitch:** The operating system for golfers — not a booking app. Booking is infrastructure; the product is where golfers organize their entire golf life.
- **Core Purpose:** A mobile-first platform spanning four pillars — **Social**, **Booking**, **Playing**, and **Golf Identity** — that replaces the fragmented golf ecosystem (separate apps for handicaps, scores, GPS, booking, coordination).
- **Killer Feature:** Group coordination — "need one more player" → post → notify → join → chat → pay → play → share scores → make friends. This solves a weekly problem every golfer has.
- **Target Users:** Golfers of all levels who want to play more often with compatible people and keep their golf life in one place.
- **Moat:** Network effects — golf history, social connections, and group coordination that compound over time. Booking APIs alone are not defensible.
- **Success Metrics:** Active communities, group rounds coordinated, user retention between rounds, tee times created/booked, messages/invites sent.

**Every feature, UI decision, and code change must support the four-pillar vision in `VISION.md`. When in doubt, prioritize social retention and group coordination over booking polish.**

## 2. Architecture Principles (Non-Negotiable)

Apply Hexagonal Architecture at system boundaries (database, booking providers, notifications, payments, storage, maps, external APIs). Avoid unnecessary abstractions inside simple domain code.

- **Ports & Adapters at the edges:** Domain holds business rules (tee time creation, privacy (private vs public), invitations, RSVPs, vacancies, social interactions). Ports define interfaces for external concerns; adapters implement them (PostgreSQL via sqlc, Google services, Mantine UI, Redux, etc.).
- **DRY Principle:** Aggressively eliminate duplication across layers. Extract reusable domain services, utilities, ports, and UI components.
- **Existing Structure Respect:** Build upon the current repo layout (`frontend/`, `internal/`, `migrations/`, `sqlc/`, `cmd/`, etc.). Application use cases live under `internal/application/{players,sessions,courses,events,feed,friends,booking,...}` with action-oriented files (`service.go`, `create.go`, …) that grow as each domain matures. Evolve toward hexagonal at boundaries without unnecessary disruption.
- **Clean Separation:** Business logic must stay in domain; never leak into UI, DB, or infrastructure.

### Guiding principles

**Principle 1 — Optimize for deleting code.**  
The best feature is the one you never had to write.

**Principle 2 — Every external service gets an adapter.**  
Today that includes Lightspeed, Google Maps, Google Auth, Stripe, SendGrid, and push notifications. Tomorrow you can swap providers without rewriting business logic.

**Principle 3 — Business logic owns the truth.**  
Never let SQL, React, or external APIs make business decisions.

**Principle 4 — Everything is observable.**  
Every important action should answer: Who? What? When? How long? Did it succeed?

## 3. Tech Stack (Locked In)
- **Frontend:** React + TypeScript, mobile-first (PWA-capable, responsive design). Use **Mantine** as the sole component library — always reference the latest Mantine v7+ documentation for components, themes, hooks, and patterns.
- **State Management:** Redux (with Redux Toolkit preferred for modern patterns).
- **Backend:** Go (latest stable) with PostgreSQL. Leverage existing sqlc for type-safe queries and migrations.
- **Styling:** Mantine theme customization for golf-inspired branding.
- **Authentication:** Prioritize Google Identity / Google Authenticator.
- **Maps:** Integrate Google Maps where relevant (course locations, tee time meetups).
- **Hosting:** Google Cloud Run (containerized Go server + static frontend). PostgreSQL on Cloud SQL. CI/CD via Cloud Build (`cloudbuild.yaml`). Artifact Registry for images. Secret Manager for runtime secrets.
- **Google Cloud:** Use services where beneficial (Auth, Maps, Cloud SQL for PostgreSQL, Storage, Pub/Sub for notifications, etc.). Developer account: errabun@gmail.com.
- **Other:** Offline/responsive considerations for mobile golf use (poor signal on courses).

**Major architectural libraries require approval.** Small utility libraries are acceptable when they reduce complexity and are actively maintained.

## 4. UI/UX & Design Responsibility
- You (Claude) serve as the design lead. Deliver fresh, modern, premium golf-inspired interfaces that feel energetic and community-oriented.
- **Core Aesthetic:** Clean, modern, golf-vibe (deep greens, teals, fresh neutrals). High contrast, large/glove-friendly tap targets, excellent touch experience.
- **Mobile-First:** Design and implement for mobile from day one. Ensure excellent PWA behavior, responsive Mantine components, and dark mode support.
- **Customization:** Built-in user options for:
  - Dark mode (Mantine native + persistence)
  - Multiple color profiles/themes
  - Accessibility (high contrast, keyboard nav, screen readers)
- **Mantine Usage:** Strictly follow latest official Mantine docs. Extend the theme for branding. Provide component breakdowns and theme code when proposing screens.
- **Golf UX Specifics:** One-tap actions for scoring/invites where possible, clear private vs public tee time flows, real-time feedback for reservations.

## 5. Core Features Scope

Organized by the four pillars defined in `VISION.md`:

**Social (retention engine)**
- Friend list, follow golfers, golf groups, community feed, photos, achievements
- Group coordination flow (killer feature): post → notify → join → chat → pay → play → share
- Clubs, leagues (Phase 2+)

**Booking (acquisition infrastructure)**
- Course search and discovery (Golf Course API today; multi-provider Phase 2)
- Tee time creation: private (invite friends) or public (open spots)
- Invitations, RSVPs, waiting lists, cancellation alerts
- Smart Tee Time Feed — proactive surface, not manual search (Phase 2)
- Provider integration (Lightspeed or similar) — Phase 1 remaining work

**Playing (on-course companion)**
- GPS, digital scorecard, live leaderboard, betting formats (Phase 2)
- Pace tracking (Phase 2+)

**Golf Identity (profile layer)**
- Rich golfer profile: handicap, rounds played, favorite course, achievements, equipment, photos
- Basic profile editing exists today; expand toward full identity layer over time

**Cross-cutting**
- Notifications for invites, reservations, messages, coordination posts
- Messaging / group chat tied to tee times and rounds
- Respect privacy rules in all social/tee time flows
- AI enhancements (recommendations, scheduling, post-round summaries) — supporting, not centerpiece

## 6. Golf Course Data
- Claude must research and propose **cost-effective** options for course data (names, locations, pars, tees, coordinates, etc.).
- Prioritize free/cheap APIs, CSV imports (e.g., golfapi.io style), or Google Maps integration. Suggest caching strategies and import pipelines that fit PostgreSQL + sqlc.

## 7. Testing Strategy
- For every new feature or potentially breaking change, implement tests **in the same PR**.
- **Frontend:** Vitest + React Testing Library (`cd frontend && npm test` / `npm run test:watch`). Prefer testing user-visible behavior; mock at the adapter/port boundary.
- **Backend:** Go `testing` + testify table tests for services (fake repos) and `httptest` for authz-sensitive HTTP paths (`go test ./...`).
- **CI:** Cloud Build runs `go test ./...` and `frontend` `npm ci && npm test && npm run build` **before** the Docker image build/deploy. Failures block deploy.
- **Focus coverage on:**
  - Domain/hexagonal logic (services)
  - Private vs public tee time rules and invitation flows
  - AuthZ / IDOR paths (JWT actor, host-only mutate, friendship accept/decline)
  - Login validation and friend-request UI actions
- **Security-sensitive changes** (authz, privacy, payments later) require a service test and at least one handler/HTTP test.
- **E2E:** Cypress is reserved for a thin smoke suite later (login → dashboard; friend request) after APIs stabilize — not a substitute for unit/RTL coverage.
- Conventional commit scopes: `test:`, `chore(ci):`, `feat(security):`.

## 8. Google Cloud & Infrastructure
- **Runtime:** Cloud Run service built from repo `Dockerfile` (multi-stage: frontend build → Go binary → distroless).
- **Database:** Cloud SQL PostgreSQL via Unix socket (`INSTANCE_CONNECTION_NAME`, `DB_USER`, `DB_PASS`, `DB_NAME`) or `DATABASE_URL` for local dev.
- **Migrations are immutable.** Never edit an old migration — always create a new one. Every migration should be reversible when practical (`migrations/*.up.sql` + matching `*.down.sql`). Local: `go run ./cmd/migrate -direction up`. Cloud SQL: `./scripts/gcp-migrate.sh`.
- **Deploy:** `cloudbuild.yaml` builds image, pushes to Artifact Registry, deploys to Cloud Run. Connect repo in Cloud Console or use `gcloud builds submit`.
- **Secrets:** JWT, DB password, API keys in Secret Manager — never in the image or repo.
- **Storage (future):** Cloud Storage for user uploads (profile photos, feed images).
- **Domain:** `findfore.com` (registered via GoDaddy). Map to Cloud Run once the app is live (Cloud Run → Manage custom domains, then update GoDaddy DNS). Use `findfore.com` (or `www.findfore.com`) in CORS `ALLOWED_ORIGINS`, PWA manifest `start_url`, and share links once mapped.
- **Cost hygiene:** Cloud Run scales to zero. Cloud SQL is scheduled via `scripts/gcp-sql-schedule.sh` (weekday 8am–8pm America/Denver). Storage + public IP still bill when stopped. Use `./scripts/gcp-sql-schedule.sh start|stop` for off-hours work.
- **Logging:** The Go server emits **JSON `log/slog`** to stdout (Cloud Run → Cloud Logging). Every request gets an `X-Request-ID` (honors inbound id or `X-Cloud-Trace-Context`). 5xx responses log the underlying `err` with `request_id` / `player_id` / route; panics are recovered and logged with stack. In Logs Explorer, filter on `jsonPayload.msg="handler_error"` or severity≥ERROR. Optional: create a log-based alert on `severity>=ERROR`. Skip Error Reporting SDK until volume justifies it.
- Leverage Google services (Auth, Maps, etc.) where they provide good integration or cost benefits.

## 9. Collaboration & Prompting Rules for Claude
- **Always reference this file, `VISION.md`, and `ARCHITECTURE.md`** at the start of sessions or major tasks.
- When told "refer to guidelines," quote relevant sections.
- Think step-by-step: show reasoning for hexagonal ports/adapters, DRY refactors, Mantine usage, and design proposals.
- For design: Provide modern Mantine code examples with multiple options when choices exist.
- For Google services or course data: Give concrete, cost-effective implementation suggestions.
- If anything is ambiguous (privacy rules, architecture boundaries, UI choices, data sources), ask clarifying questions.
- Prefer small, iterative, reviewable changes.
- Never hallucinate Mantine APIs or golf rules — align with official docs/sources.

## 10. Version Control & Documentation
- Use conventional commits.
- Keep `CHANGELOG.md` updated.
- Document key domain ports, adapters, and any non-obvious decisions.
