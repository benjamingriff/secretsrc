package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/benjamingriff/secretsrc/pkg/aws"
	"github.com/benjamingriff/secretsrc/pkg/config"
	"github.com/benjamingriff/secretsrc/pkg/models"
	"github.com/benjamingriff/secretsrc/pkg/ui/components"
)

// Screen represents the different screens in the app
type Screen int

const (
	ScreenSecretList Screen = iota
	ScreenSecretDetail
	ScreenProfileSelector
	ScreenRegionSelector
	ScreenMFAInput
)

// Model is the main Bubble Tea model
type Model struct {
	// Current screen
	currentScreen Screen

	// AWS client and state
	awsClient      *aws.Client
	currentProfile string
	currentRegion  string

	// Secret data
	secrets       []models.Secret
	selectedIndex int
	secretValue   string
	nextToken     *string
	hasMore       bool

	// Pagination state
	pageHistory []secretPage // History of loaded pages
	currentPage int          // Current page index in history

	// UI components
	grid            components.SecretGrid
	profileSelector components.ProfileSelector
	regionSelector  components.RegionSelector
	mfaInput        components.MFAInput
	keys            KeyMap

	// MFA state
	pendingMFAProfile       string
	pendingMFARegion        string
	pendingMFASourceProfile string
	mfaSerial               string

	// UI state
	loading       bool
	errorMessage  string
	statusMessage string
	width         int
	height        int
	showHelp      bool
}

// secretPage represents a page of secrets
type secretPage struct {
	secrets   []models.Secret
	nextToken *string
}

// Custom messages
type secretsLoadedMsg struct {
	secrets   []models.Secret
	nextToken *string
	err       error
}

type secretValueLoadedMsg struct {
	value string
	err   error
}

type clientChangedMsg struct {
	client  *aws.Client
	profile string
	region  string
	err     error
}

type clearStatusMsg struct{}

type clipboardCopiedMsg struct {
	success bool
	err     error
}

type mfaRequiredMsg struct {
	profile       string
	region        string
	mfaSerial     string
	sourceProfile string
}

type mfaTokenSubmittedMsg struct {
	creds awssdk.Credentials
	err   error
}

// NewModel creates a new app model
func NewModel(profile, region string) Model {
	return Model{
		currentScreen:  ScreenSecretList,
		currentProfile: profile,
		currentRegion:  region,
		keys:           DefaultKeyMap(),
		grid:           components.NewSecretGrid(80, 20),
		loading:        true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		initAWSClient(m.currentProfile, m.currentRegion),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Account for border (2 chars), header (~3 lines), footer (~3 lines), and padding
		contentHeight := msg.Height - 10
		contentWidth := msg.Width - 6

		m.grid.SetSize(contentWidth, contentHeight)
		// Only resize selectors if they're initialized (i.e., we're on their screen)
		if m.currentScreen == ScreenProfileSelector {
			m.profileSelector.SetSize(contentWidth, contentHeight)
		}
		if m.currentScreen == ScreenRegionSelector {
			m.regionSelector.SetSize(contentWidth, contentHeight)
		}
		return m, nil

	case tea.KeyMsg:
		// Global keys
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle keys based on current screen
		switch m.currentScreen {
		case ScreenSecretList:
			return m.handleSecretListKeys(msg)
		case ScreenSecretDetail:
			return m.handleSecretDetailKeys(msg)
		case ScreenProfileSelector:
			return m.handleProfileSelectorKeys(msg)
		case ScreenRegionSelector:
			return m.handleRegionSelectorKeys(msg)
		case ScreenMFAInput:
			return m.handleMFAInputKeys(msg)
		}

	case mfaRequiredMsg:
		// MFA is required, show input screen
		m.pendingMFAProfile = msg.profile
		m.pendingMFARegion = msg.region
		m.mfaSerial = msg.mfaSerial
		m.pendingMFASourceProfile = msg.sourceProfile
		m.mfaInput = components.NewMFAInput()
		m.currentScreen = ScreenMFAInput
		m.loading = false
		return m, nil

	case mfaTokenSubmittedMsg:
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("MFA authentication failed: %v", msg.err)
			m.loading = false
			// Stay on MFA screen so user can try again
			m.mfaInput.Reset()
			return m, nil
		}
		// MFA successful, cache the credentials
		profileForCache := m.pendingMFAProfile
		if m.pendingMFASourceProfile != "" {
			profileForCache = m.pendingMFASourceProfile
		}

		go func() {
			cachedCreds := config.CachedCredentials{
				AccessKeyID:     msg.creds.AccessKeyID,
				SecretAccessKey: msg.creds.SecretAccessKey,
				SessionToken:    msg.creds.SessionToken,
				ExpiresAt:       msg.creds.Expires,
			}
			_ = config.SaveCachedCredentials(profileForCache, cachedCreds) // Ignore errors
		}()

		// Create client with credentials
		m.currentScreen = ScreenSecretList
		m.loading = true
		return m, createClientWithMFACredentials(m.pendingMFAProfile, m.pendingMFARegion, msg.creds, m.pendingMFASourceProfile)

	case clientChangedMsg:
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Failed to initialize AWS client: %v", msg.err)
			m.loading = false
			return m, nil
		}
		m.awsClient = msg.client
		m.currentProfile = msg.profile
		m.currentRegion = msg.region
		m.loading = true

		// Save profile and region to config for next time
		go func() {
			cfg := &config.Config{
				LastProfile: msg.profile,
				LastRegion:  msg.region,
			}
			_ = config.Save(cfg) // Ignore errors, don't block UI
		}()

		return m, loadSecrets(m.awsClient, 50, nil)

	case secretsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Failed to load secrets: %v", msg.err)
			return m, nil
		}
		m.secrets = msg.secrets
		m.nextToken = msg.nextToken
		m.hasMore = msg.nextToken != nil
		m.grid.SetSecrets(m.secrets)
		m.errorMessage = ""

		// Update page history for the current page
		if m.currentPage < len(m.pageHistory) {
			// Updating existing page
			m.pageHistory[m.currentPage] = secretPage{
				secrets:   msg.secrets,
				nextToken: msg.nextToken,
			}
		} else {
			// New page, add to history
			m.pageHistory = append(m.pageHistory, secretPage{
				secrets:   msg.secrets,
				nextToken: msg.nextToken,
			})
		}

		return m, nil

	case secretValueLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Failed to load secret value: %v", msg.err)
			return m, nil
		}
		m.secretValue = msg.value
		m.errorMessage = ""
		return m, nil

	case clearStatusMsg:
		m.statusMessage = ""
		return m, nil

	case clipboardCopiedMsg:
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Failed to copy to clipboard: %v", msg.err)
		} else if msg.success {
			m.statusMessage = "Copied to clipboard!"
			return m, clearStatusAfter(2 * time.Second)
		}
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// handleSecretListKeys handles key presses on the secret list screen
func (m Model) handleSecretListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit

	case "enter":
		// View secret details
		secret := m.grid.SelectedSecret()
		if secret != nil {
			m.currentScreen = ScreenSecretDetail
			m.secretValue = "" // Clear previous value
		}
		return m, nil

	case "r":
		// Refresh secrets - clear pagination history
		m.loading = true
		m.nextToken = nil
		m.pageHistory = nil
		m.currentPage = 0
		return m, loadSecrets(m.awsClient, 50, nil)

	case "n":
		// Load next page
		if m.hasMore {
			m.currentPage++
			// Check if we already have this page in history
			if m.currentPage < len(m.pageHistory) {
				// Load from history
				page := m.pageHistory[m.currentPage]
				m.secrets = page.secrets
				m.nextToken = page.nextToken
				m.hasMore = page.nextToken != nil
				m.grid.SetSecrets(m.secrets)
				return m, nil
			}
			// Need to fetch new page
			m.loading = true
			return m, loadSecrets(m.awsClient, 50, m.nextToken)
		}
		return m, nil

	case "b":
		// Go to previous page
		if m.currentPage > 0 {
			m.currentPage--
			page := m.pageHistory[m.currentPage]
			m.secrets = page.secrets
			m.nextToken = page.nextToken
			m.hasMore = page.nextToken != nil || m.currentPage < len(m.pageHistory)-1
			m.grid.SetSecrets(m.secrets)
		}
		return m, nil

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "p":
		// Open profile selector
		profiles, err := aws.GetAvailableProfiles()
		if err != nil {
			m.errorMessage = fmt.Sprintf("Failed to load profiles: %v", err)
			return m, nil
		}
		m.profileSelector = components.NewProfileSelector(profiles, m.currentProfile, m.width, m.height-6)
		m.currentScreen = ScreenProfileSelector
		return m, nil

	case "g":
		// Open region selector
		regions := aws.GetCommonRegions()
		m.regionSelector = components.NewRegionSelector(regions, m.currentRegion, m.width, m.height-6)
		m.currentScreen = ScreenRegionSelector
		return m, nil
	}

	// Let the grid handle navigation and filter keys
	cmd := m.grid.Update(msg)
	return m, cmd
}

// handleSecretDetailKeys handles key presses on the secret detail screen
func (m Model) handleSecretDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Go back to list
		m.currentScreen = ScreenSecretList
		m.secretValue = "" // Clear secret value from memory
		return m, nil

	case "v":
		// View secret value
		secret := m.grid.SelectedSecret()
		if secret != nil && m.secretValue == "" {
			m.loading = true
			return m, loadSecretValue(m.awsClient, secret.Name)
		}
		return m, nil

	case "c":
		// Copy plain text
		if m.secretValue != "" {
			return m, copyToClipboard(m.secretValue, false)
		}
		return m, nil

	case "j":
		// Copy JSON formatted
		if m.secretValue != "" {
			return m, copyToClipboard(m.secretValue, true)
		}
		return m, nil
	}

	return m, nil
}

// handleProfileSelectorKeys handles key presses on the profile selector screen
func (m Model) handleProfileSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Go back to list
		m.currentScreen = ScreenSecretList
		return m, nil

	case "enter":
		// Select profile
		selectedProfile := m.profileSelector.SelectedProfile()
		if selectedProfile != "" && selectedProfile != m.currentProfile {
			// Profile changed, reinitialize client
			m.loading = true
			m.currentScreen = ScreenSecretList
			return m, initAWSClient(selectedProfile, m.currentRegion)
		}
		// No change, just go back
		m.currentScreen = ScreenSecretList
		return m, nil
	}

	// Let the profile selector handle navigation keys
	cmd := m.profileSelector.Update(msg)
	return m, cmd
}

// handleRegionSelectorKeys handles key presses on the region selector screen
func (m Model) handleRegionSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Go back to list
		m.currentScreen = ScreenSecretList
		return m, nil

	case "enter":
		// Select region
		selectedRegion := m.regionSelector.SelectedRegion()
		if selectedRegion != "" && selectedRegion != m.currentRegion {
			// Region changed, reinitialize client
			m.loading = true
			m.currentScreen = ScreenSecretList
			return m, initAWSClient(m.currentProfile, selectedRegion)
		}
		// No change, just go back
		m.currentScreen = ScreenSecretList
		return m, nil
	}

	// Let the region selector handle navigation keys
	cmd := m.regionSelector.Update(msg)
	return m, cmd
}

// handleMFAInputKeys handles key presses on the MFA input screen
func (m Model) handleMFAInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel MFA input, go back to list
		m.currentScreen = ScreenSecretList
		m.errorMessage = "MFA authentication cancelled"
		return m, nil

	case "enter":
		// Submit MFA token
		token := m.mfaInput.Value()
		if len(token) != 6 {
			m.errorMessage = "MFA code must be 6 digits"
			return m, nil
		}
		m.loading = true
		m.errorMessage = ""
		// Use source profile for MFA if this is a role assumption
		profileForMFA := m.pendingMFAProfile
		if m.pendingMFASourceProfile != "" {
			profileForMFA = m.pendingMFASourceProfile
		}
		return m, submitMFAToken(m.pendingMFAProfile, profileForMFA, m.pendingMFARegion, m.mfaSerial, token)
	}

	// Let the text input handle key presses
	cmd := m.mfaInput.Update(msg)
	return m, cmd
}

// Commands

// initAWSClient initializes the AWS client
func initAWSClient(profile, region string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Check if profile requires MFA
		mfaConfig, err := aws.GetMFAConfig(profile)
		if err == nil && mfaConfig.Required {
			// Check for cached credentials
			profileForCache := profile
			if mfaConfig.SourceProfile != "" {
				profileForCache = mfaConfig.SourceProfile
			}

			if cachedCreds, valid := config.GetCachedCredentials(profileForCache); valid {
				// Use cached credentials
				creds := awssdk.Credentials{
					AccessKeyID:     cachedCreds.AccessKeyID,
					SecretAccessKey: cachedCreds.SecretAccessKey,
					SessionToken:    cachedCreds.SessionToken,
					Source:          "CachedMFA",
					CanExpire:       true,
					Expires:         cachedCreds.ExpiresAt,
				}

				var client *aws.Client
				var clientErr error

				if mfaConfig.SourceProfile != "" {
					// Role assumption
					client, clientErr = aws.NewClientWithMFAForRole(ctx, profile, region, creds)
				} else {
					// Direct MFA
					client, clientErr = aws.NewClientWithMFA(ctx, profile, region, creds)
				}

				return clientChangedMsg{
					client:  client,
					profile: profile,
					region:  region,
					err:     clientErr,
				}
			}

			// No valid cached credentials, prompt for MFA
			return mfaRequiredMsg{
				profile:       profile,
				region:        region,
				mfaSerial:     mfaConfig.MFASerial,
				sourceProfile: mfaConfig.SourceProfile,
			}
		}

		// No MFA required or error checking, proceed normally
		client, err := aws.NewClient(ctx, profile, region)
		return clientChangedMsg{
			client:  client,
			profile: profile,
			region:  region,
			err:     err,
		}
	}
}

// loadSecrets loads secrets from AWS
func loadSecrets(client *aws.Client, maxResults int32, nextToken *string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return secretsLoadedMsg{err: fmt.Errorf("AWS client not initialized")}
		}
		ctx := context.Background()
		secrets, token, err := client.ListSecrets(ctx, maxResults, nextToken)
		return secretsLoadedMsg{
			secrets:   secrets,
			nextToken: token,
			err:       err,
		}
	}
}

// loadSecretValue loads a secret value from AWS
func loadSecretValue(client *aws.Client, secretName string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return secretValueLoadedMsg{err: fmt.Errorf("AWS client not initialized")}
		}
		ctx := context.Background()
		value, err := client.GetSecretValue(ctx, secretName)
		return secretValueLoadedMsg{
			value: value,
			err:   err,
		}
	}
}

// clearStatusAfter clears the status message after a delay
func clearStatusAfter(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// copyToClipboard copies the value to clipboard
func copyToClipboard(value string, asJSON bool) tea.Cmd {
	return func() tea.Msg {
		var toCopy string
		if asJSON {
			// Try to format as JSON
			var jsonData interface{}
			if err := json.Unmarshal([]byte(value), &jsonData); err == nil {
				prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
				if err == nil {
					toCopy = string(prettyJSON)
				} else {
					toCopy = value
				}
			} else {
				// Not valid JSON, copy as-is
				toCopy = value
			}
		} else {
			toCopy = value
		}

		err := clipboard.WriteAll(toCopy)
		return clipboardCopiedMsg{
			success: err == nil,
			err:     err,
		}
	}
}

// submitMFAToken submits the MFA token and gets session credentials
func submitMFAToken(targetProfile, profileForMFA, region, mfaSerial, token string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		creds, err := aws.GetSessionTokenWithMFA(ctx, profileForMFA, region, mfaSerial, token)
		return mfaTokenSubmittedMsg{
			creds: creds,
			err:   err,
		}
	}
}

// createClientWithMFACredentials creates an AWS client with MFA credentials
func createClientWithMFACredentials(profile, region string, creds awssdk.Credentials, sourceProfile string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		// If this profile uses a source profile (role assumption), we need to handle it differently
		var client *aws.Client
		var err error

		if sourceProfile != "" {
			// This is a role assumption profile
			client, err = aws.NewClientWithMFAForRole(ctx, profile, region, creds)
		} else {
			// Direct MFA authentication
			client, err = aws.NewClientWithMFA(ctx, profile, region, creds)
		}

		return clientChangedMsg{
			client:  client,
			profile: profile,
			region:  region,
			err:     err,
		}
	}
}
