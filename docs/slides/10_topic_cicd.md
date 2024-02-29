## CI/CD

```{image} ../img/cicd.svg
:width: 200px
```

### CI/CD
```{revealjs-fragments}
* Plan
* Code
* Build
* Test
* Release
* Deploy
* Operate
* Monitor
```

### Pipeline
```{image} ../img/pipeline-flow.png
```

### CI/CD Tools
```{image} ../img/gitlab.svg
:width: 200px
```
```{image} ../img/jenkins.svg
:width: 200px
```

### Gitlab CI
```{revealjs-fragments}
* Pipelines
* Runners
* Variables
* Components
```

### Gitlab Pipeline
```{image} ../img/gitlab-slides1.png
:width: 500px
```
```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|6-7|9-17|19-29|31-43
---
stages:
  - build
  - test
  - deploy

default:
  image: alpine

build_a:
  stage: build
  script:
    - echo "This job builds something."

build_b:
  stage: build
  script:
    - echo "This job builds something else."

test_a:
  stage: test
  script:
    - echo "This job tests something. It will only run when all jobs in the"
    - echo "build stage are complete."

test_b:
  stage: test
  script:
    - echo "This job tests something else. It will only run when all jobs in the"
    - echo "build stage are complete too. It will start at about the same time as test_a."

deploy_a:
  stage: deploy
  script:
    - echo "This job deploys something. It will only run when all jobs in the"
    - echo "test stage complete."
  environment: production

deploy_b:
  stage: deploy
  script:
    - echo "This job deploys something else. It will only run when all jobs in the"
    - echo "test stage complete. It will start at about the same time as deploy_a."
  environment: production

```

### 
```{image} ../img/gitlab-slides2.png
:width: 200px
```
```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|9-17|19-31|33-46
---
stages:
  - build
  - test
  - deploy

default:
  image: alpine

build_a:
  stage: build
  script:
    - echo "This job builds something quickly."

build_b:
  stage: build
  script:
    - echo "This job builds something else slowly."

test_a:
  stage: test
  needs: [build_a]
  script:
    - echo "This test job will start as soon as build_a finishes."
    - echo "It will not wait for build_b, or other jobs in the build stage, to finish."

test_b:
  stage: test
  needs: [build_b]
  script:
    - echo "This test job will start as soon as build_b finishes."
    - echo "It will not wait for other jobs in the build stage to finish."

deploy_a:
  stage: deploy
  needs: [test_a]
  script:
    - echo "Since build_a and test_a run quickly, this deploy job can run much earlier."
    - echo "It does not need to wait for build_b or test_b."
  environment: production

deploy_b:
  stage: deploy
  needs: [test_b]
  script:
    - echo "Since build_b and test_b run slowly, this deploy job will run much later."
  environment: production
```

### 
```{image} ../img/gitlab-slides3.png
:width: 200px
```
```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|5-18|19-24|25-35
---
# parent: /.gitlab-ci.yml
stages:
  - triggers

trigger_a:
  stage: triggers
  trigger:
    include: a/.gitlab-ci.yml
  rules:
    - changes:
        - a/*
trigger_b:
  stage: triggers
  trigger:
    include: b/.gitlab-ci.yml
  rules:
    - changes:
        - b/*
# child: /a/.gitlab-ci.yml
stages:
  - build
  - test
  - deploy

build_a:
  stage: build
  script:
    - echo "This job builds something."

test_a:
  stage: test
  needs: [build_a]
  script:
    - echo "This job tests something."

deploy_a:
  stage: deploy
  needs: [test_a]
  script:
    - echo "This job deploys something."
  environment: production
```

### Jenkins
```{revealjs-fragments}
* Job/Pipeline
* Master(Controller)
* Node/Agent
```

### Jenkins Pipeline
```{revealjs-fragments}
* Scripted
* Declarative
```

### Scripted Pipeline

```{revealjs-code-block} groovy
---
data-line-numbers: 1-2|3-11
---
# Jenkinsfile
node {
    stage('Build') {
        //
    }
    stage('Test') {
        //
    }
    stage('Deploy') {
        //
    }
}
```

### Declarative Pipeline

```{revealjs-code-block} groovy
---
data-line-numbers: 1-2|3|4-9|6-8|10-18
---
# Jenkinsfile
pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                //
            }
        }
        stage('Test') {
            steps {
                //
            }
        }
        stage('Deploy') {
            steps {
                //
            }
        }
    }
}
```
