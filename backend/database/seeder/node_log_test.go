package seeder

import (
	"reflect"
	"strings"
	"testing"
)

func TestNodeLogPermissionSeedDefinitions(t *testing.T) {
	data, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := make(map[string]string)
	for _, permission := range data.Permissions {
		if !strings.HasPrefix(permission.Name, "node_log:") {
			continue
		}
		if _, exists := got[permission.Name]; exists {
			t.Fatalf("duplicate node-log permission %q", permission.Name)
		}
		got[permission.Name] = permission.Description
	}
	want := map[string]string{
		"node_log:get":    "View node logs.",
		"node_log:remove": "Delete node logs.",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("node-log permissions = %#v, want %#v", got, want)
	}
}

func TestNodeLogRolePermissionSeedMatrix(t *testing.T) {
	data, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := make(map[string][]string)
	for _, role := range data.Roles {
		if _, exists := got[role.Name]; exists {
			t.Fatalf("duplicate role %q", role.Name)
		}
		permissions := make([]string, 0, 2)
		for _, permission := range role.Permissions {
			if strings.HasPrefix(permission, "node_log:") {
				permissions = append(permissions, permission)
			}
		}
		got[role.Name] = permissions
	}
	want := map[string][]string{
		"super": {"node_log:get", "node_log:remove"},
		"admin": {"node_log:get", "node_log:remove"},
		"user":  {"node_log:get"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("node-log role permission matrix = %#v, want %#v", got, want)
	}
}
