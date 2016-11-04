# Autograd

## Dependencies
- libgit2 v0.24

## Setup
- Create directory `/opt/autograd`
- `export AUTOGRAD_ROOT=/opt/autograd`
- Copy `configuration.yml` to `$AUTOGRAD_ROOT` and configure values

## Autograd configuration guidelines
- Grading, started, and result queue names should be of the form `cs225-grade`, `cs225-started`, and `cs225-result`, with the course code changed appropriately
- `grader_repo.repo_url` must be an SSH URL (e.g. `git@github.com:...`)
- `grader_repo.commit` can be any of the following formats:
    - Commit hash
    - Branch name: `origin/<branchname>` or `refs/remotes/origin/<branchname>`
    - Tag name: `<tagname>` or `refs/tags/<tagname>`

## Running with Docker
```bash
docker run -it --rm --name autograd \
    -v '/absolute/path/to/configuration.yml:/opt/autograd/_conf/configuration.yml' \
    -v '/absolute/path/to/deploy_key:/opt/autograd/_ssh/ssh-privatekey' \
    -v '/absolute/path/to/deploy_key.pub:/opt/autograd/_ssh/ssh-publickey' \
     prairielearn/autograd
```

## Minimal sample grading job JSON
```javascript
{
    "gid": "g1", // Grading job ID (string)
    "submission": {
        "submittedAnswer": {
            // Object format determined by PL question server.js
        }
    }
}
```
