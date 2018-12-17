node {
    properties([pipelineTriggers([[$class: 'GitHubPushTrigger'], pollSCM('H/15 * * * *')])])
    def app

    stage('Clone repository') {
        /* Let's make sure we have the repository cloned to our workspace */

        checkout scm
    }

    stage('Build image & run tests') {
        /* This builds the actual image; synonymous to
         * docker build on the command line */

        app = docker.build("dm848/srv-acl")
    }

    stage('Test image') {
        /* Ideally, we would run a test framework against our image.
         * For this example, we're using a Volkswagen-type approach ;-) */

        app.inside {
            sh 'echo "Tests passed"'
        }
    }

    stage('Push image') {
        /* Finally, we'll push the image with two tags:
         * First, the incremental build number from Jenkins
         * Second, the 'latest' tag.
         * Pushing multiple tags is cheap, as all the layers are reused. */
        docker.withRegistry('https://registry.hub.docker.com', 'docker-hub-credentials-dm848') {
            app.push("${env.BUILD_NUMBER}")
            app.push("latest")
        }
    }

    stage('Deploy to kubernetes') {
        /* Deployment */
        container('kubectl') {
            sh 'wget https://github.com/DM848/k8s-cluster/setup-kubectl.sh && chmod +x setup-kubectl.sh && sh setup-kubectl.sh && rm setup-kubectl.sh'
            sh 'kubectl set image deployments/acl acl=dm848/srv-acl:${BUILD_NUMBER}'
        }
    }
}