---
name: branching-workflow
description: Git workflow — commit directly to main during alpha, branch later
metadata:
  type: feedback
---

While the project is in alpha, commit and push directly to `main` (no feature branches / PRs). Once out of alpha, switch to branching.

**Why:** Solo early-stage development; branching overhead isn't worth it yet, and releases (version bump + tag) live on main.

**How to apply:** When asked to commit/push on `main` during alpha, do it directly — skip the usual "branch first on the default branch" step. Revisit once the version drops its alpha pre-release marker.
