#!/usr/bin/env groovy

def ruleset = "ci/go_ruleset2.0.yaml"
def custom_ruleset = "ci/custom_go_ruleset2.0.yaml"
def bob = "./bob/bob -r ${ruleset}"

stage('Cron Wrapper Image') {
    script {
            sh "${bob} -r ${custom_ruleset} image"
            sh "${bob} -r ${custom_ruleset} image-dr-check"
        }
}

stage('Package Cron Wrapper') {
    script {
        withCredentials([usernamePassword(credentialsId: 'SELI_ARTIFACTORY', usernameVariable: 'SELI_ARTIFACTORY_REPO_USER', passwordVariable: 'SELI_ARTIFACTORY_REPO_PASS')]) {
            sh "${bob} -r ${custom_ruleset} package"
            sh "${bob} -r ${custom_ruleset} delete-images"
        }
    }
}
