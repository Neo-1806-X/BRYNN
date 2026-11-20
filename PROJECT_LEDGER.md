# BRYNN — Project Ledger

**The single source of truth for decisions, status, risks and open questions.**
Paste this file at the start of every new chat (Claude or ChatGPT). Update it at the end of every working session.

| | |
|---|---|
| Ledger version | 1 |
| Last updated | 2026-11-19 |
| Phase | Phase 1 (architecture pack) and Milestone M0.1 (engineering baseline) |
| Next action (owner) | Run Milestone M0.1 ([steps](docs/milestones/M0.1-foundation.md)) and paste the real output. Start the long-lead human items: counsel (L-001 to L-007) and the pilot institution contact (O-004, O-005). |
| Next action (lead) | Deliver P1-12 (legal and safeguarding brief for counsel), then P1-03 (identity, age band, membership schema). |

## 1. How we work

| Role | Who | Responsibility |
|---|---|---|
| Owner and operator | The owner | Product and legal decisions. Runs commands and pastes complete output. Holds accounts and credentials. |
| Lead engineer and manager | Claude | Architecture, task breakdown, code, documentation, triage of reviews, honest status. |
| Independent reviewer | ChatGPT | Challenges designs and code, finds bugs and security issues, gives second opinions. |

Rules:

1. Nobody claims to have run code, installed dependencies, executed tests or deployed anything unless they did. Every "done" needs evidence (command and output) in section 8.
2. Claude delivers work as: file-by-file plan, exact repository paths, complete files (or a zip), commands, verification steps.
3. The owner pastes full terminal output, not summaries. Claude diagnoses and gives the next actionable step.
4. Reviewer findings are inputs, not orders. Claude triages each as Accepted, Rejected or Deferred with a reason (section 9). Product and legal decisions belong to the owner.
5. No paid dependency or tool without first explaining alternatives and cost. Nothing depends on Claude Code CLI or any paid agent.
6. Neither AI replaces qualified counsel or an independent security and cryptography review. Those are launch gates (section 6).
7. Ambiguity is never resolved silently: it becomes an entry in section 5.
8. Dates and versions of third-party software go stale. Anything marked "verify" is re-checked before it is relied on.

## 2. Status

| Track | Item | State |
|---|---|---|
| Phase 1 pack | P1-00 revision map | Delivered (drop 001) |
| Phase 1 pack | P1-01 technology selection report v0.9 | Delivered (drop 001), awaiting owner and ChatGPT review |
| Phase 1 pack | P1-12 legal and safeguarding brief | Next |
| Phase 1 pack | P1-03 identity, age band, membership | After P1-12 |
| Phase 1 pack | P1-07, P1-09, P1-04, P1-08, P1-10, P1-11, P1-02, P1-05, P1-06 | Queued in this order (see revision map) |
| Build | M0.1 engineering baseline and walking skeleton | Delivered (drop 001), NOT yet verified by the owner |
| Build | M0.2 database foundation (Postgres 18, migrations, tenancy and RLS skeleton) | Planned |
| Build | M0.3 identity (Kratos spike S-2, person, membership, age band) | Planned |
| Build | M0.4 authorization matrix, audit log, outbox and jobs | Planned |
| Build | M0.5 CI hardening, SBOM, vulnerability scanning, two-replica deploy drill | Planned |
| Spike | S-1 OpenMLS in the browser (WASM) | Planned, before any Messages code |
| Spike | S-2 Ory Kratos fit | Planned, M0.3 |

## 3. Confirmed decisions (made by the owner)

| ID | Decision |
|---|---|
| D-001 | Engineering first: foundation, security and data integrity before products. Never build eleven products at once. Each milestone is verified before the next starts. |
| D-002 | Age policy, subject to legal and safeguarding review: under 14 has no access to XENON or Whisper; 14 to 17 get age-appropriate versions with extra safety restrictions; 18 and over get adult functionality within platform rules. Most users are school students, mostly under 18. |
| D-003 | XENON is interest and community discovery for students, not a dating product: no romantic or sexual positioning; strict separation of minors and adults in discovery and matching; no unsolicited contact from strangers; discoverability privacy controls; blocking and reporting; abuse prevention and rate limits; no unnecessary sensitive data; no free text that creates avoidable safety risk. |
| D-004 | Whisper is anonymous to the recipient. It needs abuse prevention, rate limits, protection against harassment, threats, doxxing and exploitation, clear reporting, and no routine administrator access to submissions. Any exceptional identity investigation follows a documented, legally appropriate, restricted and auditable process. We never promise absolute anonymity, and the privacy model is explained plainly to users. |
| D-005 | Private conversations (one to one, private groups, private XENON chats after mutual connection) use genuine end-to-end encryption. Administrators and database operators cannot read them. No master key, no hidden admin reading. Group membership changes drive key management and removed members cannot decrypt later messages. Users may voluntarily report selected content. If we cannot do this securely with our resources we state the limit and stage it. We never silently substitute server-readable messaging. |
| D-006 | No dependency on Claude Code CLI or any paid tool. Development happens through chat; ChatGPT is the second reviewer. The repository stays portable. Paid dependencies need an explanation of alternatives and cost first. |
| D-007 | Claude is lead engineer and manager. The owner and ChatGPT follow that direction. |
| D-008 | Technology is researched before it is chosen, using current primary sources. Production-grade standard. No custom cryptography. Memory-safe languages where memory safety matters. No single language mandated everywhere. A technology selection report precedes implementation. |
| D-009 | India's data protection and child-safety requirements are investigated before XENON and Whisper are enabled. A minimum age of 14 does not by itself make processing lawful, and institutional consent does not by itself resolve every obligation. |

## 4. Proposed decisions (made by the lead; veto window open)

The owner or ChatGPT may object to any of these. Language and framework decisions (P-002, P-003) are treated as accepted when the owner runs the first `make check` in M0.1 step 5 without objection; the skeleton is small, so changing before then is cheap.

| ID | Proposal | Detail |
|---|---|---|
| P-001 | Modular monolith, one repository, three runtime roles (api, realtime, worker) built from one codebase; one shared web shell; BRYNN Control as a separate admin app on its own origin | From the previous review; not objected to. |
| P-002 | Backend language: Go | Supersedes my earlier TypeScript default. Reasons in the technology report section 4.1. |
| P-003 | Web client: TypeScript, React, Vite, installable PWA | Report section 4.2. |
| P-004 | Private messaging: MLS (RFC 9420) through OpenMLS compiled to WASM; mls-rs is the fallback; libsignal rejected; no custom crypto; one-to-one chats are two-member MLS groups | Report section 4.3. Subject to spike S-1. |
| P-005 | PostgreSQL 18 with row-level security; global `person` record plus per-institution `membership`; date of birth held only in the identity module and exposed to other modules as `age_band` | Detail in P1-03. |
| P-006 | Identity: Ory Kratos for credentials, sessions and recovery, with domain logic (person, age band, guardian links, memberships, devices) in our own code. Fallback: in-house identity on audited libraries | Subject to spike S-2. |
| P-007 | Background jobs and the transactional outbox on Postgres using River | Report section 4.6. |
| P-008 | No Redis or Valkey until needed (M5 or later), then Valkey. Object storage behind an S3-API interface; not MinIO | Report sections 4.7, 4.8. |
| P-009 | Deployment: Docker Compose and a reverse proxy on an India-region VPS; no Kubernetes | Report section 4.9. |
| P-010 | Conversation classes: Private (end-to-end encrypted, not readable by anyone but members) and Community (server-readable, moderated, labelled "not private") | Needs owner decision O-001. |
| P-011 | No anonymous mode inside Messages. Anonymity exists only in Whisper and XENON. A XENON match moves into Messages only by mutual consent | From the previous review. |
| P-012 | Version 1 feed ranking is rule-based (follows, house, recency, institutional priority), with no engagement profiling and no advertising | Pending counsel (L-003). |
| P-013 | Whisper content is encrypted at rest with application-level keys and has no administrator browsing interface. Sender identity is kept in a separate restricted store, released only through a break-glass process needing two authorized people, with an immutable audit entry. The privacy notice states these limits | Detail in P1-06. Pending counsel (L-005). |
| P-014 | Reporting under end-to-end encryption in early stages is user-submitted selected messages with metadata, treated by moderators as claims. Cryptographically verifiable reporting (franking) is deferred to stage E5 and needs external cryptographer review | Report section 4.3. |

## 5. Open questions

### Owner decisions

| ID | Question | My default | Needed by |
|---|---|---|---|
| O-001 | Does Messages include institution-wide Community channels that are not end-to-end encrypted (moderated, labelled not private)? | Yes, strictly separated from Private conversations (P-010) | Before M6 |
| O-002 | What can under-14 users access (Pulse, Stories, Messages)? | Only institution-managed spaces; no private messaging | Before M2 |
| O-003 | Will the BRYNN code be open source or proprietary? | Undecided. It affects which AGPL components we can combine | Before the first external contributor |
| O-004 | Who is the legal entity operating BRYNN, and who is the Designated Safeguarding Lead at the pilot institution? | None assumed | Before the pilot |
| O-005 | Which institution pilots first, and by what date? | None assumed. Note the May 2027 DPDP date (L-002) | Before M2 |
| O-006 | After a XENON mutual connection, are identities revealed or does the private chat stay pseudonymous? | Pseudonymous until both people choose to reveal | Before P1-05 |

### Legal (for qualified Indian counsel)

| ID | Question |
|---|---|
| L-001 | Is the institution or BRYNN the data fiduciary for each product, and who may rely on the school carve-out for children's data? |
| L-002 | Which verifiable parental consent mechanism satisfies Rule 10, and can the institution give consent for students? Most data fiduciary duties, including children's data, apply from about 13 to 14 May 2027. |
| L-003 | Does engagement-based ranking or automated safety scanning count as behavioural monitoring of children? |
| L-004 | POCSO sections 19 to 21 reporting workflow, evidence handling and the designated officer when private chats are end-to-end encrypted. |
| L-005 | Legal basis, process and retention for Whisper identity investigations. |
| L-006 | IT Rules 2021 intermediary duties at our size (grievance officer, takedown timelines). Rule 6 lets government extend significant-intermediary duties, including first-originator tracing, to other intermediaries; what does that mean for end-to-end encryption? |
| L-007 | The minimum age of 14 has no basis in the DPDP definition of a child (under 18). Confirm the policy is lawful with verifiable parental consent. |

### Technical

| ID | Question | Where |
|---|---|---|
| T-001 | Does OpenMLS run well enough in browsers (bundle size, persistence, Safari, low-end Android)? Read the published audit report in full | Spike S-1 |
| T-002 | Does Ory Kratos fit invite-only registration, age gating and device-bound keys? | Spike S-2 |
| T-003 | Verifiable reporting design: franking, MLS signature disclosure, or report-as-claim | Stage E5 |
| T-004 | Multi-device, backup and recovery behaviour on shared school computers | Stage E3, E4 |
| T-005 | Web client integrity: strict CSP, subresource integrity, signed bundles, Code Verify style transparency, or a native wrapper | Stage E1 |
| T-006 | Content-free push notifications and web push limits on iOS | M5 |
| T-007 | Encrypted attachments: client-side thumbnails and checks, since the server cannot scan them | M6 |
| T-008 | Choice of migration tool and database access library, with licence and maintenance checks | M0.2 |
| T-009 | Pin GitHub Actions by commit SHA and base images by digest | M0.5 |
| T-010 | Wire in govulncheck with a verified version pin | M0.2 |

## 6. Launch gates

| Gate | Requirement | Before |
|---|---|---|
| G-1 | Counsel sign-off on DPDP (including Rule 10 consent), POCSO reporting, IT Rules and the age policy | Any real student data |
| G-2 | Independent review of our end-to-end encryption integration and of identity and session design | Minors can use Messages |
| G-3 | Safeguarding policy, a named Designated Safeguarding Lead, and a tested report-and-escalate workflow | Whisper or XENON is enabled |
| G-4 | Backup restore drill passed and recorded | Production data |
| G-5 | OWASP ASVS 5.0 Level 2 self-assessment with no open High findings | Pilot |
| G-6 | Threat model reviewed by the second reviewer | M6 |

## 7. Known issues and risks

| ID | Severity | Description | Mitigation |
|---|---|---|---|
| KI-001 | Medium | Drop 001 Go code, Dockerfile, Compose file and CI workflow were not compiled or run by the author (no Go or Docker in the author's sandbox). The owner's first `make check` is their first real test. | Owner pastes output. Claude fixes on report. |
| KI-002 | Low | CI actions are pinned to major tags, not commit SHAs. | T-009 at M0.5. |
| KI-003 | Low | Base images are pinned by tag, not digest. | T-009 at M0.5. |
| KI-004 | Info | Vulnerability scanning is not wired in yet. | T-010 at M0.2. |
| KI-005 | High (process) | No budget yet for counsel or an independent security and cryptography review. Both are launch gates. | Raise early. Explore grants, university groups, and pro bono options. |
| R-001 | High | Our integration of MLS is unaudited even though OpenMLS itself was independently audited. | Keep our cryptographic surface to plumbing, publish the design, gate G-2. |
| R-002 | High | India's DPDP children's data duties apply from about May 2027. | Build consent and age-band foundations in M0.3 rather than later. |
| R-003 | Medium | Single human operator (bus factor of one). | This ledger, ADRs, portable repo, scripted setup. |
| R-004 | Medium | AI-written code can be confidently wrong. | Tests and CI, second reviewer, evidence rule, independent review of high-risk parts. |
| R-005 | Medium | A malicious or compromised operator controls code delivery to browsers and the key directory, so end-to-end encryption in a web client has a residual trust gap. | Strict CSP and SRI, signed bundles, key verification UX, transparency, native wrapper later. Disclosed to users. |

## 8. Evidence log

| Session | Check | Run by | Result |
|---|---|---|---|
| S-001 | YAML parse of `docker-compose.yml` and `.github/workflows/ci.yml` (PyYAML) | Claude | Pass |
| S-001 | `make -n check`, `make -n up`, `make -n smoke`, `make help` | Claude | Pass (dry run; Docker not available there) |
| S-001 | `sh -n scripts/check-go.sh` | Claude | Pass |
| S-001 | Go bracket balance and duplicate declaration scan; crude import usage scan | Claude | Pass. This is NOT a compile. |
| S-001 | Relative links across the 6 Markdown files resolve | Claude | Pass |
| S-001 | `go build`, `go vet`, `go test`, `docker build`, `docker compose up`, CI | Nobody yet | Not run |
| S-001 | Web research for the technology report | Claude | Sources and access date listed in the report |

## 9. Review log (ChatGPT findings and my triage)

| Date | Finding | Triage | Reason |
|---|---|---|---|
| | (none yet) | | |

## 10. Session log

- **S-001, 2026-11-19.** Owner set product policy (age, XENON, Whisper, end-to-end encrypted Messages), made no-paid-tools and chat-workflow decisions, and made Claude lead. Claude researched technology and law, delivered drop 001 (revision map, technology report v0.9, ledger, M0.1 walking skeleton) and withdrew two earlier recommendations: server-readable messages (replaced by D-005) and TypeScript everywhere (replaced by P-002).

## Appendix A. Session start prompt

> Read PROJECT_LEDGER.md (pasted below). You are working on BRYNN. Follow section 1 (How we work). Continue from "Next action". Do not claim to have run anything you did not run. Ask me for command output when you need evidence.

## Appendix B. Review request for ChatGPT

> You are the independent reviewer for BRYNN. The lead engineer is Claude. Read PROJECT_LEDGER.md and the attached files.
> Review for correctness, security (OWASP ASVS 5.0), data integrity, privacy (DPDP, minors), maintainability, test adequacy, and conflicts with confirmed decisions D-001 to D-009.
> Output a table: ID, Severity (Critical, High, Medium, Low, Info), File and line or section, Evidence (quote or reproduction), Recommended fix, Confidence (High, Medium, Low).
> Rules: do not invent facts. Cite primary sources for claims about libraries, laws or standards. Say what you could not verify. Do not rewrite files unless asked.
