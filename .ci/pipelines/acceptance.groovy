@Library('estc') _

pipeline {
  parameters {
    string(name: 'branch_specifier', defaultValue: 'refs/heads/main', description: "the Git branch specifier to build (&lt;branchName&gt;, &lt;tagName&gt;,&lt;commitId&gt;, etc.)")
  }
  agent none
  stages {
    stage('Acceptance Tests') {
      matrix {
        axes {
          axis {
            name 'ES_VERSION'
            values '8.0.0', '7.15.1'
          }
        }
        agent { label('linux && immutable && docker') }
        environment {
          HOME = "${env.JENKINS_HOME}"
          REPOSITORY = "terraform-provider-elasticstack"
          GIT_REFERENCE_REPO = "/var/lib/jenkins/.git-references/terraform-provider-elasticstack.git"
          branch_specifier = "${params?.branch_specifier}"
        }
        stages {
          stage('Checkout') {
            options { skipDefaultCheckout() }
            steps {
              estcGithubCheckout(github_org: 'elastic',
                                  repository: env.REPOSITORY,
                                  revision: env.ghprbActualCommit ?: env.CHANGE_BRANCH ?: env.BRANCH_NAME ?: env.branch_specifier,
                                  reference_repo: env.GIT_REFERENCE_REPO)
            }
          }
          stage('Tests') {
            steps {
              dir('.') {
                sh(label: 'Run ATs in docker', script: "ELASTICSEARCH_VERSION=${ES_VERSION} make docker-testacc")
              }
            }
          }
        }
      }
    }
  }
}
