Advanced Topics
===============

Automatic Setting of Timeouts
-----------------------------

While you can set the timeout manually, they can also be automatically
calculated based on the frequency and regularity of the heartbeats.

A regular heartbeat results in a low threshold (compared to the median frequency
of the heartbeats) and an irregular heartbeat sets the threshold higher so that
it doesn't expire during normal operations.

The algorithm is rather well performing in theory and modeled (and tested) using
the bundled Jupyter_ Notebook.


Monitoring
----------

Lovebeat is designed to be resistant to environmental disturbances but it can
still fail if e.g. the machine it's running on is degraded or if the network is
experiencing problems. It's a very good idea to monitor Lovebeat so that you
are confident that it's monitoring your services correctly.

External Monitoring
~~~~~~~~~~~~~~~~~~~

It is easy to have an external monitoring system find out if lovebeat if
healthy. There is an API endpoint, ``/status`` for that purpose.

Calling it will result in the following response:

.. code-block:: bash

    $ curl http://localhost:8080/status
    num_ok 4
    num_error 2
    has_error true
    good false

If you call it with the ``Accept`` HTTP header set to ``application/json``, the
following will be the response instead:

.. code-block:: bash

    $ curl -H "Accept: application/json" http://localhost:8080/status
    {
      "num_ok": 4,
      "num_error": 2,
      "has_error": true,
      "good": false
    }

``good`` will be **true** only if there are no services in **ERROR** state.

By specifying a ``?view=name`` query parameters, only services that are members
of the provided view will be used.

You can let e.g. nagios_ monitor it. There is a
provided nagios plugin in the contrib/ directory.

Lovebeat Monitoring
~~~~~~~~~~~~~~~~~~~

For more detailed monitoring, you can have two (or more) instances of Lovebeat
monitor each other. By having one or several ``notify`` sections in the
configuration file, you can specify a URL to which Lovebeat will post its
heartbeats.

.. code-block:: ini

    [[notify]]
    lovebeat = "http://some-other-host:8080"

Logging
-------

Lovebeat prints its logs to stderr. If you want the logs to be sent to the local
syslog service, add the command line switch ``-syslog``.

You can also increase the verbosity of the logs by adding ``-debug``.

Metrics reporting
-----------------

Lovebeat can send metrics to a statsd_ proxy using the UDP protocol, to allow
them to be shown in  e.g. graphite_, influxdb_ or similar.

You will get some health information about Lovebeat itself, such as the time
it takes to save its database, and also status information (as gauges) of
all services and views. This allows you to correlate service status with other
metrics you collect.

Simply specify a server and the prefix that Lovebeat will use for all metrics
in the lovebeat configuration file:

.. code-block:: ini

    [metrics]
    server = "localhost:8125"
    prefix = "lovebeat"

Behind a reverse proxy
----------------------

Lovebeat can be located behind a reverse proxy and properly handle that it's
served from a different path than the root path. Please keep in mind that the
websocket functionality requires a proxy server with proper support for them.

In nginx_, this would be a working configuration:

.. code-block:: nginx

    location /monitoring/lovebeat/ {
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_pass http://localhost:8080/;
    }

Database on S3
--------------

When designing Lovebeat, a key decision was to build a solution with as few
dependencies to other systems as possible since those systems can fail as well.
Having the database on a separate SQL server is then something we have opted out
from, but instead having a local file based database.

That works well as long as the host Lovebeat is running on is healthy and the
disk where the database is located on isn't corrupted or disappears. When
deploying Lovebeat on a transient host, such as on an auto-scaling instance on
Amazon Web Services, this will cause problems as the disk isn't persistent if
the service is restarted.

To support this common use-case, Lovebeat supports downloading and uploading its
database to an Amazon S3 bucket. On startup, Lovebeat will download the file
from the S3 storage, and every time the database is saved (defaults to once
per minute and when Lovebeat exits), the database will be uploaded to the same
S3 bucket.

To enable this, configure the database as follows:

.. code-block:: ini

    [database]
    filename = "lovebeat.db"
    interval = 60
    remote_s3_url = "s3://bucket-name/path/to/lovebeat.db"
    remote_s3_region = "eu-west-1"


.. _nagios: https://www.nagios.org/
.. _jupyter: http://jupyter.org/
.. _statsd: https://github.com/etsy/statsd
.. _graphite: http://graphite.wikidot.com/
.. _influxdb: https://influxdata.com/
.. _nginx: https://www.nginx.com/
