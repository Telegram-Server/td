package semantic

import (
	"bytes"
	"hash/crc32"
	"os"
	"testing"

	"github.com/gotd/tl"
)

// telesrvMethods is the TELESRV_EXTENSIONS.md contract: each wire ID is the TL
// CRC32 of the declaration below, independently of the IDs in telesrv.tl.
var telesrvMethods = []struct {
	name        string
	declaration string
}{
	{"telesrv.getLoginPassword", "telesrv.getLoginPassword flags:# phone_number:string = account.Password"},
	{"telesrv.signInWithPassword", "telesrv.signInWithPassword flags:# phone_number:string password:InputCheckPasswordSRP = auth.Authorization"},
	{"telesrv.signUpWithPassword", "telesrv.signUpWithPassword flags:# phone_number:string first_name:string last_name:string new_settings:account.PasswordInputSettings = auth.Authorization"},
	{"telesrv.getMyLoginPassword", "telesrv.getMyLoginPassword flags:# = account.Password"},
	{"telesrv.updateLoginPassword", "telesrv.updateLoginPassword flags:# current:InputCheckPasswordSRP new_settings:account.PasswordInputSettings = Bool"},
}

func TestTelesrvMethodsInEveryProfile(t *testing.T) {
	universe, err := LoadUniverse("../../_schema/layers/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range telesrvMethods {
		want := crc32.ChecksumIEEE([]byte(method.declaration))
		key := SemanticKey{Category: CategoryFunction, QName: method.name}
		for _, layer := range universe.Layers() {
			definition, ok := universe.Schemas[layer].ByKey[key]
			if !ok {
				t.Errorf("layer %d: %s is missing", layer, method.name)
				continue
			}
			if definition.WireID != want {
				t.Errorf("layer %d: %s wire ID = %#08x, want CRC32 %#08x", layer, method.name, definition.WireID, want)
			}
		}
	}
}

func TestTelesrvWireIDsDoNotCollideOutsideProfiles(t *testing.T) {
	universe, err := LoadUniverse("../../_schema/layers/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	others := map[uint32]string{}
	for _, file := range []string{"../../_schema/mt.tl", "../../_schema/encrypted.tl"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		schema, err := tl.Parse(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, definition := range schema.Definitions {
			others[definition.Definition.ID] = file + " " + qualifyDefinition(definition.Definition)
		}
	}
	for _, overlay := range universe.ClientRPCOverlays {
		for _, method := range overlay.Methods {
			others[method.Definition.WireID] = overlay.Name + " " + method.Definition.Key.QName
		}
	}
	for _, method := range telesrvMethods {
		id := crc32.ChecksumIEEE([]byte(method.declaration))
		if other, exists := others[id]; exists {
			t.Errorf("%s wire ID %#08x collides with %s", method.name, id, other)
		}
	}
}
