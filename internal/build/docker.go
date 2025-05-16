package build

import "os/exec"

// Build runs `docker build` in repoPath tagging the imageTag.
func Build(repoPath, imageTag string) error {
	cmd := exec.Command("docker", "build", "-t", imageTag, ".")
	cmd.Dir = repoPath
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
