# autograd

autograd is a platform for running autograders to grade PrairieLearn
coding questions. autograd is not an autograder -- it is a client
which receives grading jobs from PrairieLearn, launches the course's
own autograder, and captures the output of the autograder to send
back to PrairieLearn.

## Course grader repo
The only required file in the course's grader repo is `configuration.yml`
in the root directory of the repo. This file contains the commands
to be run by autograd. The required fields are shown below:

```yaml
grader:
  init_commands:
    # List of commands to be run at startup (working directory $AUTOGRAD_GRADER_ROOT)
  setup_commands:
    # List of commands to be run at the beginning of a grading run (working directory $AUTOGRAD_JOB_DIR)
  grade_command: # Main autograder command (working directory $AUTOGRAD_JOB_DIR)
  grade_timeout: 600 # Timeout in seconds before killing grader process
  cleanup_commands:
    # List of commands to be run at the end of a grading run (working directory $AUTOGRAD_JOB_DIR)
```

The format of a command in `configuration.yml` is a list of strings
containing the command and its arguments. For example:

```yaml
["apt-get", "install", "-y", "clang-3.5", "libc++abi-dev", "libc++-dev", "libpng-dev"]
```

The score of the grading job is the exit code of the grading script.
This score should be in the integer range 0-255.

The following environment variables are available to autograd commands:
- `AUTOGRAD_ROOT`: Path to autograd root directory (typically
  `/opt/autograd`)
- `AUTOGRAD_GRADER_ROOT`: Path to the working copy of the grader
  repo (typically `/opt/autograd/_grader`)
- `AUTOGRAD_JOB_DIR`: Path to the temp directory for the current
  job (e.g. `/opt/autograd/job_933175825`) -- not available to init
  commands as they are not associated with a specific job

## Security
Production autograd instances run in a Debian-based Docker container
as root, which means that `apt-get` can be used to install any
necessary packages. However, this also means that by default,
autograder code will be run as root. It is recommended to run the
grader command as an unprivileged user with no read access to
anything in `$AUTOGRAD_ROOT` except for `$AUTOGRAD_JOB_DIR`, as
`$AUTOGRAD_ROOT` contains the grader repo (which may contain test
cases and solutions) as well as `$AUTOGRAD_ROOT/_ssh` which contains
the SSH deploy key for the grader repo.

Add the following commands to the end of `init_commands` to create
such a user `autograd-user`:

```yaml
- ["adduser", "--system", "--no-create-home", "autograd-user"]
- ["chmod", "-R", "o-r", "$AUTOGRAD_ROOT"]
```

Then run the grader in the following manner:

```bash
chown -R autograd-user $AUTOGRAD_JOB_DIR
su -c "./grader_command" -s /bin/bash autograd-user
```

## Grading job format
Each grading job is a JSON object written to
`$AUTOGRAD_JOB_DIR/job_data.json`. The following is a minimal grading
job:

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

## Development

### Dependencies
- libgit2 v0.24

### Setup
- Create directory `/opt/autograd`
- `export AUTOGRAD_ROOT=/opt/autograd`
- Copy `configuration.yml` to `$AUTOGRAD_ROOT` and configure values

### autograd configuration guidelines
- Grading, started, and result queue names should be of the form
  `cs225-grade`, `cs225-started`, and `cs225-result`, with the course
  code changed appropriately
- `grader_repo.repo_url` must be an SSH URL (e.g. `git@github.com:...`)
- `grader_repo.commit` can be any of the following formats:
    - Commit hash
    - Branch name: `origin/<branchname>` or `refs/remotes/origin/<branchname>`
    - Tag name: `<tagname>` or `refs/tags/<tagname>`

### Running with Docker
```bash
docker run -it --rm --name autograd \
    -v '/absolute/path/to/configuration.yml:/opt/autograd/_conf/configuration.yml' \
    -v '/absolute/path/to/deploy_key:/opt/autograd/_ssh/ssh-privatekey' \
    -v '/absolute/path/to/deploy_key.pub:/opt/autograd/_ssh/ssh-publickey' \
     prairielearn/autograd
```
