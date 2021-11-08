@Library('estc') _

pipeline {
  agent { label('linux && immutable && docker') }
  environment {
    HOME = "${env.JENKINS_HOME}"
    REPOSITORY = "terraform-provider-elasticstack"
    GIT_REFERENCE_REPO = "/var/lib/jenkins/.git-references/terraform-provider-elasticstack.git"
    GIT_CREDENTIALS = "f6c7695a-671e-4f4f-a331-acdce44ff9ba"
  }
  stages {
    stage('Checkout') {
      options { skipDefaultCheckout() }
      steps {
        estcGithubCheckout(github_org: 'elastic',
                            repository: env.REPOSITORY,
                            revision: env.ghprbActualCommit ?: env.CHANGE_BRANCH ?: env.BRANCH_NAME ?: params.branch_specifier,
                            reference_repo: env.GIT_REFERENCE_REPO,
                            credentials: env.GIT_CREDENTIALS)
      }
    }
    stage('Acceptance Tests') {
      steps {
        sh(label: 'Run ATs in docker', script: 'make docker-testacc')
      }
    }
  }
}
