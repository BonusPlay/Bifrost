# Bifrost

This is a small utility project to turn a discord server (guild) into a IRC client. It joins all IRC channels with same
 name as discord channels from `IRC-Channels` category. It also creates a channel for every unique DM under `IRC-DMs` category.
 To join/leave IRC channel just add/remove discord channel with same name (without `#` prefix, since discord channel names
 cannot contain that symbol, so if you want to join `#archlinux` add a `archlinux` channel under `IRC-Channels`).

## Setup

Currently it requires a `bifrost.cfg` to be in same folder as executable. Structure:

```
{
	"irc": {
		"cert": "",     // path to certificate file
		"key": ""       // path to key file
	},
	"discord": {
		"token": "",    // your bot token
		"guild": ""     // your guild (discord server) ID
	}
}
```

Sadly, for now it also requires `IRC-Channels` and `IRC-DMs` channel groups.
