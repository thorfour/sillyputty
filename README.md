[![Docker Repository on Quay](https://quay.io/repository/thorfour/sillyputty/status "Docker Repository on Quay")](https://quay.io/repository/thorfour/sillyputty)

# sillyputty
Plugin server for slack slash commands

Sillyputty is a plugin server that will respond to slack slash commands.

## Download

`docker pull quay.io/thorfour/sillyputty`

## Run

`docker run -d -p 80:80 -p 443:443 -v /plugins:/plugins quay.io/thorfour/sillyputty /server -host <url> -email <support_email>`

This will start a TLS server that will respond to slack slash commands dynamically using the plugins provided in the `/plugins` directory.

An example plugin can be found [here](https://github.com/thorfour/trapperkeeper)

The plugin needs to have the same name as the slash command name. For example for the pick plugin above the command would have to be `/pick` for it to find the plugin named pick

A plugin needs to have a function called with the signature of `Handler(url.Values) (string, error)` for the server to work.

## Build from source

`make docker`
