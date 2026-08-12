# FindFore — Architecture

**Last Updated:** August 11, 2026  
**Status:** Living document — system shape, boundaries, and diagrams.

> **Product direction:** [`VISION.md`](./VISION.md)  
> **Day-to-day engineering rules:** [`CLAUDE.md`](./CLAUDE.md)

---

## Guiding Principles

**Principle 1 — Optimize for deleting code.**  
The best feature is the one you never had to write.

**Principle 2 — Every external service gets an adapter.**  
Today that includes Lightspeed, Google Maps, Google Auth, Stripe, SendGrid, and push notifications. Tomorrow you can swap providers without rewriting business logic.

**Principle 3 — Business logic owns the truth.**  
Never let SQL, React, or external APIs make business decisions.

**Principle 4 — Everything is observable.**  
Every important action should answer: Who? What? When? How long? Did it succeed?

Apply Hexagonal Architecture at **system boundaries** (database, booking providers, notifications, payments, storage, maps, external APIs). Avoid unnecessary abstractions inside simple domain code.

---

## Booking Provider Stack

Booking is infrastructure behind a stable interface. The frontend and domain never talk to a tee-sheet vendor directly — they go through the Go API and a booking service that depends on a provider port.

```text
Frontend
   ↓
Go API
   ↓
Booking Service
   ↓
Provider Interface
   ↓
┌──────────┬──────────┬──────────┐
│Lightspeed│  ForeUP  │ GolfNow  │
└──────────┴──────────┴──────────┘
```

```mermaid
flowchart TB
  FE[Frontend]
  API[Go API]
  BS[Booking Service]
  PI[Provider Interface]
  LS[Lightspeed]
  FU[ForeUP]
  GN[GolfNow]

  FE --> API --> BS --> PI
  PI --> LS
  PI --> FU
  PI --> GN
```

| Layer | Responsibility |
|---|---|
| Frontend | UX for discovery, invites, and booking flows — no vendor-specific logic |
| Go API | HTTP authz, request validation, response mapping |
| Booking Service | Domain rules for availability, holds, invites, and booking state |
| Provider Interface | Port that booking adapters implement |
| Lightspeed / ForeUP / GolfNow | Concrete adapters; swappable without changing the service |

### Social events vs booking inventory (today)

```text
courses
  ├── course_providers
  ├── events  → tee_time_id? → tee_times
  └── tee_times
```

- Provider owns real tee-sheet inventory; FindFore stores a **normalized representation/cache**.
- Association is always `event → tee_time` (nullable), never the reverse.
- Join/accept capacity uses `SELECT … FOR UPDATE`; `UNIQUE (player_id, event_id)` is the membership safety net.
- Production migrations must not silently delete user data without an explicit, reviewed data-migration strategy.

See **Booking domain model** below. Schema: `tee_time_providers`, reservation tables, and `events.planned_starts_at` (migrations `000009`–`000011`).

---

## Booking domain model

**Status:** Implemented in schema (`000009`–`000011`) + application/booking + Lightspeed stub adapter. HTTP booking routes and live Lightspeed HTTP mapping are still deferred.

**Locked decisions**

| Decision | Choice |
|---|---|
| Event time | **Option B:** `planned_starts_at` + optional `tee_time_id` — semantically distinct; no DB equality constraint |
| Domain | Provider-agnostic entities and state machines |
| First walkthrough | Lightspeed as concrete pressure-test; every step must still work for ForeUP (else push into the adapter) |
| Party booking | `reservations` + `reservation_players`; event ≠ reservation |

```mermaid
flowchart TB
  course[courses]
  cp[course_providers]
  tee[tee_times]
  ttp[tee_time_providers]
  res[reservations]
  rp[reservation_players]
  event[events]

  course --> cp
  course --> tee
  tee --> ttp
  tee --> res
  res --> rp
  event -->|"tee_time_id optional"| tee
```

```text
                 FindFore domain
                       │
                 ┌─────┴─────┐
                 │           │
              TeeTime    Reservation
                 │           │
                 └─────┬─────┘
                       │
                 Provider identity
                       │
              ┌────────┴────────┐
              │                 │
          Lightspeed          ForeUP
```

### Entities

| Entity | Owns | Does not own |
|---|---|---|
| `courses` | Place + IANA timezone (long-term NOT NULL) | Vendor IDs |
| `course_providers` | Immutable `(provider, external_id) → course` | Tee sheet |
| `tee_times` | FindFore slot: `course_id`, `starts_at`, cached holes/capacity/slots/price | Vendor slot id |
| `tee_time_providers` | Immutable `(provider, external_id) → tee_time` | Social invites |
| `reservations` | Party booking lifecycle against a tee time | Feed / friends |
| `reservation_players` | Party members (`player_id` and/or `guest_name`) | Provider payload |
| `events` | Social round: `planned_starts_at`, optional `tee_time_id` | Inventory / payment |

**Association rules**

- Event and reservation are **siblings** under a tee time, not the same row.
- Many reservation attempts over time; at most one **non-terminal** (`pending` / `held` / `confirmed`) per tee time (app + later partial unique).
- Planned vs tee-time instants may diverge (planned 8:00, selected 8:20). UX may warn; the DB must not reject.

**Schema notes**

- `tee_times` — cached `capacity`, `available_slots`, `price_cents`, `currency`; status `available \| held \| booked \| cancelled`.
- `tee_time_providers` — `UNIQUE(provider, external_id)`, immutable reassignment (same as `course_providers`).
- `reservations` / `reservation_players` — party booking; partial unique on active statuses per tee time.
- `events.planned_starts_at` — display play time = linked `tee_times.starts_at` if present, else planned.

### State machines

**TeeTime**

```text
available ──hold/book──► held ──confirm──► booked
    │                      │
    │                      └──expire/fail──► available
    └──sync cancel────────► cancelled
```

**Reservation**

```text
pending ──► held ──► confirmed
   │          │          │
   │          │          └──► cancelled
   │          └──expire/fail──► failed
   └──fail──► failed
```

Terminal: `cancelled`, `failed`, and `confirmed` until cancel. After fail/cancel, retry creates a **new** reservation row (unless resuming the same `pending`/`held`).

### Twelve scenario walkthroughs

Domain stays provider-agnostic. Lightspeed is the concrete walkthrough; **ForeUP health check** = “same domain transitions, different adapter DTO mapping.”

| # | Flow | FindFore domain | Lightspeed (adapter only) | ForeUP health check |
|---|---|---|---|---|
| 1 | Search tee times | Upsert `tee_times` + `tee_time_providers` by external id; query by course + window | Fetch tee sheet; map slots → domain | Same upsert; different DTO map |
| 2 | Display availability | Read cache (`available_slots`, status) | Optional light refresh | Same read model |
| 3 | User selects tee time | Client holds FindFore `tee_time_id`; no reservation yet | N/A | Same |
| 4 | Begin booking | Create `reservations` + `reservation_players`; `pending`/`held` + optional `hold_expires_at` | Hold/book API if supported | No hold API → skip `held`, confirm-only |
| 5 | Provider succeeds | `confirmed` + `external_reservation_id`; refresh tee_time cache/status | Parse confirmation id | Same transitions |
| 6 | Provider fails | `failed` + reason; tee_time stays bookable | Map error codes | Same |
| 7 | Connection lost | Idempotency / resume in-flight; unique external id prevents double-book | Safe retry of same call | Same pattern |
| 8 | User retries | Resume `pending`/`held`, or new row if terminal | Same adapter ops | Same |
| 9 | Availability changes | Sync/webhook updates cache only; never invent unknown slots | Poll or webhook | Adapter-specific sync |
| 10 | User cancels | Adapter cancel → `cancelled`; refresh tee_time | Cancel API | Same |
| 11 | Provider cancel fails | Stay `confirmed`; surface error; do not free inventory | Error mapping | Same |
| 12 | Social event linked | Set `events.tee_time_id`; keep planned time even if times differ | N/A | Same |

Highest-risk flows: **4–8** (hold, fail, retry, idempotency).

### Hexagonal placement

- Use cases: [`internal/application/booking/`](./internal/application/booking/) — `search_availability`, `hold`, `confirm`, `cancel`
- Port: `BookingProvider`; adapters under `internal/adapter/outbound/{lightspeed,foreup,...}`
- Domain never imports vendor DTOs

### Open items (first adapter kickoff — not blocking this design)

- Hold TTL defaults when the provider supports holds
- Webhook vs poll for availability sync
- Currency / money source of truth (`price_cents` cache vs live quote)
- Stale cache TTL for search results
- Derive `courses.timezone` from geo/address; retire `America/Denver` fallback

### Remaining work

1. Wire HTTP booking routes (search / begin / confirm / cancel)
2. Lightspeed live HTTP client (credentials + DTO mapping) — stub exists at `internal/adapter/outbound/lightspeed`
3. Resolve open items (hold TTL, webhook vs poll, money/cache TTL)

## Social → Identity Loop

The retention path is not a booking checkout — it is a compounding loop from connection to permanent golf identity.

```text
Friend Request
   ↓
Notification
   ↓
Activity Feed
   ↓
Round
   ↓
History
   ↓
Identity
```

```mermaid
flowchart TB
  FR[Friend Request]
  N[Notification]
  AF[Activity Feed]
  R[Round]
  H[History]
  I[Identity]

  FR --> N --> AF --> R --> H --> I
```

| Stage | Role |
|---|---|
| Friend Request | Starts a relationship; unlocks coordination |
| Notification | Pulls the user back when something social happens |
| Activity Feed | Surfaces community and “need one more” moments |
| Round | The shared experience (play + companion features) |
| History | Persists scores, photos, and who you played with |
| Identity | The durable golf profile that makes leaving costly |

---

## Application packages

Use cases live under `internal/application/<domain>/`, one Go package per domain. Prefer **action-oriented files** that grow as the domain matures:

```text
internal/application/booking/
  service.go          # Service + constructor
  book.go
  cancel.go
  availability.go
  dto.go              # only when entities are not enough
```

| Package | Role today |
|---|---|
| `players`, `sessions`, `courses`, `events`, `feed`, `friends` | Live application services |
| `booking` | Scaffold for provider booking (Lightspeed / ForeUP / GolfNow) |
| `groups`, `notifications` | Reserved; fill in as those pillars land |
| `apperr` | Shared validation errors only |

Do not invent abstractions early — split files and add DTOs when a domain earns them.

---

## How to Use This Document

- **Adding a provider?** Implement the provider interface; do not branch vendor logic into the frontend or booking service.
- **Changing booking UX?** Stay above the Go API — provider details stay behind the adapter.
- **Building social features?** Trace them through this loop — connection → return visit → round → history → identity.
- **Growing a domain?** Add action files under its `internal/application/<domain>/` package; keep HTTP handlers thin.
- **Drawing a new system?** Prefer a short mermaid diagram here; keep implementation detail in code and `CLAUDE.md`.
