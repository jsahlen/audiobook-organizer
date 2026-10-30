package cmd

import "testing"

func TestRootCommandIncludesSeriesOrderFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("series-order")
	if flag == nil {
		t.Fatal("root command missing series-order flag")
	}
	if flag.DefValue != "" {
		t.Errorf("series-order default = %q, want empty", flag.DefValue)
	}

	aliases := envAliases["series-order"]
	want := []string{"AO_SERIES_ORDER", "AUDIOBOOK_ORGANIZER_SERIES_ORDER"}
	if len(aliases) != len(want) {
		t.Fatalf("series-order aliases = %v, want %v", aliases, want)
	}
	for i := range want {
		if aliases[i] != want[i] {
			t.Errorf("series-order alias %d = %q, want %q", i, aliases[i], want[i])
		}
	}
}
