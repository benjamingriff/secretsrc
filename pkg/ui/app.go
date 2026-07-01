package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/benjamingriff/secretsrc/pkg/aws"
	"github.com/benjamingriff/secretsrc/pkg/config"
	"github.com/benjamingriff/secretsrc/pkg/models"
	"github.com/benjamingriff/secretsrc/pkg/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents the different screens in the app
type Screen int

const (
	ScreenSecretList Screen = iota
	ScreenSecretDetail
	ScreenSecretFieldSelector
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

	// Active data source (Secrets Manager or SSM Parameter Store)
	mode models.Kind

	// Entry data
	entries       []models.Entry
	selectedIndex int
	entryValue    string
	secretFields  []components.SecretField
	nextToken     *string
	hasMore       bool

	// Pagination state
	pageHistory []secretPage // History of loaded pages
	currentPage int          // Current page index in history

	// UI components
	grid            components.Grid
	fieldSelector   components.SecretFieldSelector
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

// secretPage represents a page of entries
type secretPage struct {
	entries   []models.Entry
	nextToken *string
}

// Custom messages
type entriesLoadedMsg struct {
	entries   []models.Entry
	nextToken *string
	err       error
}

type entryValueLoadedMsg struct {
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

// Mode persistence values written to the config file.
const (
	modeStringSecrets = "secrets"
	modeStringSSM     = "ssm"
)

// ParseMode converts a persisted mode string into a models.Kind, defaulting to
// Secrets Manager for any unrecognised value.
func ParseMode(s string) models.Kind {
	if s == modeStringSSM {
		return models.KindParameter
	}
	return models.KindSecret
}

// modeString converts a models.Kind into its persisted string form.
func modeString(mode models.Kind) string {
	if mode == models.KindParameter {
		return modeStringSSM
	}
	return modeStringSecrets
}

// NewModel creates a new app model
func NewModel(profile, region string, mode models.Kind) Model {
	grid := components.NewGrid(80, 20)
	grid.SetAccentColor(accentColor(mode))

	return Model{
		currentScreen:  ScreenSecretList,
		currentProfile: profile,
		currentRegion:  region,
		mode:           mode,
		keys:           DefaultKeyMap(),
		grid:           grid,
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
		contentWidth, contentHeight := m.contentViewportSize()

		m.grid.SetSize(contentWidth, contentHeight)
		// Only resize selectors if they're initialized (i.e., we're on their screen)
		if m.currentScreen == ScreenProfileSelector {
			m.profileSelector.SetSize(contentWidth, contentHeight)
		}
		if m.currentScreen == ScreenRegionSelector {
			m.regionSelector.SetSize(contentWidth, contentHeight)
		}
		if m.currentScreen == ScreenSecretFieldSelector {
			m.fieldSelector.SetSize(contentWidth, contentHeight)
		}
		return m, nil

	case tea.KeyMsg:
		// Global keys
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Tab toggles between Secrets Manager and SSM Parameter Store from any
		// screen (the shared credentials are reused).
		if msg.String() == "tab" {
			return m.toggleMode()
		}

		// Handle keys based on current screen
		switch m.currentScreen {
		case ScreenSecretList:
			return m.handleSecretListKeys(msg)
		case ScreenSecretDetail:
			return m.handleSecretDetailKeys(msg)
		case ScreenSecretFieldSelector:
			return m.handleSecretFieldSelectorKeys(msg)
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

		// Save profile, region and mode to config for next time
		mode := m.mode
		go func() {
			cfg := &config.Config{
				LastProfile: msg.client.GetProfile(),
				LastRegion:  msg.client.GetRegion(),
				LastMode:    modeString(mode),
			}
			_ = config.Save(cfg) // Ignore errors, don't block UI
		}()

		return m, loadEntries(m.awsClient, m.mode, 50, nil)

	case entriesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Failed to load %s: %v", m.sourceNoun(true), msg.err)
			return m, nil
		}
		m.entries = msg.entries
		m.nextToken = msg.nextToken
		m.hasMore = msg.nextToken != nil
		m.grid.SetEntries(m.entries)
		m.errorMessage = ""

		// Update page history for the current page
		if m.currentPage < len(m.pageHistory) {
			// Updating existing page
			m.pageHistory[m.currentPage] = secretPage{
				entries:   msg.entries,
				nextToken: msg.nextToken,
			}
		} else {
			// New page, add to history
			m.pageHistory = append(m.pageHistory, secretPage{
				entries:   msg.entries,
				nextToken: msg.nextToken,
			})
		}

		return m, nil

	case entryValueLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Failed to load value: %v", msg.err)
			return m, nil
		}
		m.entryValue = msg.value
		// Only Secrets Manager values are parsed for top-level JSON fields;
		// SSM parameters are treated as plain strings.
		if m.mode == models.KindSecret {
			m.secretFields = parseSecretFields(msg.value)
		}
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
	if m.grid.IsFiltering() {
		cmd := m.grid.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit

	case "enter":
		// View entry details
		entry := m.grid.SelectedEntry()
		if entry != nil {
			m.currentScreen = ScreenSecretDetail
			m.clearSecretValueState()
		}
		return m, nil

	case "r":
		// Refresh entries - clear pagination history
		m.loading = true
		m.nextToken = nil
		m.pageHistory = nil
		m.currentPage = 0
		return m, loadEntries(m.awsClient, m.mode, 50, nil)

	case "n":
		// Load next page
		if m.hasMore {
			m.currentPage++
			// Check if we already have this page in history
			if m.currentPage < len(m.pageHistory) {
				// Load from history
				page := m.pageHistory[m.currentPage]
				m.entries = page.entries
				m.nextToken = page.nextToken
				m.hasMore = page.nextToken != nil
				m.grid.SetEntries(m.entries)
				return m, nil
			}
			// Need to fetch new page
			m.loading = true
			return m, loadEntries(m.awsClient, m.mode, 50, m.nextToken)
		}
		return m, nil

	case "b":
		// Go to previous page
		if m.currentPage > 0 {
			m.currentPage--
			page := m.pageHistory[m.currentPage]
			m.entries = page.entries
			m.nextToken = page.nextToken
			m.hasMore = page.nextToken != nil || m.currentPage < len(m.pageHistory)-1
			m.grid.SetEntries(m.entries)
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
		m.clearSecretValueState()
		return m, nil

	case "v":
		// View entry value
		entry := m.grid.SelectedEntry()
		if entry != nil && m.entryValue == "" {
			m.loading = true
			return m, loadEntryValue(m.awsClient, m.mode, entry.Name)
		}
		return m, nil

	case "c":
		// Copy plain text
		if m.entryValue != "" {
			return m, copyToClipboard(m.entryValue, false)
		}
		return m, nil

	case "j":
		// Copy JSON formatted (Secrets Manager only; parameters are plain strings)
		if m.mode == models.KindSecret && m.entryValue != "" {
			return m, copyToClipboard(m.entryValue, true)
		}
		return m, nil

	case "k":
		// Copy a top-level JSON field value (Secrets Manager only)
		if m.mode == models.KindSecret && len(m.secretFields) > 0 {
			contentWidth, contentHeight := m.contentViewportSize()
			m.fieldSelector = components.NewSecretFieldSelector(m.secretFields, contentWidth, contentHeight)
			m.currentScreen = ScreenSecretFieldSelector
		}
		return m, nil
	}

	return m, nil
}

// handleSecretFieldSelectorKeys handles key presses on the field selector screen.
func (m Model) handleSecretFieldSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentScreen = ScreenSecretDetail
		return m, nil

	case "enter":
		field := m.fieldSelector.SelectedField()
		m.currentScreen = ScreenSecretDetail
		if field != nil {
			return m, copyToClipboard(field.CopyValue, false)
		}
		return m, nil
	}

	cmd := m.fieldSelector.Update(msg)
	return m, cmd
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

// loadEntries loads entries from AWS for the active mode
func loadEntries(client *aws.Client, mode models.Kind, maxResults int32, nextToken *string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return entriesLoadedMsg{err: fmt.Errorf("AWS client not initialized")}
		}
		ctx := context.Background()

		var entries []models.Entry
		var token *string
		var err error
		if mode == models.KindParameter {
			entries, token, err = client.ListParameters(ctx, maxResults, nextToken)
		} else {
			entries, token, err = client.ListSecrets(ctx, maxResults, nextToken)
		}

		return entriesLoadedMsg{
			entries:   entries,
			nextToken: token,
			err:       err,
		}
	}
}

// loadEntryValue loads an entry value from AWS for the active mode
func loadEntryValue(client *aws.Client, mode models.Kind, name string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return entryValueLoadedMsg{err: fmt.Errorf("AWS client not initialized")}
		}
		ctx := context.Background()

		var value string
		var err error
		if mode == models.KindParameter {
			value, err = client.GetParameterValue(ctx, name)
		} else {
			value, err = client.GetSecretValue(ctx, name)
		}

		return entryValueLoadedMsg{
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

func (m *Model) clearSecretValueState() {
	m.entryValue = ""
	m.secretFields = nil
	m.fieldSelector = components.SecretFieldSelector{}
}

// toggleMode switches between Secrets Manager and SSM Parameter Store, resetting
// list/detail state and re-fetching from the new source. It is a no-op while an
// MFA code is being entered or before the shared client is ready.
func (m Model) toggleMode() (tea.Model, tea.Cmd) {
	if m.currentScreen == ScreenMFAInput || m.awsClient == nil {
		return m, nil
	}

	if m.mode == models.KindParameter {
		m.mode = models.KindSecret
	} else {
		m.mode = models.KindParameter
	}

	// Reset all list/detail state for the new source.
	m.currentScreen = ScreenSecretList
	m.clearSecretValueState()
	m.nextToken = nil
	m.pageHistory = nil
	m.currentPage = 0
	m.errorMessage = ""
	m.loading = true
	m.grid.SetAccentColor(accentColor(m.mode))

	// Persist the mode (with the current profile/region) for next launch.
	profile, region, mode := m.currentProfile, m.currentRegion, m.mode
	go func() {
		cfg := &config.Config{
			LastProfile: profile,
			LastRegion:  region,
			LastMode:    modeString(mode),
		}
		_ = config.Save(cfg) // Ignore errors, don't block UI
	}()

	return m, loadEntries(m.awsClient, m.mode, 50, nil)
}

// sourceNoun returns a human-readable noun for the active data source.
func (m Model) sourceNoun(plural bool) string {
	if m.mode == models.KindParameter {
		if plural {
			return "parameters"
		}
		return "parameter"
	}
	if plural {
		return "secrets"
	}
	return "secret"
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
