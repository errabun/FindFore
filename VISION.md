# FindFore — North Star Vision

**Last Updated:** August 07, 2026  
**Status:** Living document — the strategic reference for all product and architectural decisions.

---

## The One-Line Pitch

> **The operating system for golfers.**

Not a tee time booking app. Booking is infrastructure — one feature among many. The product is becoming the place where golfers **organize their entire golf life**.

---

## The Problem

The golf ecosystem is incredibly fragmented:

- Every course has its own booking website
- Friend groups coordinate in text messages
- Handicaps live in one app; scores in another
- GPS is in another; tee time alerts somewhere else
- Tournament organization is somewhere else

**No one has successfully combined these into a cohesive experience.**

---

## Four Pillars

Every feature we build should map to one or more of these pillars. Booking gets users in the door; the other three create retention, identity, and network effects.

### 1. Social — *Why people come back when they aren't booking*

> Golf meets Discord meets Strava.

| Feature Area | Examples |
|---|---|
| Network | Friend list, follow golfers, find players |
| Groups | Golf groups, clubs, leagues |
| Coordination | Invite friends, public rounds, group chat |
| Engagement | Photos, achievements, community feed |

**This is the retention engine.** Users open the app on Tuesday even when they're not playing until Saturday.

### 2. Booking — *How users get into the app*

Booking should feel like part of the social experience, not a separate transaction.

| Feature Area | Examples |
|---|---|
| Discovery | Search nearby courses, live tee sheets, price comparison |
| Coordination | Invite players, split payment, waiting lists |
| Proactive | Cancellation alerts, book instantly, Smart Tee Time Feed |

**Booking is infrastructure, not the product.**

### 3. Playing — *The four-hour companion on the course*

If the app becomes their round companion, engagement increases dramatically.

| Feature Area | Examples |
|---|---|
| On-course tools | GPS, scorecard, distance to pin |
| Competition | Live leaderboard, betting, Nassau, skins, match play |
| Awareness | Pace tracking |

**This is where users spend the most continuous time in the app.**

### 4. Golf Identity — *The profile layer*

Every golfer has a rich identity that reflects who they are on the course:

```
John Smith
├── 12.3 Handicap
├── 128 Rounds Played
├── Favorite Course
├── Average Drive
├── Course Ratings (e.g., Greensburg Rating)
├── Friends
├── Upcoming Tee Times
├── Achievements
├── Equipment
└── Photos
```

Almost like a **LinkedIn profile for golfers** — something people keep updated because it reflects their golfing identity.

---

## Killer Feature: Group Coordination

If we had to pick **one feature** that makes people tell their friends about the app, it wouldn't be booking. It would be **group coordination**.

```
Friday: "Need one more player"
        ↓
Post appears in feed
        ↓
Nearby golfers get notified
        ↓
Someone joins
        ↓
Everyone chats
        ↓
Pays (split payment)
        ↓
Shows up
        ↓
Scores automatically shared
        ↓
Friends made
```

This solves a real problem golfers have **every week**.

---

## Smart Tee Time Feed

Instead of making users hunt across every course website manually, the app becomes **proactive**:

```
Tomorrow Morning

Castle Pines          Bear Dance
7:48 AM               Saturday 8:12 AM
2 Spots Left          Your foursome is available
Friends Discount      Reserve?
Weather 72°
[Book]
```

The app surfaces opportunities — it doesn't wait for users to search.

---

## AI Opportunities

AI is not the centerpiece, but it can genuinely improve the experience:

| Use Case | Value |
|---|---|
| Course recommendations | Based on budget, pace, and travel time |
| Schedule-aware tee times | Suggest times that fit everyone's calendars |
| Cancellation prediction | Alert users when spots are likely to open |
| Team balancing | Match handicaps for fair foursomes |
| Post-round summaries | Auto-generated stats and highlights |
| Practice plans | Based on scoring trends over time |

---

## The Moat: Network Effects

> **Booking APIs are not defensible. Anyone can integrate the same API.**

The moat comes from **network effects**:

```
10 golfers → 100 → 1,000 → 50,000
                ↓
Every golfer has friends
                ↓
Every group organizes here
                ↓
Courses encourage it
                ↓
Leagues use it
                ↓
People don't leave because all their golf history
and social connections are here
```

At scale, the value isn't booking tee times — it's the **community and history** built around the platform.

---

## Business Model

The social network creates recurring engagement; booking creates transaction revenue. They complement each other:

| Revenue Stream | Description |
|---|---|
| Booking commissions | Affiliate revenue from reservation partners |
| Premium subscriptions | Advanced stats, GPS, AI features |
| Course subscriptions | Promotional tools, league management, analytics |
| Sponsored content | Equipment brands, instructors, golf travel |

**We are not competing head-on with GolfNow as a marketplace.** We are trying to become the app golfers open **before, during, and after every round**.

---

## Roadmap Phases

Build toward the full vision from day one — architectural decisions should not require rebuilding core systems as the product grows.

### Phase 1 — MVP (3–6 months)

| Feature | Status |
|---|---|
| User accounts | ✅ Built (email/password + JWT) |
| Profiles | ✅ Built (edit info, password, theme) |
| Friend system | ✅ Built (add/remove friends) |
| Course search | ✅ Built (Golf Course API integration) |
| Tee time creation & coordination | ✅ Built (private/public, invite, accept/decline/join) |
| Community feed | ✅ Built (posts, reactions, replies) |
| Golf groups | ✅ Built (membership + group activity) |
| Tee time booking via provider | 🔲 Waiting on provider API access |
| Group chat | 🔲 Not started |
| Notifications | 🔲 Not started |
| Google Identity auth | 🔲 Not started |

### Phase 2 — Growth (6–12 months)

- Multi-provider booking
- Digital scorecards
- Handicap tracking
- GPS distances
- Payments and expense splitting
- Leagues
- AI tee time recommendations
- Smart Tee Time Feed

### Phase 3 — Platform (12–24 months)

- Public events and tournaments
- Marketplace for instructors and clubs
- Equipment recommendations
- Sponsorships and advertising
- Travel packages
- Comprehensive golf statistics

---

## Architectural Implications

Because we are building toward a platform — not a booking widget — every architectural decision should anticipate growth across all four pillars:

| Decision | Rationale |
|---|---|
| Hexagonal architecture (ports & adapters) | Swap booking providers, GPS services, payment processors without rewriting domain logic |
| Domain-first business rules | Social privacy, tee time visibility, and group coordination rules live in domain layer |
| Extensible user profile model | Golf Identity pillar requires rich, evolving profile data |
| Event-driven notifications (future) | Group coordination flow depends on real-time alerts |
| Activity/history persistence | Network effect moat requires users' golf history to live here permanently |

**Refer to [`CLAUDE.md`](./CLAUDE.md) for implementation standards and tech stack constraints.**  
**Refer to [`ARCHITECTURE.md`](./ARCHITECTURE.md) for system diagrams and architecture principles.**

---

## Feature Success Criteria

Every feature should satisfy **at least one** of these outcomes. If it maps to none, it does not belong in FindFore (or needs to be reframed until it does):

- **Increase rounds played** — help golfers get on the course more often
- **Reduce booking friction** — make finding and locking in a tee time easier
- **Strengthen golf friendships** — deepen connections and make playing together simpler
- **Encourage return visits** — give people a reason to open the app between rounds
- **Create permanent golf history** — capture identity, scores, and social memory that compounds over time

These criteria sit alongside the four pillars: pillars describe *where* a feature lives; these criteria describe *why it is worth building*.

---

## How to Use This Document

- **Starting a new feature?** Identify which pillar(s) it serves, which roadmap phase it belongs to, and which success criterion above it satisfies.
- **Evaluating a design?** Ask: does this feel like part of organizing your golf life, or like a disconnected transaction?
- **Making an architecture choice?** Ask: will this still work when we add playing, identity, and multi-provider booking?
- **Prioritizing work?** Group coordination and social retention come before booking polish. Booking provider integration comes when social coordination is solid. If two options both fit a pillar, prefer the one that more clearly hits a success criterion.
