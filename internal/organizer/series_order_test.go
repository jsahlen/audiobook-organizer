package organizer

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMetadataGetFullValidSeriesSourceOrder(t *testing.T) {
	sourceOrder := FieldMapping{SeriesOrder: SeriesOrderSource}

	tests := []struct {
		name     string
		metadata Metadata
		want     string
	}{
		{
			name: "multiple_series_keeps_source_order",
			metadata: Metadata{
				Series:       []string{"Z Series #3", "A Series #1", "M Series #2"},
				fieldMapping: sourceOrder,
			},
			want: "Z Series #3",
		},
		{
			name: "explicit_alphabetical_sorts",
			metadata: Metadata{
				Series:       []string{"Z Series #3", "A Series #1"},
				fieldMapping: FieldMapping{SeriesOrder: SeriesOrderAlphabetical},
			},
			want: "A Series #1",
		},
		{
			name: "invalid_first_series",
			metadata: Metadata{
				Series:       []string{InvalidSeriesValue, "Valid Series #1"},
				fieldMapping: sourceOrder,
			},
			want: "",
		},
		{
			name: "empty_series",
			metadata: Metadata{
				Series:       []string{},
				fieldMapping: sourceOrder,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metadata.GetFullValidSeries()
			if got != tt.want {
				t.Errorf("GetFullValidSeries() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyFieldMappingSourceSeriesOrder(t *testing.T) {
	metadata := Metadata{
		Title:   "Secret History",
		Authors: []string{"Brandon Sanderson"},
		Series:  []string{"The Mistborn Saga #3.5", "The Cosmere"},
		RawData: map[string]interface{}{},
	}
	mapping := DefaultFieldMapping()
	mapping.SeriesOrder = SeriesOrderSource
	metadata.ApplyFieldMapping(mapping)

	if got := metadata.GetValidSeries(); got != "The Mistborn Saga" {
		t.Errorf("GetValidSeries() = %q, want %q", got, "The Mistborn Saga")
	}
	if got := GetSeriesNumberFromMetadata(metadata); got != "3.5" {
		t.Errorf("GetSeriesNumberFromMetadata() = %q, want %q", got, "3.5")
	}
}

func TestLayoutTemplateSourceSeriesOrder(t *testing.T) {
	config := &OrganizerConfig{
		BaseDir:        "testbase",
		LayoutTemplate: "{author}/{series}/{series_number. }{title}",
	}
	calculator := NewLayoutCalculator(config, func(s string) string { return s })

	metadata := Metadata{
		Title:   "Secret History",
		Authors: []string{"Brandon Sanderson"},
		Series:  []string{"The Mistborn Saga #3.5", "The Cosmere"},
		RawData: map[string]interface{}{},
	}
	mapping := DefaultFieldMapping()
	mapping.SeriesOrder = SeriesOrderSource
	metadata.ApplyFieldMapping(mapping)

	got, err := calculator.CalculateTargetPathE(metadata)
	if err != nil {
		t.Fatalf("CalculateTargetPathE() error = %v", err)
	}
	want := filepath.Join("testbase", "Brandon Sanderson", "The Mistborn Saga", "3.5. Secret History")
	if got != want {
		t.Errorf("CalculateTargetPathE() = %q, want %q", got, want)
	}
}

func TestOrganizerConfigValidateSeriesOrder(t *testing.T) {
	tests := []struct {
		order     string
		wantError bool
	}{
		{order: "", wantError: false},
		{order: SeriesOrderAlphabetical, wantError: false},
		{order: SeriesOrderSource, wantError: false},
		{order: "first", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.order, func(t *testing.T) {
			config := &OrganizerConfig{
				BaseDir:      t.TempDir(),
				FieldMapping: FieldMapping{SeriesOrder: tt.order},
			}
			err := config.Validate()
			if (err != nil) != tt.wantError {
				t.Fatalf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil && !strings.Contains(err.Error(), "invalid series order") {
				t.Errorf("Validate() error = %q, want it to mention the series order", err)
			}
		})
	}
}
