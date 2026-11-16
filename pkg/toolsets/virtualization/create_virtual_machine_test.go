package virtualization

import "testing"

func TestQualifyNetworkName(t *testing.T) {
	tests := []struct {
		name        string
		ns          string
		networkName string
		want        string
	}{
		{name: "bare name defaults to the VM namespace", ns: "demo", networkName: "vlan1", want: "demo/vlan1"},
		{name: "namespaced name in another namespace is kept", ns: "demo", networkName: "default/vlan1", want: "default/vlan1"},
		{name: "namespaced name in the VM namespace is kept", ns: "demo", networkName: "demo/vlan1", want: "demo/vlan1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qualifyNetworkName(tt.ns, tt.networkName); got != tt.want {
				t.Errorf("qualifyNetworkName(%q, %q) = %q, want %q", tt.ns, tt.networkName, got, tt.want)
			}
		})
	}
}
