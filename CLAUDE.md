# autovibe Project - Claude Instructions

## Repository Configuration
- **AI Hakzarov's fork**: https://github.com/ai-hakzarov/autovibe (where Claude works)
- **Main repository**: https://github.com/hakzarov/autovibe (where PRs are created)
- **PRs are created in**: hakzarov/autovibe
- **View PRs at**: https://github.com/hakzarov/autovibe/pulls

## GitHub Authentication  
- Use ai-hakzarov account and token for all GitHub operations
- PR creation command: `GITHUB_TOKEN=$(gh auth token --hostname github.com --user ai-hakzarov) gh pr create --repo hakzarov/autovibe`

## Project Guidelines
- Keep PRs under 1000 lines and 10 files
- Always check CI status before presenting PRs as "ready"
- Fix any failing tests or CI issues before submission
- **ALWAYS resolve merge conflicts immediately** - never leave conflicts unresolved
- When conflicts arise, prioritize the main branch version and integrate changes carefully