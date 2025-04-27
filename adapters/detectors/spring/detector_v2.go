package spring

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// SpringBootDetectorV2 implements core.Detector for Spring Boot projects
type SpringBootDetectorV2 struct {
	logSink core.LogSink
}

// NewSpringBootDetectorV2 creates a new Spring Boot detector
func NewSpringBootDetectorV2() *SpringBootDetectorV2 {
	return &SpringBootDetectorV2{}
}

// Detect implements the Detector interface
func (d *SpringBootDetectorV2) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	d.logSink = logSink

	// Check for Maven projects
	if info, found, err := d.detectMaven(fsys); err != nil {
		return core.StackInfo{}, false, err
	} else if found {
		return info, true, nil
	}

	// Check for Gradle projects
	if info, found, err := d.detectGradle(fsys); err != nil {
		return core.StackInfo{}, false, err
	} else if found {
		return info, true, nil
	}

	return core.StackInfo{}, false, nil
}

// detectMaven checks for Maven projects (single and multi-module)
func (d *SpringBootDetectorV2) detectMaven(fsys fs.FS) (core.StackInfo, bool, error) {
	// First check for app module
	appPom, err := fs.ReadFile(fsys, "app/pom.xml")
	if err == nil {
		if containsSpringBootV2(string(appPom)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				DetectedFiles: []string{"app/pom.xml"},
			}
			d.log("detector=spring-boot found=true buildtool=maven multimodule=true")
			return info, true, nil
		}
	}

	// Then check other modules
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return core.StackInfo{}, false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "app" {
			continue
		}

		// Check each subdirectory for pom.xml
		subPom, err := fs.ReadFile(fsys, filepath.Join(entry.Name(), "pom.xml"))
		if err != nil {
			continue
		}

		if containsSpringBootV2(string(subPom)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				DetectedFiles: []string{filepath.Join(entry.Name(), "pom.xml")},
			}
			d.log("detector=spring-boot found=true buildtool=maven multimodule=true")
			return info, true, nil
		}
	}

	// Finally check root pom.xml
	pomXML, err := fs.ReadFile(fsys, "pom.xml")
	if err == nil {
		if containsSpringBootV2(string(pomXML)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				DetectedFiles: []string{"pom.xml"},
			}
			d.log("detector=spring-boot found=true buildtool=maven")
			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// detectGradle checks for Gradle projects (Groovy, Kotlin, and multi-module)
func (d *SpringBootDetectorV2) detectGradle(fsys fs.FS) (core.StackInfo, bool, error) {
	// First check for app module
	appGradle, err := fs.ReadFile(fsys, "app/build.gradle")
	if err == nil {
		if containsSpringBootV2(string(appGradle)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"app/build.gradle"},
			}
			d.log("detector=spring-boot found=true buildtool=gradle multimodule=true")
			return info, true, nil
		}
	}

	appGradleKts, err := fs.ReadFile(fsys, "app/build.gradle.kts")
	if err == nil {
		if containsSpringBootV2(string(appGradleKts)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"app/build.gradle.kts"},
			}
			d.log("detector=spring-boot found=true buildtool=gradle kotlin=true multimodule=true")
			return info, true, nil
		}
	}

	// Then check root build files
	buildGradle, err := fs.ReadFile(fsys, "build.gradle")
	if err == nil {
		if containsSpringBootV2(string(buildGradle)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"build.gradle"},
			}
			d.log("detector=spring-boot found=true buildtool=gradle")
			return info, true, nil
		}
	}

	buildGradleKts, err := fs.ReadFile(fsys, "build.gradle.kts")
	if err == nil {
		if containsSpringBootV2(string(buildGradleKts)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"build.gradle.kts"},
			}
			d.log("detector=spring-boot found=true buildtool=gradle kotlin=true")
			return info, true, nil
		}
	}

	// Finally check other modules
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return core.StackInfo{}, false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "app" {
			continue
		}

		// Check each subdirectory for build.gradle or build.gradle.kts
		subGradle, err := fs.ReadFile(fsys, filepath.Join(entry.Name(), "build.gradle"))
		if err == nil && containsSpringBootV2(string(subGradle)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{filepath.Join(entry.Name(), "build.gradle")},
			}
			d.log("detector=spring-boot found=true buildtool=gradle multimodule=true")
			return info, true, nil
		}

		subGradleKts, err := fs.ReadFile(fsys, filepath.Join(entry.Name(), "build.gradle.kts"))
		if err == nil && containsSpringBootV2(string(subGradleKts)) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{filepath.Join(entry.Name(), "build.gradle.kts")},
			}
			d.log("detector=spring-boot found=true buildtool=gradle kotlin=true multimodule=true")
			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// containsSpringBootV2 checks if the build file contains Spring Boot dependencies
func containsSpringBootV2(content string) bool {
	return strings.Contains(content, "org.springframework.boot") ||
		strings.Contains(content, "spring-boot")
}

// log sends a message to the log sink if available
func (d *SpringBootDetectorV2) log(msg string) {
	if d.logSink != nil {
		d.logSink.Log(msg)
	}
}

// Name returns the detector name
func (d *SpringBootDetectorV2) Name() string {
	return "spring-boot"
}

// SetLogSink sets the log sink for the detector
func (d *SpringBootDetectorV2) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}
