package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/1it/go-submission-service/internal/config"
)

// Manager handles email template loading, validation, and rendering
type Manager struct {
	config      *config.Config
	templates   map[string]*template.Template
	initialized bool
}

// NewManager creates a new template manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:    cfg,
		templates: make(map[string]*template.Template),
	}
}

// Initialize loads and parses all templates in the templates directory
func (m *Manager) Initialize() error {
	if m.initialized {
		return nil
	}

	templatesDir := m.config.Email.Templates.Directory
	if templatesDir == "" {
		return fmt.Errorf("templates directory not specified in configuration")
	}

	// Check if directory exists
	info, err := os.Stat(templatesDir)
	if err != nil {
		return fmt.Errorf("failed to access templates directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("templates path is not a directory: %s", templatesDir)
	}

	// Read all HTML files in the directory
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	// Parse each template
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		templatePath := filepath.Join(templatesDir, entry.Name())
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", entry.Name(), err)
		}

		m.templates[entry.Name()] = tmpl
	}

	if len(m.templates) == 0 {
		return fmt.Errorf("no templates found in directory: %s", templatesDir)
	}

	m.initialized = true
	return nil
}

// GetTemplate returns a template by name
func (m *Manager) GetTemplate(name string) (*template.Template, error) {
	if !m.initialized {
		if err := m.Initialize(); err != nil {
			return nil, err
		}
	}

	tmpl, ok := m.templates[name]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return tmpl, nil
}

// GetDefaultTemplate returns the default template
func (m *Manager) GetDefaultTemplate() (*template.Template, error) {
	defaultTemplate := m.config.Email.Templates.DefaultTemplate
	if defaultTemplate == "" {
		return nil, fmt.Errorf("default template not specified in configuration")
	}

	return m.GetTemplate(defaultTemplate)
}

// RenderEmail renders an email template with the given data
func (m *Manager) RenderEmail(templateName string, data map[string]interface{}) (string, error) {
	// If no template name is provided, use the default
	if templateName == "" {
		templateName = m.config.Email.Templates.DefaultTemplate
	}

	// Get the template
	tmpl, err := m.GetTemplate(templateName)
	if err != nil {
		return "", err
	}

	// Add standard variables if not already present
	dataWithDefaults := make(map[string]interface{})
	for k, v := range data {
		dataWithDefaults[k] = v
	}

	// Add timestamp if not present
	if _, ok := dataWithDefaults["timestamp"]; !ok {
		dataWithDefaults["timestamp"] = time.Now().Format(time.RFC3339)
	}

	// Add any configured template variables
	for k, v := range m.config.Email.Templates.Variables {
		if _, ok := dataWithDefaults[k]; !ok {
			dataWithDefaults[k] = v
		}
	}

	// Render the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, dataWithDefaults); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// ValidateTemplate checks if a template exists and can be parsed
func (m *Manager) ValidateTemplate(templateName string) error {
	if !m.initialized {
		if err := m.Initialize(); err != nil {
			return err
		}
	}

	_, err := m.GetTemplate(templateName)
	return err
}

// ListTemplates returns a list of available template names
func (m *Manager) ListTemplates() []string {
	if !m.initialized {
		if err := m.Initialize(); err != nil {
			return nil
		}
	}

	templates := make([]string, 0, len(m.templates))
	for name := range m.templates {
		templates = append(templates, name)
	}
	return templates
}
