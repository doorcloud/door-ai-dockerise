import org.gradle.api.Plugin
import org.gradle.api.Project

class SpringBootPlugin : Plugin<Project> {
    override fun apply(project: Project) {
        project.plugins.apply("org.springframework.boot")
    }
} 