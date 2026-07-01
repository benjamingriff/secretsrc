# Secret Src

A beautiful TUI (Terminal User Interface) for viewing AWS Secrets Manager secrets **and SSM Parameter Store parameters**, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Beautiful TUI**: Built with Charm's Bubble Tea framework for a polished terminal experience
- **Two sources, one app**: Browse AWS Secrets Manager secrets and SSM Parameter Store parameters, and switch between them instantly with `tab`. The whole UI recolours as a mode signal — **fluoro pink** for Secrets Manager, **fluoro green** for SSM. The last-used source is remembered between launches.
- **Shared credentials**: Both sources reuse the same profile, region, and MFA/role-assumption setup
- **On-Demand Fetching**: Values are only fetched (and SecureString parameters decrypted) when you explicitly request them (security-first)
- **Clipboard Support**: Copy full values as plain text; for JSON secrets, copy as formatted JSON or copy an individual top-level field
- **Profile & Region Switching**: Easily switch between AWS profiles and regions
- **Pagination**: Handles large numbers of secrets/parameters with built-in pagination

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
        "ssm:DescribeParameters",
        "ssm:GetParameter"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt"
      ],
      "Resource": "*"
    }
  ]
}
```

**Notes**:
- The `ssm:*` permissions are only needed if you use the SSM Parameter Store view (`tab` to switch to it).
- The `kms:Decrypt` permission is needed to read secrets encrypted with custom KMS keys and to decrypt `SecureString` parameters. If you don't have it, listing still works (metadata only), but viewing a `SecureString` value returns an access-denied error.

## Usage

### Key Bindings

#### List Screen
- `↑/k` - Move up
- `↓/j` - Move down
- `←/h` - Move left
- `→/l` - Move right
- `enter` - View details
- `tab` - **Switch source** between Secrets Manager (pink) and SSM Parameter Store (green)
- `/` - Start filtering by name
- `esc` - Clear the active filter when filtering, otherwise quit
- `space` / `pgdn` - Move to the next grid screen
- `pgup` - Move to the previous grid screen
- `p` - Switch AWS profile
- `g` - Switch AWS region
- `r` - Refresh the list
- `n` - Load next AWS page (when available)
- `b` - Load previous AWS page
- `?` - Toggle help
- `q` - Quit

`tab` works from any screen and returns you to the list in the new source.

#### Detail Screen
- `v` - View the value (decrypt and display; `SecureString` parameters are decrypted on fetch)
- `c` - Copy the value to clipboard (plain text)
- `j` - Copy the value as formatted JSON *(Secrets Manager only)*
- `k` - Copy a top-level JSON field from the loaded secret *(Secrets Manager only)*
- `esc` / `q` - Back to the list
- `ctrl+c` - Force quit

> SSM parameters are treated as plain strings, so the detail screen offers only `v` and `c` — there is no JSON field picker.

#### Profile & Region Selector Screens
- `↑/k` - Move up in list
- `↓/j` - Move down in list
- `enter` - Select profile/region and switch
- `esc` / `q` - Cancel and go back
- `/` - Filter/search (built-in)

### Workflow

1. **Browse**: Launch the app to see a list of all secrets (or parameters) in your current AWS region
2. **Switch Source**: Press `tab` to flip between Secrets Manager and SSM Parameter Store — the UI recolours to tell you which one you're in
3. **Switch Profile/Region**: Press `p` to select a different AWS profile or `g` to select a different region (shared across both sources)
4. **View Details**: Press `enter` on an item to see its metadata (name, ARN, last modified date; plus type and version for parameters)
5. **Fetch Value**: Press `v` to fetch the value on-demand (`SecureString` parameters are decrypted at this point)
6. **Copy to Clipboard**: Press `c` for plain text. For JSON secrets, press `j` for a JSON-formatted copy or `k` to choose a top-level field

## Security Considerations

- **On-Demand Fetching**: Values are never automatically fetched or displayed. You must explicitly press `v` to decrypt them. Listing (both secrets and parameters) returns metadata only — never values.
- **Memory Clearing**: Values are cleared from memory when you navigate away from the detail screen or switch source with `tab`.
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
│   │   ├── client.go               # AWS client initialization (Secrets Manager + SSM)
│   │   ├── secrets.go              # Secrets Manager operations
│   │   ├── parameters.go           # SSM Parameter Store operations
│   │   └── config.go               # Profile/region management
│   ├── models/
│   │   └── entry.go                # Shared Entry data structure (secrets + parameters)
│   └── ui/
│       ├── app.go                  # Main Bubble Tea model, mode toggle
│       ├── view.go                 # View rendering (mode-aware colours)
│       ├── keys.go                 # Key bindings
│       ├── styles.go               # Lipgloss styles + accent colours
│       └── components/
│           ├── grid.go             # Entry grid component
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

### "Failed to load secrets/parameters: AccessDeniedException"
- Ensure your AWS user/role has the required IAM permissions (see above) — note the separate `ssm:*` actions for the Parameter Store view
- Check that you're using the correct AWS profile

### "Failed to load value: AccessDeniedException" on a SecureString parameter
- Reading a `SecureString` value decrypts it, which requires `kms:Decrypt` on the parameter's KMS key. Listing still works without it.

### "No secrets/parameters found in this region"
- Verify that the items exist in the current AWS profile and region via the AWS Console or CLI
- If you rely on profile-specific regions, ensure the correct profile is selected or set `AWS_REGION`
- Remember you may just be in the other source — press `tab` to switch

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
- [x] SSM Parameter Store support (toggle with `tab`)
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
