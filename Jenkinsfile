@Library('jenkins-pipeline') _

node {
  cleanWs()

  repo = "exoscale/exoip"

  try {
    dir('src') {
      stage('SCM') {
        checkout scm
      }
      stage('gofmt') {
        gofmt(repo)
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
          }
        )
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

def gofmt(repo) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.pull()
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh 'test `gofmt -s -d -e . | tee -a /dev/fd/2 | wc -l` -eq 0'
    }
  }
}

def golint(repo, ...extras) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.pull()
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh "golint -set_exit_status -min_confidence 0.3  `go list github.com/${repo}/... | grep -v /vendor/`"
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
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh "go test -v -mod=vendor ./..."
    }
  }
}

def build(repo, ...bins) {
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
