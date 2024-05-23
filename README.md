# GPS2MQTT

Get your GPS data from various cheap GPS trackers directly (well, via MQTT) into [Home Assistant](https://www.home-assistant.io/) without the massive load of JAVA or third party hosting.

Inspired by a desire to leave [Traccar](https://traccar.org) which, don't get me wrong is a wonderful application, just overkill for simply triggering on GPS events.

I don't currently have the desire to implement the more advance features of protocols (some support writing back to the device) but this may change in future.

## Supported Protocols

* watch
* h02
* gt06

## Configuration

Configuration consists of blocks, mqtt, meta, and protocol.

### mqtt block

Provides configuration for connecting to your MQTT server

### status block

Provides configuation for the http status endpoint

### meta blocks

Each meta block defines a known tracker ID, trackers that connect and try to communicate will not be permitted to do so unless they have a corresponding meta block
These blocks also let you define a friendly name and an icon for [Home Assistant](https://www.home-assistant.io/) to use

### protocol blocks

Let you define a listening port and some timeouts

## Sample configuration

```toml
[mqtt]
ClientName = "gps2mqtt"
Keepalive = "1m"
PingTimeout = "5s"
Username = "bob"
Password = "auntie"
Brokers = [
    "tcp://homassistant.local:1883"
]

[status]
Enabled = false
Listen = "localhost:8080"

[meta."SA*91678358119"]
Name = "Motorbike Tracker"
Icon = "mdi:motorbike"

[meta."2214050251"]
Name = "Lawnmower Tracker"
Icon = "mdi:robot-mower"

[protocol.watch]
Listen=":5093"

[protocol.h02]
Listen=":5093"
```
