package rules

import (
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/stretchr/testify/assert"
)

func TestGetFacts_SpringBootGradle(t *testing.T) {
	fsys := fstest.MapFS{}
	facts, err := GetFacts(fsys, detect.RuleInfo{
		Name: "spring-boot",
		Tool: "gradle",
	})
	assert.NoError(t, err)
	assert.Equal(t, "java", facts.Language)
	assert.Equal(t, "spring-boot", facts.Framework)
	assert.Equal(t, "gradle", facts.BuildTool)
	assert.Equal(t, "./gradlew bootJar -x test", facts.BuildCmd)
	assert.Equal(t, "build/libs/*.jar", facts.Artifact)
	assert.Equal(t, []int{8080}, facts.Ports)
}

func TestGetFacts_NodePnpm(t *testing.T) {
	fsys := fstest.MapFS{}
	facts, err := GetFacts(fsys, detect.RuleInfo{
		Name: "node",
		Tool: "pnpm",
	})
	assert.NoError(t, err)
	assert.Equal(t, "javascript", facts.Language)
	assert.Equal(t, "node", facts.Framework)
	assert.Equal(t, "pnpm", facts.BuildTool)
	assert.Equal(t, "pnpm install --frozen-lockfile && pnpm run build", facts.BuildCmd)
	assert.Equal(t, "dist/**", facts.Artifact)
	assert.Equal(t, []int{3000}, facts.Ports)
}
