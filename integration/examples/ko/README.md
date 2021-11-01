### Example: ko builder

**Note:** This example is for an upcoming Skaffold feature. Please follow the
release notes to see when the feature is available.

This is an example demonstrating building a Go app with the
[ko](https://github.com/google/ko) builder.

The included [Cloud Build](https://cloud.google.com/build/docs) configuration
file shows how users can set up a simple pipeline using `skaffold build` and
`skaffold deploy`, without having to create a custom builder image.
