@Library('jenkins-pipeline') _

node {
  cleanWs()

  try {
    dir('src') {
      stage('SCM') {
        checkout scm
      }
      stage('dep') {
        godep()
      }
      updateGithubCommitStatus('PENDING', "${env.WORKSPACE}/src")
      stage('Build') {
        parallel (
          "go lint": {
            golint()
          },
          "go test": {
            test()
          },
          "go install": {
            build()
          }
        )
      }
      stage('Upload') {
        docker()
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

def godep() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.pull()
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/exoscale/exoip") {
      sh 'test `gofmt -s -d -e . | tee -a /dev/fd/2 | wc -l` -eq 0'
      sh 'cd /go/src/github.com/exoscale/exoip && dep ensure -v -vendor-only'
    }
  }
}
def golint() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/exoscale/exoip") {
      sh 'golint -set_exit_status github.com/exoscale/exoip'
      sh 'golint -set_exit_status github.com/exoscale/exoip/cmd/exoip'
      sh 'go vet github.com/exoscale/exoip'
      sh 'go vet github.com/exoscale/exoip/cmd/exoip'
    }
  }
}

def test() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/exoscale/exoip") {
      sh 'cd /go/src/github.com/exoscale/exoip && go test'
    }
  }
}

def build() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/exoscale/exoip") {
      sh 'go install github.com/exoscale/exoip/cmd/exoip'
      sh 'test -e /go/bin/exoip'
    }
  }
}

def docker() {
  // XXX figure out a proper tag based on the branch or tag
  tag = "latest"
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.build("registry.internal.exoscale.ch/exoscale/exoip:" + tag, "--network host .")
    image.push()
  }
}
