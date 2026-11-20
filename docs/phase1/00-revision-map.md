# P1-00 Revision map

Status: delivered in drop 001 (2026-11-19). Author: Claude (lead engineer).

The owner's policy decisions (D-002 to D-009 in the ledger) change the earlier material. This map says exactly what must be revised, where the revision will live, and in what order the revised Phase 1 pack arrives.

## 1. Which existing documents need revision

| # | Existing document | Sections to revise | What changes | Lands in |
|---|---|---|---|---|
| 1 | Product catalogue (complete product ecosystem) | 2.6 Alumni; 2.9 Messages; 2.10 Whisper; 2.11 XENON; 3 connections; 4 distinctions table | Alumni: adult and minor contact boundaries. Messages: remove anonymous messaging, add end-to-end encryption and the Private versus Community split. Whisper: age gate, privacy model wording. XENON: not dating, interest taxonomy, age bands, no free text before mutual connection. XENON to Messages hand-off by mutual consent only. | P1-02 |
| 2 | Backend architecture and engineering requirements | 2 domains; 3 database; 4 business logic; 6 real-time; 7 security | Add Devices and Key Directory, Safety and Reporting domains. Age band, devices and key packages in the schema. Messages: the server never sees plaintext, so delivery and read state are metadata only. Ordering of group changes in the real-time design. End-to-end encryption trust model in the security section. | P1-03, P1-04, P1-08 |
| 3 | Frontend experience and design philosophy | 3 home; 4 navigation; 5 Messages, Whisper, XENON; 9 personalization; search; notifications | Feed ranking limits for minors. Navigation and features gated by age band and membership type. Messages search is on-device only. Push notifications carry no message content. Whisper screens state anonymity limits. | P1-02, P1-05, P1-06 |
| 4 | Development priorities and engineering-first execution | Phase list; strict rules | Add Phase 0 (research, legal gates, spikes). Add the end-to-end encryption track with launch gates. Replace the CLI-based workflow assumptions with the chat workflow. | Ledger, M0.x plans |
| 5 | Additional engineering principles | 1 data classification; 2 permission matrix; 11 documentation list | New data classes (ciphertext and metadata, date of birth, guardian consent, anonymous-sender escrow, safety evidence). Permission matrix gains age band and membership type as dimensions. Add the end-to-end encryption specification, safeguarding procedure and legal register to the documentation list. | P1-07, P1-09 |
| 6 | Advanced engineering requirements | 2 privacy by design; 5 search; 6 notifications; 7 file and media; 15 incident response | Anonymity wording. Search excludes private message content. Content-free push. Private attachments are encrypted on the device, so the server cannot scan or process them; define compensating controls. Incident response adds safeguarding and POCSO escalation. | P1-04, P1-08, P1-12 |
| 7 | Final CTO-level engineering directive | 3 legal; 5 abuse prevention; 13 testing; 15 AI-assisted rules | Replace generic legal text with the verified India timeline and duties. Reporting under end-to-end encryption. Add end-to-end encryption and age-band test suites. Chat workflow and evidence format. | P1-12, P1-11, ledger |
| 8 | Multi-application ecosystem and product references | Messages reference line | "WhatsApp and Slack" becomes "WhatsApp, Slack and Discord" with the Private and Community classes made explicit. | P1-02 |
| 9 | Owner's policy and workflow decisions; engineering standards | none | These become authoritative. They are recorded as D-001 to D-009. | Ledger |

## 2. Unchanged

The eleven-product definition and the distinctions between Events and Championships, Houses and Championships, Pulse and Stories, and Alumni and Achievements stand as written. So do the portability, cost and incremental-delivery rules, and the Definition of Done.

## 3. Revised Phase 1 pack and delivery order

Long-lead items go first because they depend on people other than us.

| Order | Doc | Title | Owner's list item | Status |
|---|---|---|---|---|
| 0 | P1-00 | Revision map | | Delivered |
| 1 | P1-01 | Technology selection report | | Delivered (v0.9, for review) |
| 2 | P1-12 | Legal and safeguarding review requirements (the brief for counsel) | 13 | Next |
| 3 | P1-03 | Identity, age-band and membership architecture, including schema | 2, 3 | Queued |
| 4 | P1-07 | Permission matrix | 8 | Queued |
| 5 | P1-09 | Data classification and retention | 10 | Queued |
| 6 | P1-04 | Messages encryption architecture and group key management | 6, 7 | Queued |
| 7 | P1-08 | Threat model | 9 | Queued |
| 8 | P1-10 | API contracts | 11 | Queued |
| 9 | P1-11 | Testing strategy | 12 | Queued |
| 10 | P1-02 | Product requirements | 1 | Queued |
| 11 | P1-05 | XENON discovery and matching architecture | 4 | Queued |
| 12 | P1-06 | Whisper anonymity and reporting architecture | 5 | Queued |

Each document is delivered separately so it can be reviewed by the owner and ChatGPT before the next arrives. Each separates **confirmed decisions**, **proposed decisions** and **open legal or technical questions**.
