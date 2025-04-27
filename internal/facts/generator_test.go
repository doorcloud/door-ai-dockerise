package facts

import (
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/stretchr/testify/assert"
)

func TestBasicGenerator_Generate(t *testing.T) {
	generator := NewBasicGenerator()

	// Create a sample detect result
	detectResult := detect.NewResult()
	detectResult.StackName = "java"
	detectResult.StackVersion = "17"
	detectResult.BuildTool = "maven"
	detectResult.BuildCommand = "mvn clean package"
	detectResult.Artifact = "target/app.jar"
	detectResult.Ports = []int{8080}
	detectResult.HealthCheck = "/health"
	detectResult.Environment = map[string]string{
		"SPRING_PROFILES_ACTIVE": "prod",
	}
	detectResult.Dependencies = []string{"spring-boot-starter-web"}
	detectResult.Metadata = map[string]interface{}{
		"framework": "spring-boot",
	}

	// Generate facts
	facts, err := generator.Generate(detectResult)
	assert.NoError(t, err)

	// Verify facts
	assert.Equal(t, "java", facts.Language)
	assert.Equal(t, "17", facts.LanguageVersion)
	assert.Equal(t, "maven", facts.BuildTool)
	assert.Equal(t, "mvn clean package", facts.BuildCommand)
	assert.Equal(t, "target/app.jar", facts.Artifact)
	assert.Equal(t, []int{8080}, facts.Ports)
	assert.Equal(t, "/health", facts.HealthCheck)
	assert.Equal(t, "prod", facts.Environment["SPRING_PROFILES_ACTIVE"])
	assert.Equal(t, []string{"spring-boot-starter-web"}, facts.Dependencies)
	assert.Equal(t, "spring-boot", facts.Metadata["framework"])
}
