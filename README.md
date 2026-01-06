# Secret Src

A beautiful TUI (Terminal User Interface) for viewing and managing AWS Secrets Manager secrets, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Beautiful TUI**: Built with Charm's Bubble Tea framework for a polished terminal experience
- **AWS Integration**: Uses the AWS SDK for Go v2 to interact with Secrets Manager
- **Credential Management**: Automatically reads AWS credentials from `~/.aws/` (same as AWS CLI)
- **On-Demand Secret Fetching**: Secrets are only decrypted when you explicitly request them (security-first)
- **Clipboard Support**: Copy secret values to clipboard in both plain text and JSON formats
- **Profile & Region Switching**: Easily switch between AWS profiles and regions
- **Pagination**: Handles large numbers of secrets with built-in pagination

## Installation

### Prerequisites

- Go 1.21 or later
- AWS credentials configured (via `aws configure` or environment variables)
- Required IAM permissions (see below)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/benjamingriff/secretsrc.git
cd secretsrc

# Build the binary
go build -o secretsrc cmd/secretsrc/main.go

# Run the app
./secretsrc
```

### Install via Go

```bash
go install github.com/benjamingriff/secretsrc/cmd/secretsrc@latest
```

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
- `enter` - View secret details
- `p` - Switch AWS profile
- `g` - Switch AWS region
- `r` - Refresh secret list
- `n` - Load next page (when available)
- `?` - Toggle help
- `q` / `esc` - Quit

#### Secret Detail Screen
- `v` - View secret value (decrypt and display)
- `c` - Copy secret value to clipboard (plain text)
- `j` - Copy secret value to clipboard (JSON formatted)
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
5. **Copy to Clipboard**: Press `c` for plain text or `j` for JSON-formatted copy

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
- Try switching to a different region (future feature)
- Verify that secrets exist in the current region via AWS Console

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
- [ ] Search/filter secrets
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
