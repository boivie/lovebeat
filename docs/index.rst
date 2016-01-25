.. Lovebeat documentation master file, created by
   sphinx-quickstart on Thu Jan 21 19:29:44 2016.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Introducing Lovebeat
====================

Have you ever had a nightly backup job fail, and it took you weeks until you
noticed it? Say hi to lovebeat, a zero-configuration heartbeat monitor.

Other use-cases include creating triggers for manual tasks that you haven't
automated yet (like moving your backup tapes to an offsite location, watering
your plants or changing sheets in your bed). It can also be used for the
opposite - for finding out when things start to happen that shouldn't, like
your frontend calling deprecated methods in the backend.

Lovebeat provide a lot of different APIs. We recommend the "statsd"-compatible
UDP protocol when adding triggers to your software for minimum impact on your
performance. A TCP protocol if you don't trust UDP and HTTP protocol for curl.
And don't forget the web UI.

.. toctree::
   :maxdepth: 2

   installation
   getting_started
   alerters
   advanced
   api
   configuration
   license
