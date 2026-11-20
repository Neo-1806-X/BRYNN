# BRYNN

A unified, multi-tenant digital ecosystem for educational institutions: eleven connected products (Control, Events, Houses, Championships, Achievements, Alumni, Pulse, Stories, Messages, Whisper, XENON) sharing one identity, one design language and one set of platform services.

**Status:** pre-alpha. Phase 1 (architecture) and Milestone M0.1 (engineering baseline).
**Read first:** [`PROJECT_LEDGER.md`](PROJECT_LEDGER.md) — decisions, status, open questions, and how we work.

## Quick start

Requires Git, Docker and `make`. Nothing else is installed on your machine.

```sh
make check   # gofmt check, go vet, race-enabled tests (runs in a Go container)
make up      # build and start the API on http://127.0.0.1:8080 (waits until healthy)
make smoke   # probe /healthz, /readyz and /version
make down    # stop and remove the containers
```

If `make check` reports unformatted files, run `make fmt` and then `make check` again.

## Layout

```text
PROJECT_LEDGER.md          decisions, status, risks, evidence (read first)
apps/api/                  Go API process (modular monolith; more roles later)
scripts/check-go.sh        the Go quality gate, shared by `make check` and CI
docs/phase1/               Phase 1 architecture pack (revision map, technology report, ...)
docs/milestones/           milestone plans with exact steps and acceptance criteria
docs/adr/                  architecture decision records
.github/workflows/ci.yml   CI (portable: it only calls scripts in this repository)
```

## Ground rules

- Security, data integrity and correctness come before polish. Nothing is "done" without evidence (see the ledger).
- No custom cryptography. No paid dependency without an explained alternative.
- Private conversations are end-to-end encrypted by design; there is no server-readable fallback.
