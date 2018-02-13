#!/usr/bin/env groovy
def imageName = "docker.io/soloio/glue-discovery:" + ((env.BRANCH_NAME == "master") ? "latest" : env.BRANCH_NAME)
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
    secretEnvVar(key: 'AWS_ACCESS_KEY_ID', secretName: 'aws-e2e-secret', secretKey: 'keyid'),
    secretEnvVar(key: 'AWS_SECRET_ACCESS_KEY', secretName: 'aws-e2e-secret', secretKey: 'secretkey'),
    envVar(key: 'AWS_REGION', value: 'us-east-1'),
    envVar(key: 'IMAGE_NAME', value: imageName),
    envVar(key: 'DOCKER_CONFIG', value: '/etc/docker')
    ],
volumes: [
    hostPathVolume(hostPath: '/var/run/docker.sock', mountPath: '/var/run/docker.sock'),
    secretVolume(secretName: 'soloio-docker-hub', mountPath: '/etc/docker')
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
                    go get -u github.com/golang/dep/cmd/dep
                    mkdir ${GOPATH}/src/github.com/solo-io/
                    ln -s `pwd` ${GOPATH}/src/github.com/solo-io/glue-discovery
                    cd ${GOPATH}/src/github.com/solo-io/glue-discovery
                    dep ensure -vendor-only
                '''
            }
        }

        stage('Build') {
            container('golang') {
                echo 'Building...'
                sh '''
                    cd ${GOPATH}/src/github.com/solo-io/glue-discovery
                    dep status
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
                    cd docker
                    cp ../glue-discovery .
                    echo ${IMAGE_NAME}
                    docker build -t "${IMAGE_NAME}" .
                    docker push "${IMAGE_NAME}"
                    '''
                }
            }
        }
    }
}
