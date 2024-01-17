# di-tui
A simple terminal UI player for [di.fm Premium](http://di.fm)

![App Screenshot](https://github.com/acaloiaro/di-tui/assets/3331648/5b85343f-d098-48d8-ae98-4bd1e99e0a8b)

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

## Install

### Binary Releases

There are binary builds available in [releases](https://github.com/acaloiaro/di-tui/releases).

### With `go install`
`go install github.com/acaloiaro/di-tui@latest`

If `$GOPATH/bin` is not on your `$PATH` (modify accordingly for ZSH users `~/.zshrc`)
```
echo "export PATH=$PATH:$GOPATH/bin" >> ~/.bashrc
source ~/.bashrc
```
### Run with `nix run`

```
nix run github:acaloiaro/di-tui
```

## Authenticate

There are two authentication options

- Enter your username and password directly into `di-tui`
- If you're justifiably uncomfortable with entering your username/password into this application, copy your "Listen Key" from (https://www.di.fm/settings) and create the following file:

### ~/.config/di-tui/config.yml
```yml
token: <YOUR LISTEN KEY>
album_art: <BOOLEAN>
```

| key | description |
| --- | ----------- |
| token | Your di.fm authentication "Listen Key" found at https://www.di.fm/settings |
| album_art | Turn album art on or off |

## MPRIS/D-bus support 

MPRIS is a D-Bus specification allowing media players to be controlled in a standardized way, e.g. with `playerctl`. 

`di-tui` supports a very limited set of MPRIS commands. The limited set is to to the fact that `di-tui` is a streaming audio player, and it doesn't make sense to support `next`, `previous`, `seek`, etc., because audio streams have no next or previous track; or the ability to seek forward. 

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

## Run

`di-tui`

