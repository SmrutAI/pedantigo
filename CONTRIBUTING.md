# Contributing to Pedantigo

## Setup

```bash
# Clone
git clone git@github.com:SmrutAI/pedantigo.git
cd pedantigo

# Install dependencies
make install

# Verify setup
make test
```

## Development

```bash
make help           # Show all commands
make build          # Build
make test           # Run tests
make test-coverage  # Tests with coverage report
make fmt            # Format code
make lint           # Run linter
make pre-commit     # Run all checks before committing
```

## Code Style

- Run `make fmt` before committing
- Maintain 80%+ test coverage (enforced by `make pre-commit`)
- Follow existing patterns in the codebase

## Pull Requests

1. Fork and create a feature branch
2. Write tests for new functionality
3. Run `make pre-commit`
4. Open PR with clear description

## CI/CD

Tests run automatically on every push and PR via GitHub Actions.

### Coverage Badge Setup (Maintainers)

1. Create a public GitHub Gist (can be empty)
2. Copy the Gist ID from the URL: `gist.github.com/username/GIST_ID`
3. Create a Personal Access Token with `gist` scope
4. Add repo secrets:
   - `GIST_TOKEN` - Your PAT
   - `GIST_ID` - The Gist ID
5. Update `README.md` badge URL with actual Gist ID

## Release Tagging

Releases use semantic versioning: `v{MAJOR}.{MINOR}.{PATCH}`

```bash
# Create annotated tag
git tag -a v0.1.0 -m "Release description"

# Push to remote
git push origin v0.1.0
```

| Version | When |
|---------|------|
| `v0.x.x` | Initial development |
| `v1.0.0` | First stable release |
| Patch (`0.0.x`) | Bug fixes |
| Minor (`0.x.0`) | New features (backward compatible) |
| Major (`x.0.0`) | Breaking changes |

```bash
# Other tag commands
git tag                          # List tags
git show v0.1.0                  # Show tag details
git tag -d v0.1.0                # Delete local tag
git push origin --delete v0.1.0  # Delete remote tag
```
