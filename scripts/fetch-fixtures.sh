#!/bin/bash
set -euo pipefail

# Base directory for test fixtures
BASE_DIR="testdata/e2e"

# Create base directory if it doesn't exist
mkdir -p "$BASE_DIR"

# Function to shallow clone a repository
clone_fixture() {
    local repo=$1
    local target=$2
    local branch=${3:-main}

    echo "Cloning $repo into $target..."
    rm -rf "$target"
    git clone --depth 1 --branch "$branch" "https://github.com/$repo.git" "$target"
    echo "spring-boot" > "$target/EXPECTED_STACK"
}

# Function to create a simple Express.js app
create_express_app() {
    local target="$BASE_DIR/node/express_hello"
    echo "Creating Express.js app in $target..."
    mkdir -p "$target"
    
    # Create package.json
    cat > "$target/package.json" << EOF
{
  "name": "express-hello",
  "version": "1.0.0",
  "main": "index.js",
  "dependencies": {
    "express": "^4.18.2"
  }
}
EOF

    # Create index.js
    cat > "$target/index.js" << EOF
const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
  res.send('Hello World!');
});

app.listen(port, () => {
  console.log(\`App listening at http://localhost:\${port}\`);
});
EOF

    echo "node" > "$target/EXPECTED_STACK"
}

# Function to create a plain Java project
create_plain_java() {
    local target="$BASE_DIR/negative/plain_java"
    echo "Creating plain Java project in $target..."
    mkdir -p "$target/src/main/java"
    
    # Create pom.xml without Spring Boot
    cat > "$target/pom.xml" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>com.example</groupId>
    <artifactId>plain-java</artifactId>
    <version>1.0-SNAPSHOT</version>

    <properties>
        <maven.compiler.source>11</maven.compiler.source>
        <maven.compiler.target>11</maven.compiler.target>
    </properties>

    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>4.13.2</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
</project>
EOF

    # Create a simple Java class
    cat > "$target/src/main/java/Main.java" << EOF
public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}
EOF

    echo "unknown" > "$target/EXPECTED_STACK"
}

# Function to create a simple React app
create_react_app() {
    local target="$BASE_DIR/react/cra_template"
    echo "Creating React app in $target..."
    mkdir -p "$target/src"
    
    # Create package.json
    cat > "$target/package.json" << EOF
{
  "name": "react-app",
  "version": "0.1.0",
  "private": true,
  "dependencies": {
    "@types/node": "^16.18.0",
    "@types/react": "^18.0.0",
    "@types/react-dom": "^18.0.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-scripts": "5.0.1",
    "typescript": "^4.9.5"
  },
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "test": "react-scripts test",
    "eject": "react-scripts eject"
  }
}
EOF

    # Create index.tsx
    cat > "$target/src/index.tsx" << EOF
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
EOF

    # Create App.tsx
    cat > "$target/src/App.tsx" << EOF
import React from 'react';

function App() {
  return (
    <div>
      <h1>Hello, World!</h1>
    </div>
  );
}

export default App;
EOF

    echo "react" > "$target/EXPECTED_STACK"
}

# Clone Spring Boot Kotlin demo
clone_fixture "sdeleuze/spring-boot-kotlin-demo" "$BASE_DIR/spring/kotlin_gradle_demo" "main"

# Clone Spring PetClinic
clone_fixture "spring-projects/spring-petclinic" "$BASE_DIR/spring/petclinic_maven" "main"

# Create plain Java project (negative case)
create_plain_java

# Create React app
create_react_app

# Create Express.js app
create_express_app

echo "All test fixtures have been fetched and set up!" 