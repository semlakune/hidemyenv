# hidemyenv

`hidemyenv` keeps real `.env` values out of plaintext project files so AI coding agents can understand your environment structure without seeing your secrets.

Instead of keeping this in your workspace:

```env
DATABASE_URL=postgres://user:password@localhost:5432/app
JWT_SECRET=super-secret-value
OPENAI_API_KEY=sk-...
```

you keep secrets encrypted in `.env.hidemyenv`, expose only redacted values in `.env.safe`, and run your app through:

```sh
hidemyenv run -- npm run dev
```

Secrets are decrypted in memory and injected into the child process environment. `hidemyenv` does not write a plaintext `.env` file or print secret values.

## Why

AI coding tools often need to inspect your project files. If your workspace contains `.env`, `.env.local`, or other plaintext secret files, those values can be read accidentally.

`hidemyenv` is designed for local development workflows where:

- AI can read env key names and config shape
- AI can read `.env.example` and `.env.safe`
- AI should not read actual secret values
- Your app still receives real secrets at runtime

## Install

Install globally from a local checkout:

```sh
./scripts/install.sh
```

By default, the binary is installed to:

```text
~/.local/bin/hidemyenv
```

Make sure that directory is in your `PATH`:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

Install to a custom directory:

```sh
HIDEMYENV_INSTALL_DIR=/usr/local/bin ./scripts/install.sh
```

Check the installed version:

```sh
hidemyenv version
```

## Install From GitHub Release

After this project has GitHub Releases, users can install with:

```sh
curl -fsSL https://raw.githubusercontent.com/OWNER/hidemyenv/main/scripts/install.sh | HIDEMYENV_REPO=OWNER/hidemyenv sh
```

Install a specific version:

```sh
curl -fsSL https://raw.githubusercontent.com/OWNER/hidemyenv/main/scripts/install.sh | HIDEMYENV_REPO=OWNER/hidemyenv HIDEMYENV_VERSION=v0.1.1 sh
```

Replace `OWNER/hidemyenv` with the actual GitHub repository.

## Upgrade

Run the same installer again:

```sh
curl -fsSL https://raw.githubusercontent.com/OWNER/hidemyenv/main/scripts/install.sh | HIDEMYENV_REPO=OWNER/hidemyenv sh
```

The installer overwrites the old binary and prints the previous and installed versions when possible.

Project files are not modified during upgrade. Existing vaults such as `.env.hidemyenv` remain untouched.

## Quick Start

Go to any project:

```sh
cd /path/to/your-project
```

Initialize `hidemyenv`:

```sh
hidemyenv init
```

Add one secret manually:

```sh
hidemyenv set DATABASE_URL
```

Run your app with decrypted secrets injected at runtime:

```sh
hidemyenv run -- npm run dev
```

Check for unsafe plaintext env files:

```sh
hidemyenv doctor
```

## Import Existing `.env`

If your project already has a `.env` file with many values, import it in one command:

```sh
hidemyenv import .env
```

This will:

- Read key/value pairs from `.env`
- Encrypt values into `.env.hidemyenv`
- Regenerate `.env.safe`
- Avoid printing secret values
- Leave the original `.env` file in place

After verifying your app works with `hidemyenv run`, remove or move the plaintext file:

```sh
mv .env .env.backup
```

Then run:

```sh
hidemyenv doctor
```

## Daily Usage

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

Run any command with secrets injected:

```sh
hidemyenv run -- pnpm dev
hidemyenv run -- npm test
hidemyenv run -- go test ./...
```

Avoid testing with commands that print the entire environment, such as `env`, because those commands will print the secrets that were intentionally injected.

Use a safer check instead:

```sh
hidemyenv run -- sh -c 'test -n "$DATABASE_URL" && echo "DATABASE_URL is available"'
```

## macOS Keychain

By default, `hidemyenv` prompts for the vault password when it needs to decrypt secrets.

On macOS, you can optionally store the vault password in Keychain for the current project:

```sh
hidemyenv keychain store
```

Check status:

```sh
hidemyenv keychain status
```

Delete the stored password:

```sh
hidemyenv keychain delete
```

When a Keychain password exists, `hidemyenv set`, `hidemyenv import`, and `hidemyenv run` use it automatically. If no Keychain password exists, they fall back to prompting.

## Files

`hidemyenv init` creates or updates these files in the current project:

```text
.env.hidemyenv
.env.example
.env.safe
.hidemyenv.toml
.gitignore
```

File meanings:

- `.env.hidemyenv`: encrypted vault containing real secret values
- `.env.example`: non-secret env structure and defaults
- `.env.safe`: redacted env file safe for AI tools to read
- `.hidemyenv.toml`: non-secret local config and policy
- `.gitignore`: updated with safer defaults for env files

Files that are safe for AI to read:

- `.env.example`
- `.env.safe`
- `.hidemyenv.toml`

Files that should not remain as plaintext in the workspace:

- `.env`
- `.env.local`
- `.env.decrypted`
- `.env.plain`

## Commands

```text
hidemyenv init
hidemyenv set KEY
hidemyenv import [.env]
hidemyenv list
hidemyenv get KEY
hidemyenv safe
hidemyenv run -- <command>
hidemyenv doctor
hidemyenv keychain status|store|delete
hidemyenv version
```

## Security Model

`hidemyenv` reduces accidental exposure of local development secrets to AI coding agents by keeping real values out of plaintext workspace files.

It protects against common mistakes such as:

- Leaving `.env` visible in the project while using AI tools
- Accidentally committing plaintext env files
- Asking AI to inspect a project and exposing secret values from config files

It does not protect against:

- Malware or an attacker with full access to your machine
- An AI agent that you explicitly give the vault password to
- An AI agent that can access an already-unlocked Keychain without human approval
- Your own app printing secrets to logs
- Secrets stored elsewhere in plaintext, such as shell profiles

There is intentionally no plaintext export command in the MVP.

## Crypto

Current implementation:

- Encryption: `XChaCha20-Poly1305`
- KDF: `Argon2id`
- Vault file: `.env.hidemyenv`
- Runtime model: decrypt in memory, inject into child process environment

## Development

Run tests:

```sh
go test ./...
```

Build locally:

```sh
go build -o hidemyenv ./cmd/hidemyenv
```

Install local build globally:

```sh
./scripts/install.sh
```

## Release

Create the next patch release automatically:

```sh
./scripts/release.sh
```

Create the next minor or major release:

```sh
./scripts/release.sh minor
./scripts/release.sh major
```

Preview the next version without creating a tag:

```sh
HIDEMYENV_DRY_RUN=1 ./scripts/release.sh
```

The release script reads the latest `vMAJOR.MINOR.PATCH` tag, creates the next tag, and pushes it to `origin`. GitHub Actions will then run tests, build release assets, generate `checksums.txt`, and publish a GitHub Release.

The release installer expects assets named:

```text
hidemyenv-darwin_arm64.tar.gz
hidemyenv-darwin_amd64.tar.gz
hidemyenv-linux_arm64.tar.gz
hidemyenv-linux_amd64.tar.gz
```

Each archive must contain one executable named `hidemyenv`.
