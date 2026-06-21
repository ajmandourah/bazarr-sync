

<img height="100" alt="Screenshot_2026-06-21_183000-removebg-preview" src="https://github.com/user-attachments/assets/d6c24b41-5214-47b8-a1f4-d7c1601050a6" />

### Bulk sync your subtitles to your media.

Bazarr let you download subs for your titles automatically.
But if for for some reason you needed to sync old subtitles, wither you need to do it because you have not synced them before or you have edited them elsewhere, you will be forced to do it one by one as there is no option to bulk sync them except a per series option which won't help if you would like to sync movies or if you have several shows.
This cli tool help you achieve that by utilizing bazarr's api.

## Installation

### Go

- Install go in your system. this can be done through here. https://go.dev/doc/install
- After installation in a terminal install the module

```
go install github.com/ajmandourah/bazarr-sync/cmd/bazarr-sync@latest

```

make sure your go path is included in your path. you should be able to use the command directly with `bazarr-sync` or `bazarr-sync.exe` in windows.

### Docker

pull the image
```
sudo docker pull ghcr.io/ajmandourah/bazarr-sync:latest

```

create or copy the `config.yaml` file from the example folder. edit it to your settings. for docker you can use the bazarr container name if you have bazarr in a bridged network (not the default docker network). change the network name in the command.
Run the command in the same folder where the `config.yaml` is located. change the command to your desired functionfor example `bazarr-sync sync shows`
```
sudo docker run -it --rm \
-v ${PWD}/config.yaml:/usr/src/app/config.yaml \
--network <NETWORK_NAME> \
ghcr.io/ajmandourah/bazarr-sync:latest \
bazarr-sync sync movies
```

## Configuration

### The easy way
Just launch bazarr-sync, you will be prompted to provide bazarr address and bazarr api key. your setting will be saved in the same directory.

### The traditional way
use the provided config.yaml file as a template. fill in the required fields.
either direct the flag --config to your config file or place it in the working directiory where you bazarr-sync is located.

```yaml
bazarr_url: https://bazarr.example.com
bazarr_token: your_api_token_here
```

## Usage

### The TUI
just launch bazarr-sync without any arguments. in the TUI you can search , select multiple subtitles and send the request to send them to bazarr without the need the go through bazarr address. perfect when you know what do you wanna do . 

### The CLI
Scriptable option for all your need. 

```
Make sure to create a config.yaml file including your configuration in it.
Use the provided config file as a template.

Usage:
  bazarr-sync [command]

Examples:
bazarr-sync --config config.yaml sync movies --no-framerate-fix

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  sync        Sync subtitles to the audio track of the media

Flags:
      --config string          config file (default is ./config.yaml)
      --golden-section         Use Golden-Section Search
      --lang string            Specify language code to sync (e.g., "en", "fr", "de", "ar"). Use two-letter ISO 639-1 codes.
      --list                   list your media with their respective Radarr/Sonarr id.
  -h, --help                   help for bazarr-sync
      --no-framerate-fix       Don't try to fix framerate
      --retry-count int        Number of retries for failed syncs (exponential backoff) (default 3)
      --retry-delay duration   Base delay between retries, multiplied by attempt number (default 2s)
```

Sync all movies subtitles
```
bazarr-sync --config config.yaml sync movies
```

Sync all tv shows subtitles
```
bazarr-sync --config config.yaml sync shows
```

## Language Selection

The `--lang` flag allows you to filter which subtitles to sync by specifying a language code. Use standard two-letter ISO 639-1 language codes (e.g., `en`, `fr`, `de`, `ar`, `es`, `pt`, `it`, `ru`, `zh`, `ja`, `ko`).

```
# Sync only English subtitles
bazarr-sync --config config.yaml sync movies --lang en

# Sync only French subtitles for shows
bazarr-sync --config config.yaml sync shows --lang fr

# Combine with other flags
bazarr-sync --config config.yaml sync movies --lang de --no-framerate-fix
```

When no `--lang` flag is provided, all subtitles are processed regardless of language.

## Syncing specific movie/show subtitle

The functionality to enable syncing specefic movies/shows are added. to do so follow these steps:
- use the `--list` flag to view a list of your Shows/Movies with their respective sonarr/radarr ids. the output will shows as follows
```
Title                                                                                               | RadarrId
--------------------------------------------------------------------------------------------------------------------
3 Body Problem                                                                                      | 1304
The Apothecary Diaries                                                                              | 1043
As a Reincarnated Aristocrat, I'll Use My Appraisal Skill To Rise in the World                      | 953
Avatar: The Last Airbender (2024)                                                                   | 1341
The Banished Former Hero Lives as He Pleases                                                        | 961
```
- note the ids of your desired shows/movies to be synced
- use the usual sync command and add the flag `--radarr-id` or `--sonarr-id`
- PROFIT

Example:
```
bazarr-sync --config config.yaml sync shows --sonarr-id 1302,953,961
```

## Syncing both movies and shows

You can sync both movies and shows in the same time. what I recommend is using tmux and run the tool in 2 windows. this will assure that you won't loose progress.

## Resuming Interrupted Syncs

The tool now supports resuming sync operations after interruptions. If a sync process is interrupted (e.g., due to a network issue or manual stop), you can resume from the last processed item using the `--continue-from` flag.

```
# Resume movie sync from a specific Radarr ID
bazarr-sync --config config.yaml sync movies --continue-from 1234

# Resume show sync from a specific Sonarr ID
bazarr-sync --config config.yaml sync shows --continue-from 5678
```

## Retry Mechanism

The tool includes a built-in retry mechanism for failed sync operations:
- Default: 3 retries with exponential backoff
- Configurable via `--retry-count` and `--retry-delay` flags
- Retries help handle temporary network issues or server load problems

```
# Custom retry settings
bazarr-sync --config config.yaml sync movies --retry-count 5 --retry-delay 1s
```

## Docker Network Requirements

When running bazarr-sync in Docker, ensure the container can reach your Bazarr instance:
- Use a bridged network (not the default Docker network) for proper DNS resolution
- Specify the network name with `--network` flag
- Use the Bazarr container name as the hostname in your config
