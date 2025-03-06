# Noti release

This is the internal process I go through to release a version of Noti. I'm
just writing this down for myself.

## Tests

Make sure latest dev is green on CI.

https://github.com/variadico/noti/actions?query=workflow%3ATesting

## Increment version

* docs/CHANGELOG.md
* docs/man/noti.1.md
* docs/man/noti.yaml.5.md

Don't forget to `make man` to generate the updated man pages.

## Merge to main

```sh
git checkout main
git merge dev --ff-only
git push origin main
```

## Tests

Make sure latest main is green on CI.

https://github.com/variadico/noti/actions?query=workflow%3ATesting

## Double check

Fix anything that might have broken like CI or URLs in docs. Last chance to
change anything.

## Tag release

Once everything is ready, tag the release.

```sh
git tag 1.2.3
git push origin 1.2.3
```

## Edit GitHub release information

* Hopefully, when you pushed, GitHub Actions automatically created a release
  draft and uploaded tarballs
* Go to https://github.com/variadico/noti/actions?query=workflow%3ARelease and
  delete `noti.darwinrelease`, it's temporary junk. It's fine.
* Add CHANGELOG notes to release body
* Publish release
