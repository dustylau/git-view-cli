# ADR 0003: Separate container fixture data from the build

Status: Accepted for DEV-001

The container workflow builds a development image from Go source and tests. It mounts repository data at runtime so a fixture can be replaced without rebuilding an image layer or including anyone's Git history in the build context.

Use a new named volume for each seeded run. The seed container can write to it and the test container mounts it read-only. Host mode binds an explicitly selected checkout read-only with nested mounts disabled. The runner always uses the image ID returned by its own build and deletes only its own seeded volume. Both fixture modes share the same test command.

The Docker workflow verifies a Linux environment. The compatibility plan tracks the native runner gates separately.
