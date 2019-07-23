# Bifrost

This is a small utility project to turn a discord server (guild) into a IRC client.

It creates a new discord channel for every respective IRC channel. It also creates a channel for every unique DM.

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
	},
	"channels": [       // array of IRC channel names
		""
	]
}
```

Sadly, for now it also requires `IRC-Channels` and `IRC-DMs` channel groups.
