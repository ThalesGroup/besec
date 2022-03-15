TODO

# Contributing guidelines

### Contributing code

The GitHub Actions automatically run:

-   `make all` will build and test the code.
-   `make check` verifies that the BeSec practices are valid.

The practice format is specified in `practices/schema.json`, which you can use
to have your editor automatically validate practice definition files and
provide tooltips as you write. If you open this repo in vscode with the Red Hat
YAML plugin, this will just work with no further configuration.

Repo structure:

-   `api/`: web server implementation package
-   `cmd/`: command-line tool to run and manage the site
-   `demo/`: sample plans and practice definitions, for demos and tests
-   `deploy/`: configurations for GCP
-   `lib/`: planning library
-   `practices/`: the practices that form the input to planning
-   `store/`: database interfaces
-   `ui/`: web front-end

## Getting started

Prequisites:

-   dotnet 2.1 (required if changing the APIs).
-   goswagger (tested with v0.29)
-   firebase cli (for local development)

-   cd ui
-   npm install
-   cd ..
-   make

### Running locally

Local development is supported by the Firebase auth and firestore emulators.
Run `firebase init` from the root directory and enable both the FireStore and Auth emulators.
Then use `firebase emulators:start` to run them.

To run a local BeSec instance and populate it with some data:

```sh
# terminal 1
export FIRESTORE_EMULATOR_HOST=localhost:8088 FIREBASE_AUTH_EMULATOR_HOST=localhost:9099
./besec serve --alert-first-login=false --alert-access-request=false --disable-auth
# terminal 2
export FIRESTORE_EMULATOR_HOST=localhost:8088 FIREBASE_AUTH_EMULATOR_HOST=localhost:9099
./besec practices publish
./besec demo
```

The web interface will allow you to quickly create and log in as a new user, using the emulator.
To experiment with authorization, create a couple of users, manually authorize one like so:

```
$ ./besec users list
UID                             Email                           Display name    Provider        Status
NB7J3NG7elHbJeobxjEKdQf4OFi4    olive.olive.704@example.com     'Olive Olive'   google.com
Zn5Xz0vGWNmpPsIOfMyTHhDDcpYa    grass.raccoon.661@example.com   'Grass Raccoon' google.com
$ ./besec users authorize NB7J3NG7elHbJeobxjEKdQf4OFi4 --force # force is required if example.com isn't configured as a trusted domain
Authorized Grass Raccoon
```

Then re-launch the besec server without the `--disable-auth` flag. You can now test the frontend as both an authorized and unauthorized user.

See ui/README.md for guidance on developing the frontend.

## Pull Request Checklist

### Testing

```sh
-   firebase emulators:start
# in another terminal
-   export FIRESTORE_EMULATOR_HOST=localhost:8088
-   ./besec serve --alert-first-login=false --alert-access-request=false --port=8081 --disable-auth
# in another terminal
-   export FIRESTORE_EMULATOR_HOST=localhost:8088
-   make test
-   # equivalent to:
-   # make testgo # Go unit tests
-   # make testui # UI unit tests
-   # make testgo-integration
```

## License

All contributions must be MIT licensed.
