lovebeat
========

Have you ever had a nightly backup job fail, and it took you weeks until you noticed it? Say hi to lovebeat, a zero-configuration heartbeat monitor.

Other use-cases include creating triggers for manual stuff that you haven't automated yet (like moving your backup tapes to an offsite location, watering your plants or changing sheets in your bed). It can also be used for the opposite - for finding out when things start to happen that shouldn't, like your frontend calling deprecated methods in the backend.

Lovebeat provide a lot of different APIs. We recommend the "statsd"-compatible UDP protocol when adding triggers to your software for minimum impact on your performance. A TCP protocol if you don't trust UDP and HTTP protocol for curl. And don't forget the web UI.

Documentation
=============

For installation and usage instructions, please head over to our documentation
at http://lovebeat.readthedocs.org/

Notable software included
=========================

 * juration.js from https://github.com/domchristie/juration

License
=======

Copyright 2014-2015 Victor Boivie <<victor@boivie.com>>

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
