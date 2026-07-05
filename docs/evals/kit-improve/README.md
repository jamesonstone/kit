# Kit Improve Evals

This directory contains committed benchmark definitions for `kit improve`.

- `suites/` selects held-in and held-out task groups.
- `tasks/` contains declarative benchmark tasks.
- `fixtures/repos/` contains small checked-in repositories copied into
  disposable `.kit/improve/runs/<id>/workspaces/` directories.
- `schemas/` documents the JSON/YAML contracts used by commands and artifacts.

Generated traces, candidates, scorecards, and reports live under
`.kit/improve/` and are ignored by git.
