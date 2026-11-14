package model

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"go.yaml.in/yaml/v3"
)

// Keep the published list-only DTO in sync with its OpenAPI contract, including
// explicit nulls. Detail/profile endpoints continue using MemberResponse.
func TestMemberAccountOpenAPIContract(t *testing.T) {
	data, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	type property struct {
		Type     string `yaml:"type"`
		Nullable bool   `yaml:"nullable"`
	}
	var doc struct {
		Components struct {
			Schemas map[string]struct {
				Required   []string            `yaml:"required"`
				Properties map[string]property `yaml:"properties"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	schema := doc.Components.Schemas["MemberAccountResponse"]
	dto := reflect.TypeOf(MemberAccountResponse{})
	if len(schema.Properties) != dto.NumField() || len(schema.Required) != dto.NumField() {
		t.Fatal("OpenAPI must describe and require every account response field")
	}
	required := map[string]bool{}
	for _, name := range schema.Required {
		required[name] = true
	}
	for i := 0; i < dto.NumField(); i++ {
		field := dto.Field(i)
		name := field.Tag.Get("json")
		prop, ok := schema.Properties[name]
		if !ok || !required[name] {
			t.Fatalf("missing property %s", name)
		}
		typ := field.Type
		nullable := typ.Kind() == reflect.Pointer
		if nullable {
			typ = typ.Elem()
		}
		wantType := map[reflect.Kind]string{reflect.String: "string", reflect.Int: "integer", reflect.Bool: "boolean"}[typ.Kind()]
		if prop.Type != wantType || prop.Nullable != nullable {
			t.Fatalf("property %s: OpenAPI=%+v Go=%v", name, prop, field.Type)
		}
	}
	id, nis, xp := "10", "2024001", 0
	for _, account := range []MemberAccountResponse{
		{UserId: "1", HasProfile: true, MemberId: &id, NIS: &nis, TotalXP: &xp, Level: &xp, Username: "student", Email: "student@example.com"},
		{UserId: "2", Username: "new", Email: "new@example.com"},
	} {
		encoded, err := json.Marshal(account)
		if err != nil {
			t.Fatal(err)
		}
		var row map[string]any
		if err := json.Unmarshal(encoded, &row); err != nil {
			t.Fatal(err)
		}
		for name, prop := range schema.Properties {
			value, ok := row[name]
			if !ok || (value == nil && !prop.Nullable) {
				t.Errorf("invalid %s in %s", name, encoded)
			}
		}
		if account.HasProfile && (row["member_id"] != "10" || row["total_xp"] != float64(0)) {
			t.Fatal("profile values lost")
		}
		if !account.HasProfile && row["member_id"] != nil {
			t.Fatal("missing profile ID must be null")
		}
	}
	if doc.Components.Schemas["MemberResponse"].Properties["member_id"].Nullable {
		t.Fatal("detail/profile ID contract must remain non-nullable")
	}
}
