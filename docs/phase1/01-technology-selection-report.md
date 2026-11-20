# P1-01 Technology selection report (v0.9, for owner and ChatGPT review)

Author: Claude (lead engineer). Date: 2026-11-19. Status: draft for review; nothing here has been built except the M0.1 skeleton.

**How to read the evidence labels**

- **[V]** Verified from a source listed in section 9, accessed 2026-11-19. (P) marks a primary source, (S) a secondary one (blog, news, aggregator). Secondary facts should be re-checked before money or safety depends on them.
- **[O]** Engineering opinion. Reasonable engineers can disagree.
- **[E]** Estimate. A number I reasoned to, not one I looked up.
- **[U]** Unverified. Plausible, but check before committing.

## 1. Summary of recommendations

| Component | Recommendation | Main alternatives rejected |
|---|---|---|
| API, realtime gateway, workers | Go, one codebase, three runtime roles | TypeScript on Node, Rust, Kotlin or Java, C#, Python, Elixir |
| Web client | TypeScript, React, Vite, installable PWA | Svelte, Vue, Solid, Next.js |
| Private messaging cryptography (runs on devices) | MLS (RFC 9420) through OpenMLS compiled to WASM, behind a small Rust wrapper crate. mls-rs is the fallback | libsignal, Olm and Megolm, custom protocols |
| Database | PostgreSQL 18 with row-level security | MySQL or MariaDB, MongoDB, distributed SQL, SQLite |
| Identity | Ory Kratos (pending spike S-2). Fallback: in-house on audited libraries | Keycloak, Zitadel |
| Jobs and transactional outbox | River on Postgres | Redis queues, Kafka, Temporal |
| Cache, rate limits, pub/sub | Nothing until M5, then Valkey | Redis 8 (AGPL option) |
| Object storage | S3-API interface; SeaweedFS in development; managed S3-compatible or self-hosted in production | MinIO |
| Deployment | Docker Compose behind Caddy on an India-region VPS | Kubernetes, locked-in PaaS |
| Observability | slog JSON logs, OpenTelemetry, Prometheus, Grafana, Loki, self-hosted error tracking | Paid APM |
| Security gate | OWASP ASVS 5.0 (Level 2 for the platform, Level 3 for identity and cryptography components), STRIDE threat models, govulncheck, gitleaks, Trivy | |

## 2. Constraints and criteria

Constraints from the owner's documents: no paid developer tools (D-006); a limited budget that is "a constraint, not an excuse"; production-grade engineering (D-008); genuine end-to-end encryption for private conversations (D-005); India's DPDP regime with many users under 18 (D-002, D-009); portability and no vendor lock-in; development through chat with a single human operator.

Criteria applied to every choice, in the owner's order: security and memory safety, reliability, performance and scalability, concurrency and transaction safety, ecosystem maturity, long-term maintenance, security-update record, developer productivity, deployment portability, total cost. Two further tie-breakers for our situation: can a tiny team operate it, and can another senior team inherit it.

## 3. What changed from my earlier suggestions

| Earlier suggestion | Now | Why |
|---|---|---|
| Server-readable direct messages with a reported-conversation review | Withdrawn | You decided on genuine end-to-end encryption (D-005). |
| TypeScript end to end | Go backend, TypeScript client, Rust only for the crypto core | You asked me to weigh memory safety, supply chain and alternatives. The npm record in 4.1 and the need for a Rust crypto core changed my answer. |
| Claude Code for builds | Withdrawn | No paid tools (D-006). |
| Redis for rate limiting and fan-out | Deferred; Valkey if needed | Licence history in 4.7; most of M0 to M4 does not need it. |

## 4. Component analysis

### 4.1 Backend: Go

**Verified facts**

- [V, P] Go ships a major release every six months and supports only the two most recent. Go 1.27 (August 2026) is current; Go 1.26 (February 2026) is still supported. [V, S] Several 1.26 patch releases in 2026 carried security fixes.
- [V, P] The `go` command authenticates downloaded modules against a public, append-only checksum database (a transparency log), on by default since Go 1.13.
- [V, S] September 2025: a phished maintainer account led to malicious releases of about 18 very widely used npm packages, including chalk and debug. A self-propagating worm (Shai-Hulud) abused install-time lifecycle hooks and compromised 500 or more packages; a second wave followed in November 2025.
- [V, P] Google reports Android's memory-safety vulnerabilities fell below 20% of the total in 2025, with Rust code showing about a thousand times lower memory-safety vulnerability density than C and C++, a roughly four times lower rollback rate, and about 25% less review time.

**Comparison (opinion [O]; items marked [U] are unverified)**

| Language | Memory safety | Supply-chain posture | Concurrency and realtime | Iteration speed in a chat-driven loop | Footprint | Verdict |
|---|---|---|---|---|---|---|
| Go | Safe, except that data races can break it; the race detector in CI catches them | Strong: checksum database; typical dependency trees are small | Strong (goroutines) | Fast builds, clear errors | Small static binary | **Recommended** |
| Rust | Safe by construction | Good tooling (advisory database, cargo-audit [U]); large dependency trees are common | Strong | Slower: borrow checker, compile times | Small | Use where decisive: crypto core, parsers. Not the whole backend |
| TypeScript on Node | Safe | Weakest of the set, per the 2025 incidents | Adequate; single-threaded, workers for CPU work | Fast | Medium | Client language. Rejected for the backend |
| Kotlin or Java | Safe | Good [U] | Strong, virtual threads [U] | Moderate | Heavy | Credible alternative; most common hiring pool in India [U] |
| C# on .NET | Safe | Good [U] | Strong | Moderate | Medium to heavy | Credible alternative; no tooling overlap with our stack |
| Python | Safe | Mixed [U] | Weak for realtime [U] | Fast | Medium | Rejected |
| Elixir | Safe | Small ecosystem | Excellent for realtime | Moderate | Medium | Rejected: very small talent pool |

**Why Go [O]**

1. A small language with a large standard library is the easiest thing for another senior team to read in days, which is what "inheritable" means in practice.
2. The component that holds all institutional and student data should have the smallest supply-chain blast radius we can get, and Go's module authentication plus smaller dependency trees give that. The client still uses npm, so section 4.2 adds a strict client policy.
3. Static binaries in distroless images run as non-root with a read-only filesystem (the M0.1 skeleton does this).
4. Goroutines suit the WebSocket gateway and background workers.
5. The memory-safety gain from Rust over Go is real but much smaller than Rust over C or C++, which is what the Google data compare. The cost in velocity and contributor pool is large for us.

**Honest weaknesses.** No sum types, so state machines need discipline and an exhaustiveness linter [U]. Data races can break memory safety, so `-race` stays in the quality gate. Error handling is verbose. Go has no long-term-support line, so we must move to a new minor release at least yearly.

**Where memory-unsafe code remains.** Image and video parsing (libvips, ffmpeg) is classic exploit territory. It runs in isolated worker containers with no network, tight resource limits and re-encoding of every output (4.8).

**Reversibility.** The skeleton is about 850 lines of Go, tests included. If you or ChatGPT strongly prefer TypeScript or Kotlin, the cost to switch is low before M0.2 and rises steeply afterwards.

### 4.2 Web client: TypeScript, React, Vite, PWA

**Opinion [O].** React has the largest ecosystem of accessible component primitives (Radix or React Aria [U]) and the largest pool of people who can inherit it. Svelte, Vue and Solid are good frameworks with smaller accessibility ecosystems. Next.js adds a Node server runtime and framework coupling that an authenticated single-page app does not need. TanStack Query for server state; Vitest and Playwright for tests [U].

**Client dependency policy** (motivated by the verified npm incidents in 4.1):

- Commit the lockfile. Install with `npm ci` or the frozen-lockfile equivalent. Disable dependency lifecycle scripts by default [U: confirm the exact setting for the chosen package manager at setup].
- Pin exact versions. Hold new versions back 14 days. Review every added dependency.
- Build only in CI. Strict Content Security Policy. No third-party scripts on the app origin. Subresource integrity for anything hosted elsewhere.

### 4.3 Private messaging cryptography

**What D-005 requires.** Encryption on the sender's device; only recipients decrypt; no master key; no administrator or database-operator access; group membership changes drive key changes; removed members cannot read later messages; voluntary reporting; a staged plan instead of any server-readable fallback.

**Protocol: MLS [O], on verified facts**

- [V, P] RFC 9420 (July 2023) defines MLS for groups from two to thousands of members with forward secrecy and post-compromise security. The RFC also notes that sender-key designs struggle to give post-compromise security cheaply.
- [V, P] RFC 9750 (April 2025) describes the architecture, including the Delivery Service and Authentication Service roles we will implement.
- [O] Use MLS for every private conversation: one protocol, one implementation, one audit surface. A one-to-one chat is a two-member group. Adding and removing members are MLS commits that advance the group's epoch. Spike S-1 includes an automated test that a removed member cannot decrypt anything sent after removal.

**Library candidates**

| Candidate | Licence | Browser story | Audit and maturity | Verdict |
|---|---|---|---|---|
| OpenMLS (Rust) | MIT [V, S: confirm the LICENSE file] | CI builds the wasm32 target but does not test it [V, P]. A third-party project ships it compiled to WASM in browser clients [V, S] | Independent security audit by SRLabs, funded by the Sovereign Tech Agency: eight findings, one rated High, fixes released in 0.8.1 and 0.7.3 [V, S]. Reported in use by Wire, XMTP and others [V, S] | **Primary** |
| mls-rs (Rust, AWS Labs) | Apache-2.0 or MIT [V, P] | WASM builds supported; Web Crypto and Rust Crypto providers marked experimental [V, P] | Validated for RFC 9420 conformance, but the README says it has not had a full third-party audit [V, P] | **Fallback** |
| libsignal | AGPLv3; the README says use outside Signal is unsupported [V, P] | No browser build [U] | n/a | Rejected |
| Olm and Megolm (vodozemac) | Apache-2.0 [U] | WASM exists [U] | Group ratchet has weaker post-compromise properties and Matrix-specific semantics [O] | Rejected |
| ts-mls (TypeScript) | [U] | Native JavaScript | [U] | Watch only; not evaluated |
| Custom protocol | | | | Rejected by D-008 |

**Architecture sketch [O]**

- **Client crypto core.** A small Rust crate (`brynn-crypto-core`) wraps OpenMLS and exposes about ten operations (create device identity, make key packages, create group, add, remove, encrypt, decrypt, export state, import state). It runs in a dedicated Web Worker. State lives in IndexedDB, encrypted under a non-extractable WebCrypto key. A shared-computer mode keeps nothing on disk and therefore keeps no history.
- **Authentication Service.** Our identity service issues device credentials bound to a person and their age band.
- **Delivery Service.** A blind store-and-forward service in Go. It stores key packages, welcome messages, commits and application messages as opaque blobs, gives each conversation a monotonic sequence number, orders commits per epoch inside a database transaction, and delivers only to current members. It holds no keys.
- **Metadata the server necessarily sees:** who is in which conversation, when messages are sent, their sizes, device identifiers, and IP addresses at connection time. Minimize it and set retention.
- **No-admin-read proof.** There is no API and no database column that can hold plaintext or keys. A canary test sends known plaintext through the whole system and then searches the database, logs and object storage for it. The build fails if it is found.
- **Attachments in private chats** are encrypted on the device before upload, with the key carried inside the message. The server stores opaque blobs.

**Staged plan**

| Stage | Scope | Exit criteria |
|---|---|---|
| E0 | Spike S-1. Server seams: device registry, key package store, ciphertext-only envelope with a suite and version field. Threat model. Canary harness | Spike evidence recorded; schema reviewed by ChatGPT |
| E1 | One-to-one chats. One active device per account. Keys in IndexedDB. Shared-computer mode | Two-client end-to-end tests pass; canary passes; CSP and integrity measures in place |
| E2 | Private groups. Add and remove members. Server-ordered commits | Automated removal test; concurrency tests on commit ordering |
| E3 | Multiple devices (QR linking), device revocation, safety numbers | Add and revoke tests; verification screen tried with real students |
| E4 | Encrypted backup and recovery with a user-held recovery key. "No recovery key, no history" stated plainly | Restore drill; user comprehension check |
| E5 | Verifiable reporting, key transparency, post-quantum suites when stable | External cryptographer review (gate G-2) |

Messages ships to minors only after E2 and gate G-2. Until E3, one device per account is a real usability limit for students who use a phone and a school computer, and we will say so.

**Limits we state openly [O]**

1. The server sees metadata (above).
2. In a web client the operator delivers the code. [V, S] Meta built the open-source Code Verify extension precisely because a web app can be served different code than expected. Until we have strict CSP, subresource integrity, signed bundles and ideally a native wrapper, a malicious or compromised operator could in principle ship a client that leaks keys (R-005).
3. Until safety numbers (E3) or key transparency (E5) exist, an operator who controls the key directory could in principle substitute keys at enrollment.
4. End-to-end encryption does not stop recipients forwarding or screenshotting, and does not protect a compromised device.
5. There is no server-side scanning of private content, including for child sexual abuse material or grooming. Compensating controls are structural: age-band and institution-scoped contact rules, no contact from strangers, rate limits on new contacts, blocking and reporting, and a tested safeguarding workflow. Legal duties are open question L-004.
6. A lost device loses its history until E4.
7. Suspending an account stops delivery at once, but cryptographic removal from a group needs a commit from a remaining member's device. P1-04 will define who issues it and how fast.
8. Message search is on-device only. Push notifications carry no message content. Private attachments cannot be scanned or thumbnailed by the server.

**Reporting [V, S]** Message franking was introduced by Facebook for end-to-end encrypted Messenger and formalized by Grubbs, Lu and Ristenpart (CRYPTO 2017); a 2025 paper extends the idea to franking transcripts for groups. Franking composes a new construction on top of the messaging protocol, which is exactly where we should not improvise. **Proposal [O]:** in E1 to E4, reports are user-submitted selected messages plus metadata, which moderators treat as claims. Verifiable reporting waits for E5 and external review.

### 4.4 Database: PostgreSQL 18

- [V, P] The project supports each major version for five years. PostgreSQL 18 was released 25 September 2025 and is supported until 14 November 2030. PostgreSQL 19 was in beta as of the August 2026 release notes I found; check whether it is generally available before final selection.
- [O] Row-level security as a backstop to scoped queries, strong constraints, transactional DDL for safe expand-and-contract migrations, and a very large operational knowledge base. Start on 18; revisit once 19 has several minor releases.
- [O] Rejected: MySQL or MariaDB (weaker constraint and RLS story), MongoDB (our integrity needs are relational), distributed SQL such as CockroachDB or YugabyteDB (scale and licence terms we do not need [U]), SQLite (single node, awkward with rolling deploys).
- [U] Decide at M0.2, with licence and maintenance checks: migration tool (goose, golang-migrate, Atlas), access layer (pgx with sqlc), backup tooling (pgBackRest or WAL-G) and restore drills.

### 4.5 Identity: Ory Kratos, with a fallback

- [V, S] Keycloak is Apache-2.0 and a CNCF incubating project, but it is a heavy Java service. Ory Kratos and Hydra are Apache-2.0, with some features under an enterprise licence. Zitadel moved to AGPL-3.0 with version 3 (one source still lists Apache-2.0, so treat the older licence as outdated).
- [O] Authentication and recovery flows are where hand-written code most often fails, and recovery is critical for us: with end-to-end encryption, an account takeover could enroll an attacker's device. A mature headless identity service reduces that risk. Everything specific to BRYNN (person, age band, guardian links, memberships, devices) stays in our code, and date of birth never goes into Kratos.
- [O] Keycloak rejected for weight and tenancy model. Zitadel rejected for AGPL unless we choose to open-source BRYNN (O-003).
- Spike S-2 decides. If Kratos cannot do invite-only registration, age gating and device-bound enrollment cleanly, fall back to an in-house module using argon2id, a vetted WebAuthn library, server-side sessions and an ASVS Level 2 checklist.

### 4.6 Background jobs and the transactional outbox: River

- [V, S] River is a Go job queue on Postgres under MPL-2.0 that supports enqueuing jobs inside the same transaction as your data changes: a job exists only if the transaction commits.
- [O] That property is exactly what an outbox needs. A hand-built `SKIP LOCKED` queue is more code to get right. Redis queues lose same-transaction enqueue. Kafka and Temporal are heavy for our scale.

### 4.7 Real-time and cache

- [O] WebSocket gateway in the Go realtime role. Choose the library at M6 after checking maintenance [U]. Start with Postgres LISTEN and NOTIFY for cross-instance signals; the database is the source of truth, not the socket.
- [V, S] Redis changed to RSALv2 and SSPLv1 in March 2024 and added an AGPLv3 option with Redis 8 in May 2025. Valkey is the Linux Foundation fork under BSD-3-Clause, backed by AWS, Google and Oracle.
- [O] Until M5, rate limiting uses the reverse proxy and in-process limits. When several instances need shared limits or fan-out, adopt Valkey.

### 4.8 Object storage and media

- [V, S] MinIO's community edition stopped publishing prebuilt images in October 2025, entered maintenance mode on 3 December 2025, was declared no longer maintained in February 2026, and the repository was archived in 2026. Alternatives reported: SeaweedFS (Apache-2.0), Garage (AGPLv3), RustFS (Apache-2.0), Ceph (heavier).
- [O] Code against the S3 API behind a Go interface. Use SeaweedFS in development. Choose production storage at M2 on cost, India region and backup features.
- [O] Media workers: isolated containers, no network, CPU and memory limits, re-encode every output, strip location metadata (EXIF GPS) from images, and malware-scan public uploads (ClamAV [U]). Private encrypted attachments cannot be scanned server-side.

### 4.9 Deployment, observability, CI and security tooling

- [O] Docker Compose and Caddy for automatic HTTPS [U: confirm Caddy's licence and current version]. Zero-downtime application deploys on one VPS: two API replicas behind the proxy, draining through `/readyz` (implemented in M0.1) before stopping. The single Postgres node remains a single point of failure, so backups and restore drills matter more than redundancy at this stage. No Kubernetes: the operating cost does not buy anything we need yet.
- [O] Observability: structured JSON logs now, then OpenTelemetry, Prometheus, Grafana and Loki, and a self-hosted open-source error tracker with personal data scrubbed. Never log message content, tokens or personal data.
- [O] CI: GitHub Actions on the free tier [U], but every check lives in `scripts/` so CI can move (Forgejo, GitLab). Planned tools [U]: govulncheck, gitleaks, Trivy, OSV-Scanner, Renovate, Semgrep, Playwright, OWASP ZAP, k6.
- [V, S] OWASP ASVS 5.0.0 was released on 30 May 2025 with Level 1 to 3 requirements. [O] Target Level 2 for the platform and Level 3 for identity and cryptography components, and use STRIDE for each threat model.

## 5. Licensing summary

| Component | Licence | Implication |
|---|---|---|
| OpenMLS | MIT [V, S] | Permissive |
| mls-rs | Apache-2.0 or MIT [V, P] | Permissive |
| libsignal | AGPLv3 [V, P] | Rejected |
| River | MPL-2.0 [V, S] | File-level copyleft; fine as an unmodified dependency |
| Ory Kratos | Apache-2.0, some enterprise features [V, S] | Fine for the core features |
| Keycloak | Apache-2.0 [V, S] | Not selected |
| Zitadel | AGPL-3.0 since v3 [V, S] | Avoid unless BRYNN is open source |
| Redis 8 | AGPLv3 option [V, S] | Avoid; use Valkey (BSD-3-Clause) |
| MinIO | AGPL; community edition archived [V, S] | Rejected |
| SeaweedFS, RustFS | Apache-2.0 [V, S] | Candidates |
| Garage | AGPLv3 [V, S] | Possible as a separate service |
| Go, PostgreSQL, React, Vite, Grafana stack | BSD-style, PostgreSQL licence, MIT, AGPL for Grafana and Loki [U] | Run AGPL services separately and unmodified |

Whether BRYNN's own code is open or proprietary (O-003) affects AGPL components. Using unmodified AGPL software as a separate service generally does not extend obligations to our code, but counsel should confirm [O].

## 6. Infrastructure and cost

**Assumptions [E]:** one pilot institution, 1,000 to 3,000 registered users, at most 500 concurrent, at most 100 GB of media in year one.

**Requirements [O]:** one VPS with 2 vCPU and 4 GB RAM can run two API replicas, Postgres, Caddy and light monitoring. A second VPS for staging is better but optional at the start. Object storage for media and backups, an email provider, a domain. Push services have free tiers [U].

| Item | Monthly estimate |
|---|---|
| App VPS, 2 vCPU and 4 GB, India region | About US$20 to 24 [V, S: aggregator listings for Vultr and DigitalOcean, which have India regions; smaller plans start near US$6] |
| Staging VPS | US$0 to 24 [E] |
| Object storage and backups | US$1 to 10 [E] |
| Domain | About US$1 [E] |
| Transactional email | US$0 to 15 [E] (free tiers often cover pilot volume [U]) |
| Monitoring, CI, push | US$0 [U] |
| **Infrastructure total** | **About US$25 to 60 per month [E]** |

**The real costs are not infrastructure [E].** Qualified counsel (gate G-1) and an independent security and cryptography review (gate G-2) are likely to cost far more than a year of hosting, and neither is budgeted yet (KI-005). That is the item to solve early.

## 7. Spikes and verification tasks

- **S-1 OpenMLS in the browser (3 working days).** On Chrome, Firefox, Safari and a low-end Android phone: create an identity and key package; create a two-member then three-member group; send and receive; add and remove a member; prove a removed member cannot decrypt later messages (automated test); persist and reload state from IndexedDB. Record WASM size, cold-start time and memory. Read the published audit report in full. Output: go or no-go, with mls-rs as the fallback.
- **S-2 Ory Kratos fit.** Invite-only registration, age gate, password plus passkey, recovery, session validation from Go, resource use, licence check. Output: go or no-go, with the in-house fallback.
- **S-3 Two-replica drain drill (M0.5).** Roll the API with live traffic and show zero failed requests.
- **T-tasks** in the ledger (T-008 to T-010) cover the migration tool, access layer, SHA and digest pinning, and vulnerability scanning.

## 8. What would change these decisions

- S-1 fails or shows unacceptable size or battery cost on low-end phones: switch to mls-rs, or defer web Messages and ship a native wrapper first.
- Counsel advises that end-to-end encryption for minors is untenable in India: Messages becomes adults-only until a lawful design exists. D-005 is yours to revisit with that advice.
- You prefer Kotlin or TypeScript, or ChatGPT finds a concrete blocker in Go: change before M0.2.
- The MLS ecosystem or the audit findings move: re-run section 4.3.
- A pilot institution requires SAML or district SSO: add Keycloak or Kratos enterprise features as a broker.

## 9. Sources (accessed 2026-11-19; URLs as returned by search)

**Cryptography**

- RFC 9420, The Messaging Layer Security (MLS) Protocol: https://www.rfc-editor.org/rfc/rfc9420.html (P)
- RFC 9750, MLS Architecture: https://www.rfc-editor.org/rfc/rfc9750.html (P)
- OpenMLS repository: https://github.com/openmls/openmls (P); documentation for 0.8.0: https://docs.rs/openmls/=0.8.0/ (P)
- OpenMLS audit coverage: https://heise.de/-11316951 (S)
- mls-rs repository and docs: https://github.com/awslabs/mls-rs and https://awslabs.github.io/mls-rs (P)
- libsignal repository: https://github.com/signalapp/libsignal (P)
- Code Verify listing and coverage: https://apps.apple.com/no/app/code-verify/id6475794471 (S); https://www.ghacks.net/2022/03/11/meta-releases-code-verify-extension-for-whatsapp-web/ (S)
- Message franking: https://eprint.iacr.org/2017/664 (P); transcript franking: https://arxiv.org/html/2507.19391v1 (P)

**Law (for counsel to confirm; not legal advice)**

- DPDP Rules 2025 timeline: https://www.mirvolegal.com/insights/dpdp-rules-2025-compliance-timeline/ (S)
- Verifiable parental consent and children's data: https://www.sansalegal.com/post/how-to-design-a-verifiable-parental-consent-process-under-the-dpdp-rules-2025 (S); https://www.mondaq.com/india/privacy-protection/1833814/parental-consent-requirements-for-businesses-collecting-data-from-minors (S); https://dpdpa.dcomply.in/rules/ (S)
- IT Rules 2021 traceability and thresholds: https://sflc.in/what-originator-or-traceability-provision-information-technology-intermediary-guidelines-and/ (S); https://ssrana.in/?p=16931 (S)
- POCSO reporting duties of intermediaries: https://deccanherald.com/india/social-media-intermediaries-cant-claim-protection-if-no-due-diligence-observed-supreme-court-3203175 (S); https://newsable.asianetnews.com/india/sc-seeks-centre-s-response-on-plea-for-stronger-cseam-reporting-rules-articleshow-qls7z8t.html (S)

**Platform and tooling**

- Go 1.27 release notes: https://go.dev/doc/go1.27 (P); Go version guide: https://stevenpg.com/references/go-version-guide/ (S)
- Go checksum database design: https://go.googlesource.com/proposal/+/refs/heads/master/design/25530-sumdb.md (P); https://security.googleblog.com/2023/06/supply-chain-security-for-go-part-2.html (P)
- PostgreSQL versioning policy: https://www.postgresql.org/support/versioning/ (P)
- River: https://pkg.go.dev/github.com/riverqueue/river and https://github.com/riverqueue/river (P)
- npm incidents: https://socura.co.uk/threat-alerts/supply-chain-attack-npm-debug-and-chalk-packages (S); https://versprite.com/resources/security/npm-supply-chain-attack-shai-hulud-worm-compromises-500-packages/ (S); https://www.fortiguard.com/threat-signal-report/6201/npm-shai-hulud-supply-chain-attack (S)
- Memory safety in Android: https://blog.google/security/rust-in-android-move-fast-fix-things/ (P)
- MinIO status and alternatives: https://pinggy.io/blog/minio_archived_self_hosted_s3_alternatives/ (S); https://bex.co/blog/2026/07/09/minio-death-garage-seaweedfs-rustfs (S); https://www.gigenet.com/blog/minio-alternatives/ (S)
- Redis and Valkey licensing: https://analyticsindiamag.com/ai-features/redis-doesnt-want-developers-to-hate-it-anymore (S); https://basekv.pages.dev/articles/redis-agpl-license-explained (S)
- Identity providers: https://wz-it.com/en/blog/keycloak-alternatives/ (S); https://www.authgear.com/post/best-open-source-identity-providers (S); https://skycloak.io/?p=3288 (S)
- OWASP ASVS 5.0: https://www.appknox.com/cyber-security-jargons/owasp-asvs (S); https://quality.arc42.org/standards/owasp-asvs (S)
- India VPS pricing: https://valebyte.com/en/vps/india/ (S); https://apicalculators.com/compare/vultr-vs-digitalocean (S)
