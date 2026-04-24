# hidemyenv

`hidemyenv` is an AI-safe local secret runner. It keeps real `.env` values out of plaintext workspace files while still letting your app receive secrets at runtime.

The core workflow is:

```sh
hidemyenv run -- npm run dev
```

Secrets are decrypted in memory and injected into the child process environment. They are not written to `.env`, printed by `hidemyenv`, or included in `.env.safe`.

## Status

Early MVP.

Implemented:

- Password-based encrypted vault at `.env.hidemyenv`
- `XChaCha20-Poly1305` authenticated encryption
- `Argon2id` key derivation
- Redacted `.env.safe` generation
- Runtime env injection with `hidemyenv run -- <command>`
- `doctor` checks for unsafe plaintext env files and missing `.gitignore` rules
- Opt-in macOS Keychain storage for the vault password

Not implemented yet:

- Team sharing
- Cloud sync
- Plaintext export, intentionally omitted from the MVP

## Global Install

From this repository, install globally to `~/.local/bin`:

```sh
./scripts/install.sh
```

Or choose a different install directory:

```sh
HIDEMYENV_INSTALL_DIR=/usr/local/bin ./scripts/install.sh
```

Make sure the install directory is in your `PATH`:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

After that, run `hidemyenv` from any project:

```sh
cd /path/to/another-project
hidemyenv init
hidemyenv set DATABASE_URL
hidemyenv run -- npm run dev
```

## Bash Install From Release

After GitHub Releases are published, users can install with:

```sh
curl -fsSL https://raw.githubusercontent.com/OWNER/hidemyenv/main/scripts/install.sh | HIDEMYENV_REPO=OWNER/hidemyenv sh
```

Install a specific version:

```sh
curl -fsSL https://raw.githubusercontent.com/OWNER/hidemyenv/main/scripts/install.sh | HIDEMYENV_REPO=OWNER/hidemyenv HIDEMYENV_VERSION=v0.1.0 sh
```

The release installer expects assets named like:

```text
hidemyenv-darwin_arm64.tar.gz
hidemyenv-darwin_amd64.tar.gz
hidemyenv-linux_arm64.tar.gz
hidemyenv-linux_amd64.tar.gz
```

Each archive should contain a single executable named `hidemyenv`.

## Build From Source

```sh
go build -o hidemyenv ./cmd/hidemyenv
```

Then move the binary somewhere on your `PATH` if desired.

## Usage

Initialize a project:

```sh
hidemyenv init
```

Add a secret:

```sh
hidemyenv set DATABASE_URL
```

List secret keys without values:

```sh
hidemyenv list
```

Show a redacted value:

```sh
hidemyenv get DATABASE_URL
```

Regenerate `.env.safe`:

```sh
hidemyenv safe
```

Run your app with decrypted secrets injected into its environment:

```sh
hidemyenv run -- npm run dev
```

Check workspace safety:

```sh
hidemyenv doctor
```

Optionally store the vault password in macOS Keychain for this project:

```sh
hidemyenv keychain store
```

Check or remove the stored keychain item:

```sh
hidemyenv keychain status
hidemyenv keychain delete
```

When a keychain password exists for the current project, `hidemyenv set` and `hidemyenv run` use it automatically. If no keychain password exists, they fall back to prompting for the vault password.

## Files

`hidemyenv init` creates:

- `.env.hidemyenv`: encrypted vault
- `.env.example`: non-secret structure/defaults
- `.env.safe`: redacted file safe for AI tools to read
- `.hidemyenv.toml`: local policy/config
- `.gitignore`: updated with safe default env rules

Files that are safe for AI to read:

- `.env.example`
- `.env.safe`
- `.hidemyenv.toml`

Files that should not exist as plaintext in the workspace:

- `.env`
- `.env.local`
- `.env.decrypted`
- `.env.plain`

## Security Model

`hidemyenv` reduces accidental secret exposure to coding agents by keeping real values out of readable files.

It does not protect against:

- Malware or an attacker with full local machine access
- An AI agent that is explicitly given the vault password
- An AI agent that can access an already-unlocked keychain without human approval
- Apps that print their own environment variables to logs
- Secrets stored elsewhere in plaintext, such as shell profiles

Default behavior is intentionally conservative: there is no plaintext export command in the MVP.

## Development

Run tests:

```sh
go test ./...
```

Build:

```sh
go build ./cmd/hidemyenv
```

## Release

Create and push a version tag:

```sh
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will run tests, build release assets, generate `checksums.txt`, and publish a GitHub Release.

The release assets are named for the bash installer:

```text
hidemyenv-darwin_arm64.tar.gz
hidemyenv-darwin_amd64.tar.gz
hidemyenv-linux_arm64.tar.gz
hidemyenv-linux_amd64.tar.gz
```
