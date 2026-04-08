package aws

import "testing"

func TestGetDefaultRegion(t *testing.T) {
	t.Run("prefers AWS_REGION", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-west-2")
		t.Setenv("AWS_DEFAULT_REGION", "eu-west-1")

		if region := GetDefaultRegion(); region != "us-west-2" {
			t.Fatalf("expected AWS_REGION to win, got %q", region)
		}
	})

	t.Run("falls back to AWS_DEFAULT_REGION", func(t *testing.T) {
		t.Setenv("AWS_REGION", "")
		t.Setenv("AWS_DEFAULT_REGION", "eu-west-1")

		if region := GetDefaultRegion(); region != "eu-west-1" {
			t.Fatalf("expected AWS_DEFAULT_REGION fallback, got %q", region)
		}
	})

	t.Run("returns empty when no env override exists", func(t *testing.T) {
		t.Setenv("AWS_REGION", "")
		t.Setenv("AWS_DEFAULT_REGION", "")

		if region := GetDefaultRegion(); region != "" {
			t.Fatalf("expected empty region override, got %q", region)
		}
	})
}
