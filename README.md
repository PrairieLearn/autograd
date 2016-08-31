# autograd

## Dependencies
- libgit2 v0.24

## Setup
- Create directory `/opt/autograd`
- `export AUTOGRAD_ROOT=/opt/autograd`
- Copy `configuration.yml` to `$AUTOGRAD_ROOT` and configure values

## Autograd configuration guidelines
- Grading queue and results queue names should be of the form `cs225-grade` and `cs225-results`, with the course code changed appropriately
- `grader_repo.repo_url` must be an SSH URL
- `grader_repo.commit` can be any of the following formats:
    - Commit hash
    - Branch name: `refs/remotes/origin/<branchname>`
    - Tag name:  `refs/tags/<tagname>`
