# Copilot onboarding for `pulumi-webflow`

Concise, task-agnostic instructions so an agent can work efficiently without extra repo exploration. Trust these steps first; search only when something here is missing or incorrect. Keep changes minimal and follow the Pulumi provider boilerplate patterns.

## What this repo is
- Unofficial Pulumi **native provider** for Webflow (Go provider + generated SDKs for Node/TS, Python, Go, .NET, Java).
- Large repo with many docs/examples; core source lives in `provider/`, generated SDKs in `sdk/`, examples in `examples/`, docs in `docs/`.
- Toolchain versions managed by **mise** (`.config/mise.toml`): Go `latest`, Node `20.19.5`, Python `3.11.8`, .NET `8.0.414`, Java `corretto-11`, Pulumi `latest`, pulumictl `0.0.50`, schema-tools `0.6.0`, golangci-lint `2.7.2`, yarn `1.22.22`. Pulumi home pinned to `.pulumi` in repo.

## Bootstrap (do this first)
1. From repo root, activate tools: `eval "$(mise activate bash)" && mise install`. This installs pulumictl/schema-tools etc. and avoids “pulumictl: not found” during builds.
2. Ensure Go/Pulumi in PATH: `go version`, `pulumi version`.
3. For IDE/tests, export Webflow token if needed: `export WEBFLOW_API_TOKEN="your-token"` (required only for integration/example tests).

## Build, lint, and test (validated commands)
- **Provider unit tests (validated locally):**
  - `make test_provider` (runs `go test -short` in `provider/`, ~2–3 minutes). Succeeds locally; saw harmless preamble `pulumictl: not found` when tools not installed—run mise to suppress.
- **Code generation rule (critical):** After any change in `provider/`, run `make codegen` to refresh `provider/cmd/.../schema.json` and all SDKs. CI fails if the worktree is dirty.
- **Build provider only:** `make provider` (outputs `bin/pulumi-resource-webflow`).
- **Full build (provider + SDKs):** `make build` (calls `make build_sdks`; heavier).
- **Lint:** `make lint` (golangci-lint using `.golangci.yml`; workflow temporarily rewrites `go:embed` to ` goembed`).
- **Example/integration tests:** `make test_examples` or `cd examples && go test -tags=all ...` need `WEBFLOW_API_TOKEN` (skips when missing).
- **Language SDK builds (after codegen):** `make build_nodejs|build_python|build_go|build_dotnet|build_java`; they expect dependencies from mise and may write artifacts under `sdk/*`.

## Project layout shortcuts
- `provider/`: Go provider implementation (`provider.go`, `config.go`, `auth.go`, `*_resource.go`, tests). Entry binary at `provider/cmd/pulumi-resource-webflow/main.go`; schema extracted to `provider/cmd/.../schema.json`.
- `sdk/`: Generated; do not hand-edit. Language-specific READMEs under each SDK.
- `examples/`: Extensive Pulumi programs plus Go-based tests (`examples/*.go`). Many tests skip without `WEBFLOW_API_TOKEN`.
- `docs/`: Guides, API docs (`docs/api/*.md`), troubleshooting, sprint artifacts. `CLAUDE.md` repeats the “run make codegen after provider changes” rule.
- `Makefile`: All build/test targets and codegen steps; uses `pulumictl convert-version` and `pulumi package gen-sdk`.
- Config & lint: `.config/mise.toml`, `.golangci.yml`, `.pulumi.version`.
- GitHub Actions: `build.yml` (push) runs codegen, provider build/tests, SDK matrix build, example tests, and lint; `run-acceptance-tests.yml` runs on `comment /run-acceptance-tests`; `pull-request.yml` posts PR comment for maintainers; `release.yml` handles publishing.

## Working tips to avoid CI friction
- Always run `make codegen` after touching `provider/` Go code, then commit generated changes.
- Ensure tools from mise are installed before any Make target to avoid missing pulumictl/schema-tools.
- Keep changes surgical; avoid editing generated `sdk/` files directly—regenerate instead.
- Integration/example tests require `WEBFLOW_API_TOKEN`; provider unit tests do not.
- If linters complain about `go:embed`, note CI temporarily rewrites directives; do not commit those rewrites.
- CI checks for a clean worktree; ensure `git status` clean after codegen/build.

## When to search
Use search only if these instructions lack a needed detail or observed behavior diverges (e.g., new targets or errors). Otherwise rely on this file for fast execution.
