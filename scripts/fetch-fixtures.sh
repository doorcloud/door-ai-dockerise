#!/bin/bash

set -e

# Create directories if they don't exist
mkdir -p testdata/e2e/spring_positive/{kotlin_gradle_demo,maven_multimodule,gradle_groovy_simple,gradle_kts_multi} testdata/e2e/non_spring/plain_java

# Clone Spring Boot Kotlin demo
if [ ! -d "testdata/e2e/spring_positive/kotlin_gradle_demo/.git" ]; then
    git clone --depth=1 --filter=blob:none https://github.com/sdeleuze/spring-boot-kotlin-demo testdata/e2e/spring_positive/kotlin_gradle_demo
fi

# Clone Spring Boot multi-module Maven project
if [ ! -d "testdata/e2e/spring_positive/maven_multimodule/.git" ]; then
    git clone --depth=1 --filter=blob:none https://github.com/deepaksrivastav/spring-boot-multimodule testdata/e2e/spring_positive/maven_multimodule
fi

# Clone Spring Boot Gradle Groovy simple project
if [ ! -d "testdata/e2e/spring_positive/gradle_groovy_simple/.git" ]; then
    git clone --depth=1 --filter=blob:none https://github.com/orbartal/Smallest-spring-boot-sample testdata/e2e/spring_positive/gradle_groovy_simple
fi

# Clone Spring Boot Gradle Kotlin DSL multi-module project
if [ ! -d "testdata/e2e/spring_positive/gradle_kts_multi/.git" ]; then
    git clone --depth=1 --filter=blob:none https://github.com/daggerok/spring-boot-gradle-kotlin-dsl-example testdata/e2e/spring_positive/gradle_kts_multi
fi

# Clone non-Spring Java project
if [ ! -d "testdata/e2e/non_spring/plain_java/.git" ]; then
    git clone --depth=1 --filter=blob:none https://github.com/iluwatar/java-design-patterns testdata/e2e/non_spring/plain_java
fi 