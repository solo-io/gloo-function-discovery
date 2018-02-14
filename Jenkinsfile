#!/usr/bin/env groovy
def imageName = "docker.io/soloio/glue-discovery"
def imageTag = (env.BRANCH_NAME == "master") ? "latest" : env.BRANCH_NAME
podTemplate(label: 'glue-discovery-builder', 
containers: [
    containerTemplate(
        name: 'golang', 
        image: 'golang:1.9.0', 
        ttyEnabled: true, 
        command: 'cat'),
    containerTemplate(
        name: 'docker',
        image: 'docker:17.12',
        ttyEnabled: true,
        command: 'cat')
    ],
envVars: [
    envVar(key: 'IMAGE_NAME', value: imageName),
    envVar(key: 'IMAGE_TAG', value: imageTag),
    envVar(key: 'DOCKER_CONFIG', value: '/etc/docker')
    ],
volumes: [
    hostPathVolume(hostPath: '/var/run/docker.sock', mountPath: '/var/run/docker.sock'),
    secretVolume(secretName: 'soloio-docker-hub', mountPath: '/etc/docker'),
    secretVolume(secretName: 'soloio-github', mountPath: '/etc/github')
]) {

    properties([
        parameters ([
            booleanParam (
                defaultValue: false,
                description: 'Run end to end tests?',
                name: 'RUN_E2E'),
            booleanParam (
                defaultValue: false,
                description: 'Publish Docker image?',
                name: 'PUBLISH')
        ])
    ])

    node('glue-discovery-builder') {
        
        stage('Init') { 
            container('golang') {
                echo 'Setting up workspace for Go...'
                checkout scm
                sh '''
                    export OLD_DIR=$PWD
                    cp /etc/github/id_rsa $PWD
                    chmod 400 $PWD/id_rsa
                    export GIT_SSH_COMMAND="ssh -i $PWD/id_rsa -o \'StrictHostKeyChecking no\'"
                    git config --global url."git@github.com:".insteadOf "https://github.com"
                    curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep
                    mkdir -p ${GOPATH}/src/github.com/solo-io/
                    ln -s `pwd` ${GOPATH}/src/github.com/solo-io/glue-discovery
                    cd ${GOPATH}/src/github.com/solo-io/glue-discovery
                    dep ensure -vendor-only
                    rm $OLD_DIR/id_rsa
                '''
            }
        }

        stage('Build') {
            container('golang') {
                echo 'Building...'
                sh '''
                    cd ${GOPATH}/src/github.com/solo-io/glue-discovery
                    CGO_ENABLED=0 GOOS=linux go build
                '''
            }
        }

        stage('Test') {
            container('golang') {
                echo 'Testing....'
                sh '''
                    cd ${GOPATH}/src/github.com/solo-io/glue-discovery
                    go test  -race -cover `go list ./... | grep -v "e2e\\|demo"`
                '''
            }
        }

        stage('Integration') {
            if (env.BRANCH_NAME == 'master' || params.RUN_E2E) {
                container('golang') {
                    echo 'Integration tests'
                    sh '''
                        cd ${GOPATH}/src/github.com/solo-io/glue-discovery
                        echo go test ./e2e
                    ''' 
                }
            }
        }

        stage('Publish') {
            if (env.BRANCH_NAME == 'master' || params.PUBLISH) {
                container('docker') {
                    echo 'Publish'
                    sh '''
                    export HASH=`git log -n 1 --pretty=format:%h`
                    cd docker
                    cp ../glue-discovery .
                    echo ${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_NAME}:${HASH}
                    docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" -t "${IMAGE_NAME}:${HASH}" .
                    docker push "${IMAGE_NAME}:${IMAGE_TAG}"
                    docker push "${IMAGE_NAME}:${HASH}"
                    '''
                }
            }
        }
    }
}
