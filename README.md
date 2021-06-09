# Battlesnake Template

(Go, [Google appengine standard environment][gae])

### Required reading

- [Readme of the battlesnake Go example][starter-snake-readme] (This repo is based on that one, but with a different deployment strategy)

### Setup

1.  Clone this repo.

1. Create a new Google Cloud Project (or select an existing project).

1. [Initialize your App Engine app with your project][app-engine-nodejs-docs].

1. Enable the [App Engine Admin API][app-engine-admin-api] on your project.

1.  [Create a Google Cloud service account][sa] or select an existing one.

1.  Add the the following [Cloud IAM roles][roles] to your service account:

    - `App Engine Admin` - allows for the creation of new App Engine apps

    - `Service Account User` -  required to deploy to App Engine as service account

    - `Storage Admin` - allows upload of source code

    - `Cloud Build Editor` - allows building of source code

1.  [Download a JSON service account key][create-key] for the service account.

1.  Add the following [secrets to your repository's secrets][gh-secret]:

    - `GCP_PROJECT`: Google Cloud project ID

    - `GCP_SA_KEY`: the downloaded service account key

[starter-snake-readme]: https://github.com/BattlesnakeOfficial/starter-snake-go/blob/master/README.md
[gae]: https://cloud.google.com/appengine
[sm]: https://cloud.google.com/secret-manager
[sa]: https://cloud.google.com/iam/docs/creating-managing-service-accounts
[gh-runners]: https://help.github.com/en/actions/hosting-your-own-runners/about-self-hosted-runners
[gh-secret]: https://help.github.com/en/actions/configuring-and-managing-workflows/creating-and-storing-encrypted-secrets
[setup-gcloud]: https://github.com/google-github-actions/setup-gcloud/
[roles]: https://cloud.google.com/iam/docs/granting-roles-to-service-accounts#granting_access_to_a_service_account_for_a_resource
[create-key]: https://cloud.google.com/iam/docs/creating-managing-service-account-keys
[app-engine-admin-api]: https://console.cloud.google.com/apis/api/appengine.googleapis.com/overview
[app-engine-nodejs-docs]: https://cloud.google.com/appengine/docs/standard/nodejs/console#console