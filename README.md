lovebeat
========

Have you ever had a nightly backup job fail, and it took you weeks until you noticed it? Say hi to lovebeat, a zero-configuration heartbeat monitor.

Other use-cases include creating triggers for manual stuff that you haven't automated yet (like moving your backup tapes to an offsite location, watering your plants or changing sheets in your bed). It can also be used for the opposite - for finding out when things start to happen that shouldn't, like your frontend calling deprecated methods in the backend.

Lovebeat provide a lot of different APIs. We recommend the "statsd"-compatible UDP protocol when adding triggers to your software for minimum impact on your performance. A TCP protocol if you don't trust UDP and HTTP protocol for curl. And don't forget the web UI.

Installation and running
========================

If you can use docker, simply:

    $ docker run -it -p 8127:8127/udp -p 8127:8127/tcp -p 8080:8080 boivie/lovebeat

If you just want to use it (and not develop it), download the binary matching
your architecture and OS at:

https://github.com/boivie/lovebeat/releases

If you have go installed, simply:

    $ mkdir go
    $ cd go
    $ export GOPATH=`pwd`
    $ go get github.com/boivie/lovebeat
    $ cd src/github.com/boivie/lovebeat
    $ make
    $ ./lovebeat

Developers will want to install go-bindata to re-generate the assets from the
data/ directory. Then type "make" to build it.

Key Concepts
============

  * *service*
    This is the name of the "thing" you want to monitor. It's just like the
    *bucket* in graphite. It's recommended to create some sort of hierarchy,
    separated by periods. See below in the examples for some details

  * *state*
    A service can be in different states. "OK" is the state you want to keep
    your services in. You can set two timeouts, a *warning* and an *error* timeout.
    The service will change state into WARNING or ERROR when the timeouts have
    expired. Both timeouts are optional.

  * *view*
    A view shows a filtered subset of your services. You specify a regular expression
    and all services whose identifiers match this regexp will be part of the view.

    The views will inherit state from their services. If all services are OK, the
    view will be OK. But if any service is in WARNING or ERROR state, the view
    will be in WARNING or ERROR state.

  * *alarms*
    If it possible (and recommended) to have an external monitoring system check
    the status of lovebeat. There is an API endpoint, /status?view=name,
    that you can let e.g. nagios monitor. (check the 'contrib' directory for
    a provided nagios plugin)

    Another option is to let lovebeat send mails or do a HTTP POST to your web
    service whenever a view changes state. Every view allows you to set these
    alarms.

Automatic Setting of Timeouts
=============================

While you can set the error and/or warning timeout manually, they can also be
automatically calculated based on the frequency and regularity of the heartbeats.

A regular heartbeat results in a low threshold (compared to the median frequency
of the heartbeats) and an irregular heartbeat sets the threshold higher so that
it doesn't expire during normal operations.

The algorithm is rather well performing in theory and modeled (and tested) using
the bundled Jupyter Notebook bundled in the docs/ directory.

Protocol Details
================

UDP and TCP protocol
--------------------

We use the graphite protocol to trigger heartbeats and to set warning/error
timeouts. UDP is great for generating heartbeats from within your application
since the performance cost is very small and your application will not be affected
if the lovebeat server isn't running for any reason.

  * To trigger a heartbeat, send a counter value >= 0 to "<service>.beat"
  * To set a warning timeout (in seconds), set the gauge value of "<service>.warn"
  * To set an error timeout (in seconds), set the gauge value of "<service>.err"
  * To clear a value, set the timeout to -1.
  * To set the timeout to be automatically calculated, set the timeout to -2.
  * A shortcut for setting 'err' to -2 and issuing a beat (since this is a
    fairly common pattern), send a counter to "<service>.autobeat"

Examples:

    # UDP
    $ echo "invoice.mailer.beat:1|c" | nc -4u -w0 localhost 8127
    
    # TCP
    $ echo "invoice.mailer.warn:3600|g" | nc -c localhost 8127

    # TCP, setting warn to 'auto'
    $ echo -e "invoice.mailer.warn:-2|g\ninvoice.mailer.beat:1|c" | nc -c localhost 8127

    # UDP, sending a beat and auto-generating an error threshold
    $ echo "invoice.mailer.autobeat:1|c" | nc -4u -w0 localhost 8127

You can even put a statsd proxy in front of lovebeat if you don't want to send
UDP packets outside your localhost.

HTTP Protocol
-------------

There are several HTTP endpoints. The API is used by the web UI and is fairly
complete.

The most important ones include:

  * POST to /api/services/<service_id>,
    to generate a heartbeat.
  * To set a warning timeout, add the form field "warn-tmo"
  * To set an error timeout, add the form field "err-tmo"
  * To get status of all services in a view, GET /status?view=something
  * ... and some more. Not everything is unfortunately documented yet.


Examples:

    $ curl -X POST http://localhost:8080/api/services/invoice.mailer
    $ curl http://localhost:8080/status
    $ curl -d err-tmo=1800 http://localhost:8080/api/services/invoice.mailer

Web UI
------

Just point your browser to [http://localhost:8080/](http://localhost:8080)

Configuration
=============

You don't need to write a configuration file to get started (just launch
the executable), but some settings need to be specified if you want to
use advanced features such as SMTP mail notifications.

Please see the provided 'lovebeat.cfg' file where all the settings are
documented.

Note that lovebeat by default reads /etc/lovebeat.cfg but you can override
this by specifying the '-config <file>' argument when starting lovebeat. If
no configuration file is specified, sensible defaults are used.

Building a docker image from source
===================================

     $ docker run --rm -v $(pwd):/src -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder


Notable software included
=========================

 * juration.js from https://github.com/domchristie/juration

License
=======

Copyright 2014-2015 Victor Boivie <<victor@boivie.com>>

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

