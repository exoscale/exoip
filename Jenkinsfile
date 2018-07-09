@Library('jenkins-pipeline') _

node {
  cleanWs()

  repo = "exoscale/exoip"
  def image

  try {
    dir('src') {
      stage('SCM') {
        checkout scm
      }
      stage('dep') {
        godep(repo)
      }
      updateGithubCommitStatus('PENDING', "${env.WORKSPACE}/src")
      stage('Build') {
        parallel (
          "go lint": {
            golint(repo, "cmd/exoip")
          },
          "go test": {
            test(repo)
          },
          "go install": {
            build(repo, "exoip")
          },
          "docker build": {
            image = docker(repo)
          }
        )
      }
      stage('Upload') {
        image.push()
      }
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

def godep(repo) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.pull()
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh 'test `gofmt -s -d -e . | tee -a /dev/fd/2 | wc -l` -eq 0'
      sh "cd /go/src/github.com/${repo} && dep ensure -v -vendor-only"
    }
  }
}

def golint(repo, ...extras) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh "golint -set_exit_status -min_confidence 0.6 `go list github.com/${repo}/... | grep -v /vendor/`"
      sh "go vet `go list github.com/${repo}/... | grep -v /vendor/`"
      sh "cd /go/src/github.com/${repo} && gometalinter ."
      for (extra in extras) {
        sh "cd /go/src/github.com/${repo} && gometalinter ./${extra}"
      }
    }
  }
}

def test(repo) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh "cd /go/src/github.com/${repo} && go test"
    }
  }
}

def build(repo, ...bins) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      for (bin in bins) {
        sh "go install github.com/${repo}/cmd/${bin}"
        sh "test -e /go/bin/${bin}"
      }
    }
  }
}

def docker(repo) {
  def branch = getGitBranch()
  def tag = getGitTag() ?: (branch == "master" ? "latest" : branch)
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    return docker.build("registry.internal.exoscale.ch/${repo}:${tag}", "--network host .")
  }
}
