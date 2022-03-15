# BeSec

<tagline>

Run `./besec` to get help on the available commands.

# Build

Run `make release`.

There are various dependencies for building and deploying, notably Go and NPM. See `build.Dockerfile` which specifies a
build environment with all of the dependencies installed.

To build using this container, create an image with `docker build -f build.Dockerfile -t builder .` and use it by
mounting the repository directory within the container.

# Run

You will need Google cloud credentials available on the host.

Invoke it with arguments like this: `./besec serve -v --no-emulator`.
You can then browse to http://localhost:8080.

See the `deploy` directory for the configuration to deploy the site to Google Cloud Platform.

To run the application locally, run the Firestore emulator - e.g. `gcloud beta emulators firestore start --host-port localhost:8088` - and then set the environment variable
when running the app: `FIRESTORE_EMULATOR_HOST=localhost:8088 ./besec ...`

To run under docker, you first need to build the binary - the main Dockerfile is used for deployment, not building.

-   `make release`
-   `docker build . -t besec`
-   `docker run -p8080:8080 --env GOOGLE_APPLICATION_CREDENTIALS=/gcp.json --mount type=bind,source=/home/<username>/.config/gcloud/application_default_credentials.json,destination=/gcp.json besec serve -v --no-emulator`

# Configure

Config options can be set using any of:

-   commandline flags (run `./besec help` for a comprehensive listing of all options);
-   in the `config.yaml` file, which can be overridden by entries in a `config.local.yaml` file. Keys have the same name as the commandline flags.;
-   environment variables, names are the same as the commandline flags, in upper case, prefixed with `BESEC_`, and with dashes replaced with underscores.

A base set of practices are included, however you may wish to add, remove, or modify these. You can replace the base
set entirely, using the `--practices-dir` option to point to your own definitions.

Alternatively you can build on the base set using the `--local-practices-dir` option. In your local practices directory,
files with names that match base practices will be applied on top of them:

-   Change anything (except the practice ID) - the delta file doesn't need to be a complete practice,
    for example it could just contain a `name:` entry, to re-title the practice.
-   To change which tasks are included in a practice, override the `tasks: []` list with just the task IDs you want to
    include (referencing IDs from both the base and delta `taskDefinitions`).
-   To remove an existing base practice, create a matching file with the entry `remove: true` in it.
    To add a practice, create a file which doesn't match any base practices.
-   Any named field in the file can be overridden in the delta, but text-level changes aren't supported (e.g. you have to
    replace the entire description of a task to change one word.)

For an example of all of this in practice, look in the `demo/` practice directories, and then publish the definitions with: `./besec practices publish -v --practices-dir=demo/practices --local-practices-dir=demo/local-practices --schema-file=practices/schema.json`.

# Manage

Depending on the operation, administrators need one of two forms of authentication:

-   Permission to impersonate a service account - set the terraform `cli-admins` variable appropriately.
    Note that this service account should _not_ have any API keys generated for it.
-   A current OAuth identity token. A log-in flow isn't implemented for this, so you need to extract it from a current session with the site!

Administrators can manage the site by running:

-   `besec practices` - to publish practice definitions. You'll need to do this the first time you run the app and then whenever you change the definitions. This is also automatically run as part of the master pipeline. Uses service account impersonation.
-   `besec users` - to view users and authorize new users. Uses service account impersonation.
-   `besec demo` - to create some demo projects, plans, and practice definitions. Requires an API token.

## Users

Users can log in using any of the identity providers configured in the config file under `auth`, and can in
theory be any of the types of provider supported by [Google Identity
Platorm](https://cloud.google.com/identity-platform/docs/concepts-authentication) (including any OIDC and SAML
provider). In practice, only Google and SAML identity providers are currently supported.

If the provider config has `whitelisted: true` set, the user will have access to the system. Otherwise, they can still
log in but will not get access until and admin authorizes them. If alerts have been configured (`alerts` and one of the
`slack-webhook-*` options has been set), admins are notified when a new user tries to log in but is not authorized.

The `trusted-domains` configuration entry is a convenience to users of the CLI to prevent accidentally adding users
from untrusted domains.
