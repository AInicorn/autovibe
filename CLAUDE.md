# AutoVibe Project Configuration

## Git Configuration
- Working repository: git@github.com:ai-hakzarov/AutoVibe.git
- PRs should be opened in: https://github.com/AInicorn/AutoVibe

## GitHub Authentication
- Always authenticate as ai-hakzarov: `gh auth switch --user ai-hakzarov`
- May need to clear GITHUB_TOKEN first: `unset GITHUB_TOKEN`

## GitHub PR Creation
When creating PRs for this project, use:
```bash
GITHUB_TOKEN=$(gh auth token --hostname github.com --user ai-hakzarov) gh pr create --repo AInicorn/AutoVibe
```

This ensures PRs are opened in the main repository (AInicorn/AutoVibe) while working from the ai-hakzarov fork.