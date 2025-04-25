package test

import (
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/stretchr/testify/assert"
)

func TestNodeJS_Detect(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected bool
	}{
		{
			name: "express dependency",
			files: map[string]string{
				"package.json": `{
					"dependencies": {
						"express": "^4.17.1"
					}
				}`,
			},
			expected: true,
		},
		{
			name: "koa dependency",
			files: map[string]string{
				"package.json": `{
					"dependencies": {
						"koa": "^2.13.0"
					}
				}`,
			},
			expected: true,
		},
		{
			name: "express dev dependency",
			files: map[string]string{
				"package.json": `{
					"devDependencies": {
						"express": "^4.17.1"
					}
				}`,
			},
			expected: true,
		},
		{
			name: "start script",
			files: map[string]string{
				"package.json": `{
					"scripts": {
						"start": "node server.js"
					}
				}`,
			},
			expected: true,
		},
		{
			name: "dev script",
			files: map[string]string{
				"package.json": `{
					"scripts": {
						"dev": "nodemon server.js"
					}
				}`,
			},
			expected: true,
		},
		{
			name:     "no package.json",
			files:    map[string]string{},
			expected: false,
		},
		{
			name: "no relevant dependencies or scripts",
			files: map[string]string{
				"package.json": `{
					"dependencies": {
						"lodash": "^4.17.21"
					},
					"scripts": {
						"test": "jest"
					}
				}`,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{}
			for name, content := range tt.files {
				fsys[name] = &fstest.MapFile{
					Data: []byte(content),
				}
			}

			detector := &rules.NodeJS{}
			detected, err := detector.Detect(fsys)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, detected)
		})
	}
}
