package detect

import (
	"testing"
	"testing/fstest"
)

func TestDetect_NodeVariants(t *testing.T) {
	cases := []struct {
		name     string
		files    map[string]string
		wantTool string
	}{
		{"npm", map[string]string{"package.json": "{}", "package-lock.json": ""}, "npm"},
		{"yarn", map[string]string{"package.json": "{}", "yarn.lock": ""}, "yarn"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{}
			for n, c := range tc.files {
				fsys[n] = &fstest.MapFile{Data: []byte(c)}
			}
			got, err := Detect(fsys)
			if err != nil {
				t.Fatalf("Detect() err=%v", err)
			}
			if got.Tool != tc.wantTool {
				t.Errorf("got tool %q, want %q", got.Tool, tc.wantTool)
			}
		})
	}
}
