@Library('jenkins-pipeline') _

node {
  cleanWs()

  repo = "exoscale/exoip"

  try {
    def image

    dir('src') {
      stage('SCM') {
        checkout scm
      }
      updateGithubCommitStatus('PENDING', "${env.WORKSPACE}/src")
      stage('gofmt') {
        gofmt()
      }
      stage('build') {
        parallel (
          "go lint": {
            golint(repo, "cmd/exoip")
          },
          "go test": {
            test()
          },
          "go install": {
            build("exoip")
          },
          "docker": {
            image = docker()
          },
        )
      }
    }
    stage('push') {
      image.push()
    }
  } catch (err) {
    currentBuild.result = 'FAILURE'
    throw err
  } finally {
    if (!currentBuild.result) {
      currentBuild.result = 'SUCCESS'
    }
    updateGithubCommitStatus(currentBuild.result, "${env.WORKSPACE}/src")
    cleanWs cleanWhenFailure: false
  }
}

def gofmt() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.pull()
    image.inside("-u root --net=host") {
      sh 'test $(gofmt -s -d -e $(find -iname "*.go" | grep -v "/vendor/") | tee -a /dev/fd/2 | wc -l) -eq 0'
    }
  }
}

def golint(repo, ...extras) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.pull()
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh "cd /go/src/github.com/${repo} && golangci-lint run ."
      for (extra in extras) {
        sh "cd /go/src/github.com/${repo} && golangci-lint run ./${extra}"
      }
    }
  }
}

def test() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.inside("-u root --net=host") {
      sh "go test -v -mod=vendor ./..."
    }
  }
}

def build(...bins) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.inside("-u root --net=host") {
      for (bin in bins) {
        sh "cd cmd/${bin} && go install -mod=vendor"
        sh "test -e /go/bin/${bin}"
      }
    }
  }
}

def docker() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def branch = getGitBranch().replace("/", "-")
    def tag = getGitTag() ?: (branch == "master" ? "latest" : branch)
    def version = tag.replaceAll(~/^v(?=\d)/, "")
    def ref = sh("git rev-parse HEAD")
    def date = sh('date -u +"%Y-%m-%dT%H:%m:%SZ"')
    return docker.build("registry.internal.exoscale.ch/exoscale/exoip:" + tag, "--network=host --no-cache --build-arg VCS_REF=$ref --build-arg BUILD_DATE=$date --build-arg VERSION=$version .")
  }
}
