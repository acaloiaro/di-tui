# di-tui

A simple terminal UI player for [di.fm Premium](http://di.fm)

[![Packaging status](https://repology.org/badge/vertical-allrepos/di-tui.svg)](https://repology.org/project/di-tui/versions)

![App Screenshot](https://github.com/acaloiaro/di-tui/assets/3331648/5b85343f-d098-48d8-ae98-4bd1e99e0a8b)

---

This app began as di.fm player, but now supports the whole Audio Addict network

- Classical Radio
- DI.fm
- Radio Tunes
- Rock Radio
- Jazz Radio
- Zen Radio


## Run / Install

### Binary Releases

Binary releases are available in [releases](https://github.com/acaloiaro/di-tui/releases).

### `go install`

`go install github.com/acaloiaro/di-tui@latest`

### Nixpkgs (currently unstable channel only)

`di-tui` is currently in nixpkgs/nixos-unstable.

### Run the flake

```
nix run github:acaloiaro/di-tui
```

## Usage

### Authenticate

There are two authentication options

- Enter your username and password directly into `di-tui` with the `--username` and `--password` switches
- If you're justifiably uncomfortable with entering your username/password into this application, copy your "Listen Key" from (https://www.di.fm/settings) and create the following file:

#### ~/.config/di-tui/config.yml
```yml
token: <YOUR LISTEN KEY>
album_art: <BOOLEAN>
```

| key | description |
| --- | ----------- |
| token | **string** Your di.fm authentication "Listen Key" found at https://www.di.fm/settings |
| album_art |  **boolean** Enable/disable album ASCII art |

### Choose a network 

DI.fm is the default network, but other audio addict networks can be chosen with the `--network` switch. 

| switch value | network |
| --- | ----------- |
| classicalradio | Classical Radio [https://classicalradio.com](https://classicalradio.com) |
| di | DI.fm [https://di.fm](https://di.fm) |
| radiotunes | Radio Tunes [https://radiotunes.com](https://radiotunes.com) |
| rockradio |  Rock Radio [https://rockradio.com](https://rockradio.com)|
| jazzradio | Jazz Radio [https://jazzradio.com](https://jazzradio.com)|
| zenradio |  Zen Radio [https://zenradio.com](https://zenradio.com)|

### Favorites

Favorites must be edited on the web player (e.g. at [DI.fm](https://di.fm)) and cannot be
edited via this app. It used to be possible to edit favorites from this app, but DI.fm
introduced limitations on their API.

## Dependencies

### PulseAudio

Both linux and MacOS depend on pulseaudio to be running.

#### MacOS

By default, pulseaudio on MacOS runs as "root", which is not ideal. PulseAudio is best run by non-root users. By symbolically linking the pulseaudio plist file into your user's `~/Library/LaunchAgents/`, it runs as your user.

```
brew install pulseaudio
ln -s $(brew info pulseaudio | grep "/usr/local/Cellar" | awk '{print $1}')//homebrew.mxcl.pulseaudio.plist ~/Library/LaunchAgents
brew services start pulseaudio
```

#### Debian / Ubuntu

`apt install pulseaudio`

## MPRIS/D-bus support 

MPRIS is a D-Bus specification allowing media players to be controlled in a standardized way, e.g. with `playerctl`. 

`di-tui` supports a very limited set of MPRIS commands. The limited set is due to the fact that `di-tui` is a streaming audio player, and it doesn't make sense to support `next`, `previous`, `seek`, etc., because audio streams have no next or previous track; or the ability to seek forward. 

### Supported MPRIS commands

`play-pause` Toggles Play/Pause, e.g. `playerctl --player=di-tui play-pause` toggles play/pause on di-tui if it's the active player 

### Supported MPRIS metadata

`track` The currently playing track

`artist` The currently playing artist 

`status` The status of the player, e.g. `playing`, `paused`, `stopped`

`playerName` The name of the player: `di-tui`

## Configuration

### Themes

By default, `di-tui` respects your terminal's color scheme. However, there are four color settings that one can change by adding a `theme` to `config.yml`.

**Tomorrow-Night** inspired theme

```yml
theme:
  primary_color: "#81a2be"
  background_color: "#2a1f1a"
  primary_text_color: "#969896"
  secondary_text_color: "#81a2be"
```


