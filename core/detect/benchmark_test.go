package detect

import (
	"context"
	"os"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/spring"
)

func BenchmarkDetectSpringBoot(b *testing.B) {
	// Create test filesystems for different Spring Boot projects
	fsystems := []struct {
		name string
		path string
	}{
		{"simple-maven", "testdata/spring/maven-single"},
		{"simple-gradle", "testdata/spring/gradle-groovy"},
		{"deep-nested", "testdata/sb_maven_parent/child/grandchild"},
		{"mixed-builders", "testdata/mixed_builders"},
		{"kotlin-dsl", "testdata/spring/gradle-kotlin"},
	}

	detector := spring.NewSpringBootDetectorV3()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, fs := range fsystems {
			fsys := os.DirFS(fs.path)
			detector.Detect(ctx, fsys, nil)
		}
	}
}
