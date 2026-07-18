# Feature Flagging System — User Story Backlog

*Status: Draft — working document*

## How to use this doc

- Stories are grouped by epic. Each has an ID, actor, story, and acceptance criteria (AC).
- Priority column: **Must-have (v1)**, **Nice-to-have (v1.x)**, or **Later phase**.
- Add comments/edits inline as the backlog evolves; treat IDs as stable references for tickets/issues.

---

## Epic 1: Flag Definition (Developer)

| ID | Priority | Story |
| --- | --- | --- |
| US-1.1 | Must-have | As a developer, I want to create a new feature flag with a unique key, description, and default state, so that I can gate new functionality before it's fully tested. |
| US-1.2 | Must-have | As a developer, I want to define targeting rules (user segments, percentage rollout, environment, attributes), so that I can control exactly who sees a flag's variant. |
| US-1.3 | Must-have | As a developer, I want to archive or delete a flag once it's fully rolled out, so that the system doesn't accumulate stale flags. |
| US-1.4 | Must-have | As a developer, I want to version flag configuration changes, so that I can roll back a bad change quickly. |
| US-1.5 | Must-have | As a developer, I want to declare a flag's value type (boolean, integer, string, or JSON) when I create it, so that clients and the server agree on how to parse and validate its value. |
| US-1.6 | Nice-to-have | As a product owner, I want to enable/disable/archive multiple flags in one action (e.g., filtered by tag or owner), so that I can respond quickly during an incident or cleanup pass without repeating the same change one flag at a time. |

### Epic 1: Acceptance criteria

- **US-1.1**: Flag keys unique per environment/namespace; creation requires key, description, and default value; validation prevents duplicate keys.
- **US-1.2**: Rules support boolean, multivariate, and percentage-based rollout; rules can combine multiple attribute conditions (AND/OR).
- **US-1.3**: Deletion requires confirmation; archived flags hidden from active lists but retained for audit history.
- **US-1.4**: Every config change is versioned with timestamp and author; rollback restores prior version and triggers propagation.
- **US-1.5**: Value type is explicit at creation (boolean, integer, string, or JSON) and cannot silently change afterward without a new flag; matches the value-type enum flags are transmitted with over the wire.
- **US-1.6**: Bulk actions require the same confirmation policy as a single change to the highest-impact flag/environment in the selection; bulk action produces one audit entry per affected flag, linked to a shared batch ID.

---

## Epic 2: Shared State & Dynamic Evaluation (System)

| ID | Priority | Story |
| --- | --- | --- |
| US-2.1 | Must-have | As a system, I need flag state persisted in a shared store, so that all application instances read consistent state regardless of region or node. |
| US-2.2 | Must-have | As a developer, I want flag changes to propagate to running applications without a redeploy, so that I can toggle behavior in near real-time. |
| US-2.3 | Must-have | As a system, I need a fallback/default value strategy when the shared store is unreachable, so that applications degrade gracefully instead of failing. |

### Epic 2: Acceptance criteria

- **US-2.1**: Consistency guarantees documented; replication lag measurable and bounded.
- **US-2.2**: Propagation latency target defined (e.g., <2s); no application restart required.
- **US-2.3**: Local cache/last-known-good value used on outage; incident logged and alertable.

---

## Epic 3: Application Integration

| ID | Priority | Story |
| --- | --- | --- |
| US-3.1 | Must-have | As an application, I want to subscribe to changes for specific flags, so that I'm notified in real time instead of polling. |
| US-3.2 | Must-have | As an application, I want to evaluate a flag against local context (user ID, session attributes, environment), so that targeting rules resolve correctly per request. |
| US-3.3 | Must-have | As an application, I want to programmatically override or update a flag's value from within my own runtime context, so that I can implement local kill-switches or context-driven adjustments. |
| US-3.4 | Nice-to-have | As a developer, I want a client SDK available for multiple languages, so that integration is consistent across services. |
| US-3.5 | Must-have | As a developer, I want the SDK to tell me *why* a flag evaluated the way it did for a given context (e.g., which rule matched, or that it fell back to default), so that I can debug unexpected behavior without engineering support from the flag platform team. |
| US-3.6 | Must-have | As an application, I want to authenticate to the server with a credential scoped to a specific environment, so that a compromised or misconfigured client can't read or write flags outside its intended scope. |

### Epic 3: Acceptance criteria

- **US-3.1**: Supports subscribing to a single flag, a set, or all flags in a namespace.
- **US-3.2**: Evaluation resolves against a passed-in context object without a network round-trip per call.
- **US-3.3**: Application-originated writes are authenticated/authorized, versioned like any other change, and tagged "origin: application" in the audit log.
- **US-3.4**: SDKs expose consistent subscribe/evaluate/update methods and handle reconnect/retry transparently.
- **US-3.5**: Evaluation result includes a machine-readable reason code (e.g., `rule_match`, `default_fallback`, `flag_disabled`) plus the matched rule ID, if any; exposed identically across SDKs.
- **US-3.6**: Credentials are issued per environment (see Epic 7); a credential scoped to Staging is rejected by the server if used to query or write Production flags.

---

## Epic 4: SPA UI/UX for Flag Management (Developer, Product Owner)

### 4A. Flag Overview / Dashboard

| ID | Priority | Story |
| --- | --- | --- |
| US-4.1 | Must-have | As a product owner, I want a dashboard listing all flags with name, description, status, and owner, so that I can scan the state of everything at a glance. |
| US-4.2 | Nice-to-have | As a developer or product owner, I want to distinguish "safe" flags from "in transition" flags at a glance, so that I know where to pay attention. |
| US-4.14 | Must-have | As a developer, I want to create a new flag from the dashboard (key, description, value type, default value, initial environment), so that I don't need direct API/CLI access to get a flag into the system. |

### Epic 4A: Acceptance criteria

- **US-4.1**: List shows flag name, key, status badge (On/Off/Partial X%), environment, last modified, owner. Search by name/key; filter by environment, status, owner, tags. Sort by last modified, name, status. Empty/no-results states designed.
- **US-4.2**: Visual treatment differentiates stable vs. actively-changing flags; flags changed in last 24h flagged as "recent."
- **US-4.14**: Form validates key uniqueness and value type live; new flag defaults to Off/safe value and requires an explicit action to enable in Production; creation is an audited action like any other change (see 4C).

### 4B. Flag Detail View

| ID | Priority | Story |
| --- | --- | --- |
| US-4.3 | Must-have | As a product owner, I want to open a flag and see its targeting configuration in plain language, so that I understand who sees what without needing engineering help. |
| US-4.4 | Must-have | As a product owner, I want to toggle a flag or adjust rollout percentage, with environment-appropriate confirmation, so that I don't make an accidental high-impact change. |
| US-4.5 | Must-have | As a product owner, I want to see the estimated blast radius of a change before confirming, so that I can make an informed decision. |

### Epic 4B: Acceptance criteria

- **US-4.3**: Rules render as readable sentences (e.g., "100% of users in Beta segment, 20% of everyone else"); toggle between plain view and raw/advanced view.
- **US-4.4** (confirmation policy):
  - **Production always requires confirmation** — not configurable, cannot be disabled by any role.
  - Non-production environments (Dev, Staging, QA, etc.) have confirmation **off by default**, but can be **toggled on per environment** by an Admin.
  - Environment-level confirmation setting is itself an auditable change (who enabled/disabled it, when).
  - UI clearly labels which environments currently require confirmation.
- **US-4.5**: Pre-change summary shows subscriber count and, where available, estimated affected user count; shown in the same confirmation modal as US-4.4.

### 4C. Audit & History

| ID | Priority | Story |
| --- | --- | --- |
| US-4.6 | **Must-have** | As a developer or product owner, I want a timeline of every change made to a flag, so that I can correlate an incident with a specific change. |
| US-4.7 | **Must-have** | As a developer, I want to (optionally or required) leave a comment when making a flag change, so that the audit trail has context, not just diffs. |
| US-4.8 | **Must-have** | As a product owner, I want to filter audit history by person, date range, or environment, so that I can investigate a specific incident quickly. |
| US-4.13 | **Must-have** | As a developer or product owner, I want to view/export the audit trail for a flag even after it's archived or deleted, so that historical accountability isn't lost during cleanup. |

### Epic 4C:Acceptance criteria

- **US-4.6**: Timeline shows actor, timestamp, before/after value, one-line reason/comment if provided; entries expandable for full detail.
- **US-4.7**: Comment field present on every change action; configurable as required or optional per environment (e.g., required in Production).
- **US-4.8**: Filters by person, date range, environment.
- **US-4.13**: Audit entries retained and queryable after archive/delete; archived flags remain searchable in a dedicated "archived" view for audit purposes.

### 4D. Collaboration & Notifications

| ID | Priority | Story |
| --- | --- | --- |
| US-4.9 | Nice-to-have (v1.1) | As a developer, I want to subscribe (as a human) to notifications when a specific flag changes, so that I'm alerted if someone else touches something I own. |
| US-4.10 | Later phase | As a product owner, I want to see who else is currently viewing or recently edited a flag, so that I avoid conflicting changes with a teammate. |

### Epic 4D:Acceptance criteria

- **US-4.9**: Per-flag "watch" toggle; notification via in-app indicator and optionally email/Slack.
- **US-4.10**: Lightweight presence indicator ("Jane viewed this 5 min ago") — not full real-time collaborative editing.

### 4E. Access & Safety

| ID | Priority | Story |
| --- | --- | --- |
| US-4.11 | Must-have | As a product owner without engineering permissions, I want the UI to clearly show which actions I'm allowed to take, so that I don't attempt a blocked action. |
| US-4.12 | **Must-have** | As a developer, I want a "revert to previous version" button directly on a flag's history entry, so that rolling back doesn't require manually reconstructing the old config. |

### Epic 4E:Acceptance criteria

- **US-4.11**: Disabled controls show a tooltip explaining the required role, rather than failing silently or after submission.
- **US-4.12**: One-click revert re-applies the exact prior configuration and creates a new audit entry noting it was a revert (linking to the original change).

---

## Epic 6: Environments & Promotion (Developer, Product Owner)

Environments are referenced throughout (US-1.1, US-4.4, US-4.7, US-3.6) but were never defined as a first-class concept in their own right — this epic closes that gap.

| ID | Priority | Story |
| --- | --- | --- |
| US-6.1 | Must-have | As an Admin, I want to create and name environments (e.g., Dev, Staging, Production), so that flag state, confirmation policy, and access can differ per environment. |
| US-6.2 | Must-have | As a developer, I want to promote a flag's configuration from one environment to another (e.g., Staging → Production), so that I don't have to manually re-enter targeting rules I already validated. |
| US-6.3 | Nice-to-have | As a product owner, I want to compare a flag's configuration across two environments side by side, so that I can spot drift before promoting. |

### Epic 6: Acceptance

- **US-6.1**: Environments are named, orderable (for promotion direction), and cannot be deleted while flags or credentials still reference them.
- **US-6.2**: Promotion is an explicit action, subject to the target environment's confirmation policy (US-4.4), and produces an audit entry in both source and target environments referencing each other.
- **US-6.3**: Diff view highlights differing rules/values/rollout percentages per flag between two selected environments.

---

## Epic 7: Identity, Roles & API Credentials (Admin)

RBAC is flagged in this doc's own open-questions log as unspecified, and nothing currently covers how a human account or an application credential comes into existence.

| ID | Priority | Story |
| --- | --- | --- |
| US-7.1 | Must-have | As an Admin, I want to define Viewer/Editor/Admin roles with a fixed set of permitted actions per role, so that access matches responsibility. |
| US-7.2 | Must-have | As an Admin, I want to invite and manage user accounts (assign role, environment scope, deactivate), so that I control who can see or change what. |
| US-7.3 | Must-have | As a developer, I want to generate an API credential (FQDP handshake token or HTTP API key) scoped to one environment, so that an application can authenticate without using a human's credentials. |
| US-7.4 | Must-have | As an Admin, I want to revoke or rotate an application credential, so that I can cut off access immediately if it's compromised or a service is decommissioned. |

### Epic 7: Acceptance criteria

- **US-7.1**: Role → permitted-action matrix is documented and enforced identically by the UI (US-4.11) and the API/FQDP layer — no role check exists only client-side.
- **US-7.2**: Invited users authenticate via their own credentials (not shared logins); deactivation takes effect immediately, including on already-issued tokens.
- **US-7.3**: Credential is bound to exactly one environment (see US-3.6) and one or more scopes (e.g., read-only vs. read-write); shown once at creation, stored hashed thereafter.
- **US-7.4**: Revocation takes effect on next handshake/request, without requiring a server restart; revocation is an audited action.

---

## Epic 8: Operations & Resilience (Operator)

Toggly's premise is a Raft cluster built for high request rates, but today only generic "metrics" exist as a nice-to-have — nothing covers cluster visibility, abuse protection, incident response, or disaster recovery.

| ID | Priority | Story |
| --- | --- | --- |
| US-8.1 | Must-have | As an operator, I want to see cluster health (leader node, node count, replication lag) in the admin UI, so that I know if the system is degraded before users report issues. |
| US-8.2 | Must-have | As an operator, I want new and unauthenticated connections rate-limited per IP/credential, so that a misbehaving or malicious client can't overwhelm the server. |
| US-8.3 | Must-have | As an operator, I want a single "force disable" action on a flag that bypasses normal propagation timing, so that I can kill a flag cluster-wide immediately during an incident. |
| US-8.4 | Nice-to-have | As an Admin, I want to export all flag definitions (and import them into another environment or a fresh cluster), so that I can back up configuration or recover from data loss. |

### Epic 8: Acceptance criteria

- **US-8.1**: UI shows current leader, node list with health status, and an alert state if the cluster lacks quorum.
- **US-8.2**: Rate limits are configurable per deployment; breaches are logged and the offending connection is closed, per `fqdp.md`'s Security Considerations.
- **US-8.3**: Force-disable is itself confirmation-gated (like any Production change) and produces an audit entry distinct from a normal toggle, tagged as an emergency action.
- **US-8.4**: Export produces a complete, versioned snapshot (all flags, rules, environments) in a re-importable format; import validates for key collisions before applying.

---

## Cross-cutting / Non-functional

| ID | Priority | Story |
| --- | --- | --- |
| US-5.1 | Nice-to-have | As an operator, I want metrics on flag evaluation counts, latency, and store health, so that I can monitor system performance. |
| US-5.2 | Must-have | As a security-conscious developer, I want all write operations (UI, API, SDK) authenticated and authorized consistently, so that no path bypasses governance. |
| US-5.3 | Nice-to-have | As a developer, I want the system to behave predictably during a partial network partition, so that flags don't silently diverge without an alert. |
| US-5.4 | Nice-to-have (v1.1) | As a product owner, I want to schedule a flag change (e.g., enable at a specific time), so that timed launches don't require someone to manually flip the flag at the right moment. |
| US-5.5 | Later phase | As a developer, I want to declare that one flag depends on another (e.g., Flag B only relevant if Flag A is enabled), so that the UI can warn me about inconsistent configurations. |

### Non-functional Acceptance criteria

- **US-5.4**: Scheduled change is visible in the flag's detail view before it fires, cancellable up until execution, and produces an audit entry attributing the change to "scheduled: <user who scheduled it>" rather than a live actor.
- **US-5.5**: Explicitly deferred past v1 — noted here so its absence reads as a decision, not an oversight; revisit once real dependency chains show up in practice.

---

## FQDP Protocol Alignment

How the `fqdp.md` wire protocol maps to this backlog — called out separately since it's protocol-fit commentary, not a set of user stories.

**Good fit** — these stories map directly onto FQDP as specified:

- US-2.2 (real-time propagation) and US-3.1 (subscribe to flags) → FQDP's Subscribe/Update messages with per-subscription sequence numbers.
- US-2.3 (fallback on outage) → FQDP's reconnect/backfill-by-last-known-sequence flow.

**Over-engineered relative to current v1 scope**: the Admin (0xF0) restricted channel has no backing story anywhere in this backlog — nothing calls for a protocol-level remote-admin command. The dual binary+JSON wire format and the minor-version feature-negotiation bits in HandshakeAck exist to support broad multi-language SDK compatibility, but multi-language SDKs are only Nice-to-have (US-3.4) for v1. Recommend deferring that protocol sophistication (JSON fallback, feature negotiation, Admin channel) until SDK breadth is actually a v1 requirement, and keeping the v1 server implementation to Handshake/Query/Subscribe/Update/Heartbeat only.

**Where FQDP under-serves the backlog** (the protocol needs new capability, not just an implementation, to satisfy these stories):

- No FQDP message carries full targeting *rules* — QueryResponse/Update only carry a resolved value + version. US-3.2 requires the client to evaluate against local context "without a network round-trip per call," which means the client needs the rule definitions themselves, not just a pre-resolved value. Needs a distinct message type (e.g., a `RuleSet`/`FlagDefinition` payload) alongside the existing value-only Query/Update.
- No client→server "write/override" message type distinct from the restricted Admin channel. US-3.3 requires an authenticated application-originated write, but FQDP's client→server messages are only Handshake/Query/Subscribe/Unsubscribe/Ack/Heartbeat — Admin is explicitly "restricted," not meant for routine app writes.
- Update/Query messages carry no actor or comment metadata. US-4.6/US-4.7 need every change attributed and (optionally) commented in the audit trail — if a write originates over FQDP (US-3.3) rather than the HTTP API, the protocol currently has nowhere to carry that attribution.

**Where FQDP doesn't apply, by design**: Epic 4 (the SPA) always talks to the HTTP API, never FQDP directly — browsers can't hold raw TCP sockets. Worth stating explicitly so no future Epic 4 story is scoped assuming direct FQDP access from the browser.

---

## Open questions / decisions log

- Distributed data store technology: **not yet decided** (deliberately excluded from stories above).
- Confirmation policy: **decided** — Production always required, other environments configurable per-environment by Admins.
- Presence indicators (US-4.10): **deprioritized** to a later phase in favor of audit-trail depth.
- RBAC role model (Viewer / Editor / Admin) referenced in US-4.11 and elsewhere — **now specified in Epic 7** (US-7.1); permission matrix itself still needs to be authored as a follow-up artifact.
- Flag-to-flag dependencies (US-5.5): **decided** — explicitly deferred past v1.
- FQDP rule-payload, app-originated write message type, and change-attribution metadata: **not yet decided** — see "FQDP Protocol Alignment" above; needs a protocol addendum before Epic 3/Epic 6 stories that depend on FQDP writes can be implemented.
