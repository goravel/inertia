# Roadmap — goravel/inertia

The official [Inertia.js](https://inertiajs.com) v3 adapter for
[Goravel](https://github.com/goravel/framework). Tracks what's shipped and
what's next.

---

## Current state

**Latest published:** `v1.18.0` · official package under the `goravel/`
organization.

The adapter is feature-complete for both first-class stacks (Vue 3 and React):

| Area | Status |
|------|--------|
| Render (HTML initial + X-Inertia JSON), version check (409) | ✅ |
| Per-request v3 props (Defer/Optional/Always/Merge/DeepMerge/Prepend/Scroll/Once/Prop) | ✅ |
| Shared props — hybrid: `share()` middleware (per-request) + facade `Share`/`ShareFunc` | ✅ |
| Session flash + validation errors bridged into `props.flash` / `props.errors` | ✅ |
| Inertia-aware redirects (303/302) + external `Location` | ✅ |
| Vite integration — HMR dev (`public/hot`) + hashed prod build | ✅ |
| Asset versioning from manifest hash | ✅ |
| SSR + **automatic CSR fallback** when SSR is unreachable | ✅ |
| `inertia:install` artisan command — **Vue 3 or React** demo scaffold | ✅ |
| `HandleInertiaRequests` publishable middleware (Laravel-style) | ✅ |
| `package:install` setup (auto-registers ServiceProvider) | ✅ |
| Tests — core coverage ~89% | ✅ |
| Mocks (`mocks/Inertia.go`) for consumer controller tests | ✅ |
| README / THIRD_PARTY_NOTICES / LICENSE | ✅ |

### Key architectural fact

The Go side (adapter, manager, props, middleware, SSR fallback) is **fully
agnostic to the JS framework** — the Inertia protocol is framework-independent.
Everything frontend-specific lives in `console/stubs/` and is selected by
`inertia:install`. **Adding a new stack = new stubs + installer flag, no Go core
changes.**

---

## Versioning

As an official Goravel package, releases **track the framework's minor line**:
framework `v1.18.x` → `goravel/inertia` `v1.18.x`. This tells users at a glance
which Goravel version a release targets.

> Pre-adoption `v0.x` tags used the old `github.com/eddyjj92/goravel-inertia`
> module path and are superseded by `v1.18.0`.

---

## Next

### Documentation & CI (post-adoption)
- [ ] Inertia page in [`goravel/docs`](https://github.com/goravel/docs)
      (on the `upgrade/v1.19.0` branch).
- [ ] Align CI with the org convention (mirror
      [`goravel/redis`](https://github.com/goravel/redis/tree/master/.github/workflows)),
      add a Windows test job.

### Quality
- [ ] Automated tests for `setup/` (the `package:install` path).
- [ ] `CHANGELOG.md`.
- [ ] Scaffolded `INERTIA.md` records the chosen stack.
- [ ] Optional interactive stack prompt when `--stack` is omitted.

### Scaffolding
- Keep the Vue + React scaffolding in-repo and refine it deeply over time
  (shared stub abstraction, demo polish).

---

## Conventions

- Work on a feature/fix branch → PR into `master` → maintainer review.
- Commits authored by the maintainer only (no co-author trailers).
- Each change ships its own tests.
