# Watch protocol?

I'm not documenting the whole protocol, you can find a word document [on the traccar website](https://www.traccar.org/protocols/)

## Message Format

`[Vendor*DeviceID*Length*Content]`

 * Vendor 2 bytes
 * Unsure if DeviceID is variable
 * Length is hex encoded big endian unsigned int16 of the length of the content
 * Content always starts with a keyword defining the message (eg: LK, UD, AL)

## Messages

### Hello

Seems to be sent at first connect, documentation says it's sent every 5 minutes after that but my unit doesn't seem to.

#### From GPS

`[CS*YYYYYYYYYY*LEN*LK]` or `[CS*YYYYYYYYYY*LEN*LK,steps,rolling times,battery]` or similar.

#### Respond with

`[CS*YYYYYYYYYY*LEN*LK]`

### CCID (ICCID)

Seems to be undocumented in the documentation I found but my unit sends it once it gets a LK response

#### From GPS

`[CS*YYYYYYYYYY*LEN*CCID,xxxxxxxxxxxxxxxxxxxx]`

### Location Update

When the unit is awake it'll send this message through fairly consistently, data is really only valid if field 3 is A

#### From GPS

`[CS*YYYYYYYYYY*LEN*UD,Location Data]`

### Blind Spot Update

Honestly, no idea?

#### From GPS

`[CS*YYYYYYYYYY*LEN*UD2,Location Data]`

### Alarm

Will keep reporting alarm state if no response is sent back, might stop SMS reports?

#### From GPS

`[CS*YYYYYYYYYY*LEN*AL,Location Data]`

#### Respond with

`[CS*YYYYYYYYYY*LEN*AL]`

## Location Data

Comma seperated data.

Not all these fields are returned, at least not by my device, excuse formatting and engrish, I just copypastad

01) Date	120414	 (day month year)2014 year 4 month 12 day
02) Time	101930	 (Time minute seconds)10 hour 19 minute 30 seconds
03) If position function or not	A	A:Position V:No position
04) Latitude	22.564025	Definite in DD.DDDDDD format,The latitude is :22.564025.
05) Latitude identification	N	N represent North latitude ,S represents South latitude.
06) Longitude	113.242329	Definite in DDD.DDDDDD format ,The longitude is :113.242329.
07) Longitude	E	E represents east longitude ,W represents west longitude
08) Speed	5.21	5.21miles/hour.
09) Direction	152	The direction is in 152 degree.
10) Alititude	100	The unit is meter
11) Satellite	9	Represent the satellite number
12) GSM signal strength	100	Means the current GSM signal strength (0-100)
13) Battery	90	Means the current battery grade percentage.
14) Pedometer	1000	The steps is 1000
15) Rolling times	50	Rolling number is 50 times.
16) The terminal statement	00000000	Represents in Hexadecimal string,and the meaning isas following:The high 16bit is alarming,the low 16 bit is statement.
	Bit(from 0)        Meaning(1 effective)
	0                  Low-battery
	1                  Out of fence
	2                  In fence
	3                  Bracelet statement
	16                 SOS alarming
	17                 Low-battery alarming
	18                 Out of fence alarming
	19                 In fence alarming
	20                 Watch-take-off alarming
17) Base station number	4	Report the station number,0 is no reporting
18) Connect to station ta	1	GSM time delay
19) MCC country code	460	460 represents china
20) MNC network number	02	02 represents china mobile
21) The area code to connect base station	10133	Area code
22) Serial number to connect base station	5173	Base station serial code
23) The signal strength	100	Signal strength
24) The nearby station1  area code	10133	Area code
25) The nearby station 1 serial number	5173	Station serial number
26) The nearby base station 1 signal strength	100	Signal strength
27) The nearby station2  area code	10133	Area code
28) The nearby station 2 serial number	5173	Station serial number
29) The nearby base station 2 signal strength	100	Signal strength
30) The nearby station3  area code	10133	Area code
31) The nearby station 3 serial number	5173	Station serial number
32) The nearby base station 3 signal strength	100	Signal strength
