# Secret Src

A beautiful TUI (Terminal User Interface) for viewing and managing AWS Secrets Manager secrets, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Beautiful TUI**: Built with Charm's Bubble Tea framework for a polished terminal experience
- **AWS Integration**: Uses the AWS SDK for Go v2 to interact with Secrets Manager
- **Credential Management**: Automatically reads AWS credentials from `~/.aws/` (same as AWS CLI)
- **On-Demand Secret Fetching**: Secrets are only decrypted when you explicitly request them (security-first)
- **Clipboard Support**: Copy full secret values as plain text or JSON, and copy individual top-level JSON fields
- **Profile & Region Switching**: Easily switch between AWS profiles and regions
- **Pagination**: Handles large numbers of secrets with built-in pagination

## Installation

### Prerequisites

- Go 1.21 or later
- `make` (for the convenience targets below)
- AWS credentials configured (via `aws configure` or environment variables)
- Required IAM permissions (see below)

### Local Development

```bash
# Clone the repository
git clone https://github.com/benjamingriff/secretsrc.git
cd secretsrc

# Show available tasks
make help

# Run the app directly from source
make run

# Run the test suite
make test

# Build a local binary in the repo root
make build
./secretsrc
```

### Install on Your Machine

```bash
cd /path/to/secretsrc
make install
```

`make install` runs `go install ./cmd/secretsrc`, which installs `secretsrc` into:

- `$(go env GOBIN)` if `GOBIN` is set
- otherwise `$(go env GOPATH)/bin`

If `secretsrc` is not directly runnable after install, add your Go bin directory to your `PATH`.

For `zsh`, if `GOBIN` is not set, this is the usual setup:

```bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

You can then confirm the install with:

```bash
command -v secretsrc
secretsrc
```

### Install Without Using the Repo

If you want to install the latest published version directly from Go without building from your checkout:

```bash
go install github.com/benjamingriff/secretsrc/cmd/secretsrc@latest
```

## Updating a Local Install

### If You Installed from This Repo Checkout

```bash
cd /path/to/secretsrc
git pull --ff-only
make install
```

If you have local uncommitted changes in the repo, either commit them, stash them, or reapply them after pulling before you reinstall.

### If You Installed with `go install ...@latest`

```bash
go install github.com/benjamingriff/secretsrc/cmd/secretsrc@latest
```

That fetches the latest version again and replaces the installed `secretsrc` binary in your Go bin directory.

## AWS Credentials Setup

Secret Src uses the same credential chain as the AWS CLI:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. Shared config file (`~/.aws/config`)

To set up credentials:

```bash
aws configure
```

You can also use the `AWS_PROFILE` and `AWS_REGION` environment variables to override defaults:

```bash
export AWS_PROFILE=myprofile
export AWS_REGION=us-west-2
./secretsrc
```

If you do not set `AWS_REGION`, Secret Src will let the AWS SDK resolve the region from your shared AWS config for the selected profile.

## Required IAM Permissions

Your AWS user or role needs the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:ListSecrets",
        "secretsmanager:DescribeSecret",
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "kms:ViaService": "secretsmanager.*.amazonaws.com"
        }
      }
    }
  ]
}
```

**Note**: The KMS decrypt permission is only needed if your secrets are encrypted with custom KMS keys.

## Usage

### Key Bindings

#### Secret List Screen
- `↑/k` - Move up
- `↓/j` - Move down
- `←/h` - Move left
- `→/l` - Move right
- `enter` - View secret details
- `/` - Start filtering by secret name
- `esc` - Clear the active filter when filtering, otherwise quit
- `space` / `pgdn` - Move to the next grid screen
- `pgup` - Move to the previous grid screen
- `p` - Switch AWS profile
- `g` - Switch AWS region
- `r` - Refresh secret list
- `n` - Load next AWS page (when available)
- `b` - Load previous AWS page
- `?` - Toggle help
- `q` - Quit

#### Secret Detail Screen
- `v` - View secret value (decrypt and display)
- `c` - Copy secret value to clipboard (plain text)
- `j` - Copy secret value to clipboard (JSON formatted)
- `k` - Copy a top-level JSON field value from the loaded secret
- `esc` / `q` - Back to secret list
- `ctrl+c` - Force quit

#### Profile & Region Selector Screens
- `↑/k` - Move up in list
- `↓/j` - Move down in list
- `enter` - Select profile/region and switch
- `esc` / `q` - Cancel and go back
- `/` - Filter/search (built-in)

### Workflow

1. **Browse Secrets**: Launch the app to see a list of all secrets in your current AWS region
2. **Switch Profile/Region**: Press `p` to select a different AWS profile or `g` to select a different region
3. **View Details**: Press `enter` on a secret to see its metadata (name, ARN, last modified date)
4. **Decrypt Secret**: Press `v` to fetch and decrypt the secret value (on-demand for security)
5. **Copy to Clipboard**: Press `c` for plain text, `j` for JSON-formatted copy, or `k` to choose a top-level field from a JSON object secret

## Security Considerations

- **On-Demand Fetching**: Secret values are never automatically fetched or displayed. You must explicitly press `v` to decrypt them.
- **Memory Clearing**: Secret values are cleared from memory when you navigate away from the detail screen.
- **Alternate Screen**: The app uses the terminal's alternate screen buffer, so secrets don't remain in scrollback history.
- **Clipboard Persistence**: Be aware that copied secrets will remain in your clipboard after the app closes. Clear your clipboard if needed.

## Project Structure

```
secretsrc/
├── cmd/
│   └── secretsrc/
│       └── main.go                 # Application entry point
├── pkg/
│   ├── aws/
│   │   ├── client.go               # AWS client initialization
│   │   ├── secrets.go              # Secrets Manager operations
│   │   └── config.go               # Profile/region management
│   ├── models/
│   │   └── secret.go               # Data structures
│   └── ui/
│       ├── app.go                  # Main Bubble Tea model
│       ├── view.go                 # View rendering
│       ├── keys.go                 # Key bindings
│       ├── styles.go               # Lipgloss styles
│       └── components/
│           ├── list.go             # Secret list component
│           ├── profile_selector.go # Profile selection
│           └── region_selector.go  # Region selection
├── go.mod
├── go.sum
└── README.md
```

## Troubleshooting

### "AWS credentials not found"
- Run `aws configure` to set up your credentials
- Or set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables

### "Failed to load secrets: AccessDeniedException"
- Ensure your AWS user/role has the required IAM permissions (see above)
- Check that you're using the correct AWS profile

### "No secrets found in this region"
- Verify that secrets exist in the current AWS profile and region via the AWS Console or CLI
- If you rely on profile-specific regions, ensure the correct profile is selected or set `AWS_REGION`

### Clipboard not working on Linux
- The `atotto/clipboard` library requires X11 on Linux
- Install `xclip` or `xsel`: `sudo apt-get install xclip`

## Roadmap

- [x] List secrets with pagination
- [x] View secret details
- [x] On-demand secret value fetching
- [x] Clipboard copy (plain text & JSON)
- [x] Interactive profile selector
- [x] Interactive region selector
- [x] Search/filter secrets
- [ ] Secret versioning support
- [ ] Create/update/delete secrets
- [ ] Secret rotation status
- [ ] Export secrets to file

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm
- Styled with [Lipgloss](https://github.com/charmbracelet/lipgloss)
- Uses [Bubbles](https://github.com/charmbracelet/bubbles) components
- AWS SDK for Go v2
