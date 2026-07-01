package aws

import (
	"reflect"
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/benjamingriff/secretsrc/pkg/models"
)

func TestEntryFromParameterMetadata(t *testing.T) {
	modified := time.Date(2026, time.July, 1, 9, 30, 0, 0, time.UTC)

	md := ssmtypes.ParameterMetadata{
		Name:             awssdk.String("/myapp/prod/db-url"),
		ARN:              awssdk.String("arn:aws:ssm:eu-west-2:123456789012:parameter/myapp/prod/db-url"),
		Description:      awssdk.String("Primary Postgres URL"),
		Type:             ssmtypes.ParameterTypeSecureString,
		Version:          7,
		LastModifiedDate: &modified,
	}

	got := entryFromParameterMetadata(md)
	want := models.Entry{
		Kind:             models.KindParameter,
		Name:             "/myapp/prod/db-url",
		ARN:              "arn:aws:ssm:eu-west-2:123456789012:parameter/myapp/prod/db-url",
		Description:      "Primary Postgres URL",
		Type:             "SecureString",
		Version:          "7",
		LastModifiedDate: &modified,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entryFromParameterMetadata() = %+v, want %+v", got, want)
	}
}

func TestEntryFromParameterMetadataHandlesNilFields(t *testing.T) {
	got := entryFromParameterMetadata(ssmtypes.ParameterMetadata{})

	want := models.Entry{
		Kind:    models.KindParameter,
		Type:    "",
		Version: "0",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entryFromParameterMetadata(empty) = %+v, want %+v", got, want)
	}
}
