## Alexa Smart Home Virtual Buttons

## Background

It's sometimes useful to have Alexa announce things and have these triggered
from real world actions (e.g. "Garage door is open", "Washing machine has
finished").

If you've used [Home Assistant](https://www.home-assistant.io/) then you
might know of an unofficial [Alexa plugin](https://github.com/keatontaylor/alexa_media_player).  This is a pretty clever piece of work, with a lot of
reverse-engineered API discovery.  And, yes, it can make these announcements.

However I have a problem with it; because it's all unofficial APIs it
sometimes break.  And, in my case, weekly it would lock out my Amazon
account with multiple authentication failures.  If Amazon would just publish
an official API then this wouldn't be a problem, but... _sigh_.

So since I didn't need most of the I wondered if there was another way

## The official way

Alexa recently got the ability to trigger routines from Smart Home
sensors; specifically Door Sensors and Motion Sensors.  If a Door Sensor
switched to "open" then a routine can be triggered.  And one of the actions
can be to announce... anything you want it to say.

So all we need to do is add some "virtual buttons" (masquerading as
door sensors) which can be triggered by a HTTPS request and we're done!

Some searching around the 'net found a couple of solutions out there; I
haven't tried them but I'm guessing (based on their description) that
they work in a similar way.

If you're willing to pay someone else to do the heavy lifting then these
are possible solutions:

[Sinric](https://sinric.pro): First 3 devices free, $3/device/yr after that.

[Virtual Buttons](virtualbuttons.com): First device free; 2 devices $12/yr, 5 devices $24/year, 10 devices $36/year.

Note: I'm not endorsing these at all; they're just what I found.  They also
likely to be 100 times more polished than this (pretty user interfaces).
This code is "raw".

If you're willing to put in the effort in setting this up then we can
use the code in this repo to create a Smart Home skill that presents as
door sensors, and allows you to use them as Alexa routine triggers.

## Design

A Smart Home skill has to be hosted in Amazon Lambda, so I decided to make
this "Cloud Native".  We use a DynamoDB table to hold the button
definitions (and some other additional data), the Lambda Function (written
in [Go](https://golang.org)) and a HTTP API Gateway to allow for control.

The Lambda hosting and DynamoDB are "always free" within the limits we
will be using it.  The API Gateway costs $1 for 1,000,000 requests in a month.
If we called it once every 5 minutes for a whole month that's under 10,000
requests, which may cost 1c.  More realistly this will cost zero.  The
[COSTS](COSTS.txt) provides a detailed analyse.

## Installation
Please read the [Installation documents](install/README.md)

## Usage
Please read the [Usage](usage/README.md)

