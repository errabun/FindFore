# FindFore Development Guidelines for Claude

**Last Updated:** April 04, 2026
**Project Name:** FindFore
**Repository:** https://github.com/errabun/FindFore
**Version:** 1.0

## 1. Project Overview & Vision
- **Core Purpose:** FindFore is a mobile-first social community platform for golfers. It combines Facebook-style networking (profiles, following, feed, interactions) with seamless tee time scheduling and coordination.
- **Key Experience:** Users create tee times (private with specific friend invites OR public where anyone can reserve open spots), message about rounds, build their golf network, discover groups, and book/play together more easily.
- **Target Users:** Golfers of all levels who want to play more often with compatible people.
- **Unique Value Proposition:** The social golf app that makes finding playing partners and coordinating tee times effortless and fun.
- **Success Metrics:** Active communities, tee times created/booked, user retention, messages/invites sent, and successful group rounds.

**Every feature, UI decision, and code change must support this social + tee-time coordination vision.**

## 2. Architecture Principles (Non-Negotiable)
- **Hexagonal Architecture (Ports & Adapters):** Apply strictly on both backend (Go) and frontend (TypeScript/React).
  - Domain layer holds all business rules (tee time creation, privacy (private vs public), invitations, RSVPs, vacancies, social interactions).
  - Ports define interfaces for external concerns.
  - Adapters implement ports (PostgreSQL via sqlc, Google services, Mantine UI, Redux, etc.).
- **DRY Principle:** Aggressively eliminate duplication across layers. Extract reusable domain services, utilities, ports, and UI components.
- **Existing Structure Respect:** Build upon the current repo layout (`frontend/`, `internal/`, `migrations/`, `sqlc/`, `seed/`, etc.). Evolve it toward full hexagonal without unnecessary disruption.
- **Clean Separation:** Business logic must stay in domain; never leak into UI, DB, or infrastructure.

## 3. Tech Stack (Locked In)
- **Frontend:** React + TypeScript, mobile-first (PWA-capable, responsive design). Use **Mantine** as the sole component library — always reference the latest Mantine v7+ documentation for components, themes, hooks, and patterns.
- **State Management:** Redux (with Redux Toolkit preferred for modern patterns).
- **Backend:** Go (latest stable) with PostgreSQL. Leverage existing sqlc for type-safe queries, migrations, and seed scripts.
- **Styling:** Mantine theme customization for golf-inspired branding.
- **Authentication:** Prioritize Google Identity / Google Authenticator.
- **Maps:** Integrate Google Maps where relevant (course locations, tee time meetups).
- **Hosting:** Currently Heroku (Procfile present). Actively propose and assist migration to cheaper Google Cloud alternatives (Cloud Run, etc.) suitable for low-traffic development phase.
- **Google Cloud:** Use services where beneficial (Auth, Maps, Cloud SQL for PostgreSQL, Storage, Pub/Sub for notifications, etc.). Developer account: errabun@gmail.com.
- **Other:** Offline/responsive considerations for mobile golf use (poor signal on courses).

**Never introduce new libraries, frameworks, or major stack changes without explicit approval.**

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
- User profiles, following, community feed, interactions
- Create tee time: private (invite specific friends) or public (open vacancies anyone can claim)
- Invitations, RSVPs, messaging tied to tee times
- Notifications for invites, reservations, messages
- Course discovery/integration (using researched cost-effective data sources)
- Respect privacy rules in all social/tee time flows

## 6. Golf Course Data
- Claude must research and propose **cost-effective** options for course data (names, locations, pars, tees, coordinates, etc.).
- Prioritize free/cheap APIs, CSV imports (e.g., golfapi.io style), or Google Maps integration. Suggest caching strategies and import pipelines that fit PostgreSQL + sqlc.

## 7. Testing Strategy
- For every new feature or potentially breaking change, implement tests.
- Frontend: React Testing Library (RTL) strongly preferred.
- Backend: Standard Go testing (with testify if helpful).
- Focus on:
  - Domain/hexagonal logic
  - Private vs public tee time rules and invitation flows
  - Any business logic that affects user trust or data integrity

## 8. Google Cloud & Infrastructure
- Leverage Google services (Auth, Maps, etc.) where they provide good integration or cost benefits.
- Plan for eventual migration from Heroku to Google Cloud Run / Cloud SQL, etc., during development while keeping costs low.

## 9. Collaboration & Prompting Rules for Claude
- **Always reference this file** at the start of sessions or major tasks.
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
