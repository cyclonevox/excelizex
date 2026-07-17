package style

import (
	"encoding/json"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// Style is a named excelize style definition.
type Style interface {
	Name() string
	ExcelStyle() *excelize.Style
	Append(other Style) Style
}

// DefaultStyle is a concrete Style built from an excelize.Style.
type DefaultStyle struct {
	name string
	raw  *excelize.Style
}

// New creates a named style.
func New(name string, s *excelize.Style) *DefaultStyle {
	return &DefaultStyle{name: name, raw: s}
}

func (d *DefaultStyle) Name() string { return d.name }

func (d *DefaultStyle) ExcelStyle() *excelize.Style { return d.raw }

func (d *DefaultStyle) Append(other Style) Style {
	if other == nil {
		return d
	}
	merged, err := mergeStyles(d.raw, other.ExcelStyle())
	if err != nil {
		return &DefaultStyle{name: d.name + "+" + other.Name(), raw: d.raw}
	}

	return &DefaultStyle{name: d.name + "+" + other.Name(), raw: merged}
}

func mergeStyles(a, b *excelize.Style) (*excelize.Style, error) {
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}
	ab, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	var am, bm map[string]any
	if err := json.Unmarshal(ab, &am); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bb, &bm); err != nil {
		return nil, err
	}
	for k, v := range bm {
		if v == nil {
			continue
		}
		am[k] = v
	}
	out, err := json.Marshal(am)
	if err != nil {
		return nil, err
	}
	merged := new(excelize.Style)
	if err := json.Unmarshal(out, merged); err != nil {
		return nil, err
	}

	return merged, nil
}

// Registry resolves named styles to excelize style IDs within one workbook.
type Registry struct {
	f    *excelize.File
	defs map[string]Style
	ids  map[string]int
}

// NewRegistry creates a style registry bound to a workbook file.
func NewRegistry(f *excelize.File) *Registry {
	return &Registry{
		f:    f,
		defs: make(map[string]Style),
		ids:  make(map[string]int),
	}
}

// Register adds or replaces a named style definition.
func (r *Registry) Register(s Style) error {
	if s == nil || s.Name() == "" {
		return fmt.Errorf("style: invalid style")
	}
	r.defs[s.Name()] = s

	return nil
}

// RegisterDefaults installs built-in styles used by common style tags.
func (r *Registry) RegisterDefaults() error {
	defaults := []Style{
		New("notice", &excelize.Style{
			Font:      &excelize.Font{Bold: true, Color: "#FF0000"},
			Alignment: &excelize.Alignment{WrapText: true},
			Protection: &excelize.Protection{Locked: true},
		}),
		New("header", &excelize.Style{
			Font:       &excelize.Font{Bold: true},
			Protection: &excelize.Protection{Locked: true},
		}),
		New("header-red", &excelize.Style{
			Font:       &excelize.Font{Bold: true, Color: "#FF0000"},
			Protection: &excelize.Protection{Locked: true},
		}),
		New("body", &excelize.Style{
			NumFmt:     49,
			Protection: &excelize.Protection{Locked: false},
		}),
		New("body-blue", &excelize.Style{
			NumFmt: 49,
			Fill: excelize.Fill{
				Type:    "pattern",
				Pattern: 1,
				Color:   []string{"#DAEEF3"},
			},
			Protection: &excelize.Protection{Locked: false},
		}),
		New("locked", &excelize.Style{Protection: &excelize.Protection{Locked: true}}),
		New("unlocked", &excelize.Style{Protection: &excelize.Protection{Locked: false}}),
	}
	for _, s := range defaults {
		if err := r.Register(s); err != nil {
			return err
		}
	}

	return nil
}

// Resolve combines style parts (via Append) and returns an excelize style ID.
func (r *Registry) Resolve(parts ...string) (int, error) {
	if len(parts) == 0 {
		return 0, fmt.Errorf("style: empty style list")
	}
	key := joinParts(parts)
	if id, ok := r.ids[key]; ok {
		return id, nil
	}
	var combined Style
	for _, part := range parts {
		if part == "" {
			continue
		}
		s, ok := r.defs[part]
		if !ok {
			return 0, fmt.Errorf("style: unknown %q", part)
		}
		if combined == nil {
			combined = s
		} else {
			combined = combined.Append(s)
		}
	}
	if combined == nil {
		return 0, fmt.Errorf("style: no resolvable styles in %v", parts)
	}
	id, err := r.f.NewStyle(combined.ExcelStyle())
	if err != nil {
		return 0, fmt.Errorf("style: new style: %w", err)
	}
	r.ids[key] = id

	return id, nil
}

func joinParts(parts []string) string {
	out := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		if out != "" {
			out += "+"
		}
		out += p
	}

	return out
}

// SplitRole splits a comma-separated style tag into role parts; each role may use + for Append.
func SplitRole(tag []string, role int) []string {
	if role >= len(tag) {
		return nil
	}
	raw := tag[role]
	if raw == "" {
		return nil
	}
	var parts []string
	for _, p := range splitPlus(raw) {
		if p != "" {
			parts = append(parts, p)
		}
	}

	return parts
}

func splitPlus(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '+' {
			out = append(out, trim(s[start:i]))
			start = i + 1
		}
	}
	out = append(out, trim(s[start:]))

	return out
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}

	return s
}
