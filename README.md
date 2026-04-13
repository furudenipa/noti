# NOTICE: noti is moving to codeberg.org/roble/noti 🚚

---

github.com/variadico/noti is moving to [codeberg.org/roble/noti](https://codeberg.org/roble/noti)

Effective immediately, the latest development source code will be hosted at
https://codeberg.org/roble/noti/

Over the past few years, I've been trying out different code forges. This last
year, I found that I've been using my Codeberg account the most. The thing I
value the most about Codeberg is that it's an Open Source platform (based on
Forgejo). I also like the UI/UX Codeberg has to offer.

**Migration plan**

- [x] Create repo at [codeberg.org/roble/noti](https://codeberg.org/roble/noti)
- [x] Setup CI at [codeberg.org/roble/noti](https://codeberg.org/roble/noti)
- [x] Copy issues to [codeberg.org/roble/noti/issues](https://codeberg.org/roble/noti/issues)
- [x] Copy releases to [codeberg.org/roble/noti/releases](https://codeberg.org/roble/noti/releases)
- [x] Resolve open pull requests on github.com/variadico/noti
- [x] Post notice on github.com/variadico/noti
- [x] Disable pull requests at github.com/variadico/noti
- [ ] Update consumer URLs to point to [codeberg.org/roble/noti](https://codeberg.org/roble/noti)
- [ ] Disable issues at github.com/variadico/noti
- [ ] Archive/stop updating github.com/variadico/noti
- [ ] Release noti 3.9.0 at [codeberg.org/roble/noti](https://codeberg.org/roble/noti)

**Timeline**

I expect the entire process to take months, both because of life and also to
give people time to learn about this change.


**Other notes**

The maintainer is not changing. I'm [variadico on
GitHub](https://github.com/variadico/noti), but [roble on
Codeberg](https://codeberg.org/roble). I'm the same person. This isn't a supply
chain hack.

I'll update the readme as necessary with any updates.

---

# noti

[![status-badge](https://ci.codeberg.org/api/badges/16710/status.svg)](https://ci.codeberg.org/repos/16710)

Monitor a process and trigger a notification.

Never sit and wait for some long-running process to finish. Noti can alert you
when it's done. You can receive messages on your computer or phone.

![macOS Banner Notification]

## Services

Noti can send notifications on a number of services.

| Service    | macOS | Linux | Windows |
| ---------- |:-----:|:-----:|:-------:|
| Banner     | ✔     | ✔     | ✔       |
| Speech     | ✔     | ✔     | ✔       |
| BearyChat  | ✔     | ✔     | ✔       |
| Keybase    | ✔     | ✔     | ✔       |
| Mattermost | ✔     | ✔     | ✔       |
| Pushbullet | ✔     | ✔     | ✔       |
| Pushover   | ✔     | ✔     | ✔       |
| Pushsafer  | ✔     | ✔     | ✔       |
| Simplepush | ✔     | ✔     | ✔       |
| Slack      | ✔     | ✔     | ✔       |
| Telegram   | ✔     | ✔     | ✔       |
| Zulip      | ✔     | ✔     | ✔       |
| Twilio     | ✔     | ✔     | ✔       |
| GChat      | ✔     | ✔     | ✔       |
| Chanify    | ✔     | ✔     | ✔       |
| Bark       | ✔     | ✔     | ✔       |
| ntfy       | ✔     | ✔     | ✔       |

Check the [screenshots] directory to see what the notifications look like on different platforms.

## Installation

Install noti with these commands.

**AUR**

```shell
# Directly
git clone https://aur.archlinux.org/noti.git
cd noti
makepkg -sirc

# Or with a helper such as yay, paru...
```

**Nix**

```shell
nix-shell -p noti
```

**Homebrew**

```shell
brew install noti
```

**Scoop**

```shell
scoop bucket add main
scoop install main/noti
```

**Download pre-built binary**

```shell
(
    VER=3.8.0
    OS=linux
    ARCH=amd64
    curl -OL https://codeberg.org/roble/noti/releases/download/$VER/noti$VER.$OS-$ARCH.tar.gz
    tar xf noti$VER.$OS-$ARCH.tar.gz
)
```

Or download it with your browser from the [latest release] page.

### From source

If you want to build from the source, then build like this.

```shell
# build binary
make build
```

## Configuration

Noti reads configuration from a YAML file.

It uses the following search order (unless overridden by the --file flag):

- `./.noti.yaml`
- Then:
  - If `$XDG_CONFIG_HOME` is set:
    - `$XDG_CONFIG_HOME/noti/noti.yaml`
  - Otherwise:
    - `$HOME/.config/noti/noti.yaml`

There is an example configuration file, and a JSON schema for it, in the
[docs](https://codeberg.org/roble/noti/src/branch/main/docs/) directory.

- [docs/noti.example.yaml](https://codeberg.org/roble/noti/src/branch/main/docs/noti.example.yaml)
- [docs/noti.schema.json](https://codeberg.org/roble/noti/src/branch/main/docs/noti.schema.json)

## Examples

Just put `noti` at the beginning or end of your regular commands. For more details, check the [docs].

Display a notification when `tar` finishes compressing files.

```sh
noti tar -cjf music.tar.bz2 Music/
```

Add `noti` after a command, in case you forgot at the beginning.

```sh
clang foo.c -Wall -lm -L/usr/X11R6/lib -lX11 -o bizz; noti
```

If you already started a command but forgot to use `noti`, then you can do this to get notified when that process' PID disappears.

```sh
noti --pwatch 1234
```

You can also press `ctrl+z` after you started a process. This will temporarily suspend the process, but you can resume it with `noti`.

```
$ dd if=/dev/zero of=foo bs=1M count=2000
^Z
zsh: suspended  dd if=/dev/zero of=foo bs=1M count=2000
$ fg; noti
[1]  + continued  dd if=/dev/zero of=foo bs=1M count=2000
2000+0 records in
2000+0 records out
2097152000 bytes (2.1 GB, 2.0 GiB) copied, 12 s, 175 MB/s
```

Additionally, `noti` can send a message piped from stdin with `-`.

```sh
$ make test 2>&1 | tail --lines 5 | noti -t "Test Results" -m -
```

[macos banner notification]: https://codeberg.org/roble/noti/raw/branch/main/docs/screenshots/macos_banner.png
[screenshots]: https://codeberg.org/roble/noti/src/branch/main/docs/screenshots
[latest release]: https://codeberg.org/roble/noti/releases/latest
[docs]: https://codeberg.org/roble/noti/src/branch/main/docs/noti.md
