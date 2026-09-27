# Migration fixtures

Each directory is the exact output of `kit init` from one released Kit version,
run in an empty Git repository (see `scripts/harvest-legacy-fingerprints.sh`);
`.kit-release` names the release. `.env` and `.envrc` are omitted because they
are local-only and never committed, which also reproduces a clean CI checkout.

| Fixture | Release | Shape |
| --- | --- | --- |
| `v1` | v1.0.0 | scaffold v1: verbose entry files, no scaffold version, obsolete config keys |
| `v2` | v1.0.60 | scaffold v2: routing entry files plus `docs/agents/*` support docs |
| `v3-precontract` | v3.0.24 | scaffold v3 before the managed contract block, workflow manifests, retired rules |
| `phase2` | v3.0.25 | managed contract block with transitional support docs and retired rules |
