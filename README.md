![bazarr-sync](https://github.com/ajmandourah/bazarr-sync/assets/27051374/6c4acde4-bb9b-4172-8c67-c985c7994b28)
### Bulk sync your subtitles to your media.

Bazarr let you download subs for your titles automatically.
But if for some reason you needed to sync old subtitles, wither you need to do it because you have not synced them before or you have edited them elsewhere, you will be forced to do it one by one as there is no option to bulk sync them except a per series option which won't help if you would like to sync movies or if you have several shows.
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
sudo docker run -it \
-v ${PWD}/config.yaml:/usr/src/app/config.yaml \
--network <NETWORK_NAME> \
ghcr.io/ajmandourah/bazarr-sync:latest \
bazarr-sync sync movies
```

## Configuration
use the provided config.yaml file as a template. fill in the required fields.
either direct the flag --config to your config file or place it in the working directiory where you bazarr-sync is located.
```yaml
#config file example, please don't use quotes
###########################
#
#Address: the address of your bazarr instance. this can be either an ip address or a url (if you reverse proxy bazarr), 
#this can also be bazarr's container name if you use docker, make sure bazarr-sync instance is in the same network as bazarr and the network not the default
#docker network as name resolution won't happen there.
Address: <bazarr_address>
#
#Port: bazarrs port. by default bazarr uses 6767. in case of reverse proxy, you can use 443 or 80 as per your configuration 
Port: <port>
#
#protocol: this can be http or https
Protocol: https
#
#ApiToken: you can get this from bazarr setting>general . no quotes needed.
ApiToken: <Bazarr_api_token>
```
## Usage:

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
      --config string      config file (default is ./config.yaml)
      --golden-section     Use Golden-Section Search
  -h, --help               help for bazarr-sync
      --no-framerate-fix   Don't try to fix framerate
```
Sync all movies subtitles
```
bazarr-sync --config config.yaml sync movies
```
Sync all tv shows subtitles
```
bazarr-sync --config config.yaml sync shows
```
