Advanced Topics
===============

Automatic Setting of Timeouts
-----------------------------

While you can set the error and/or warning timeout manually, they can also be
automatically calculated based on the frequency and regularity of the heartbeats.

A regular heartbeat results in a low threshold (compared to the median frequency
of the heartbeats) and an irregular heartbeat sets the threshold higher so that
it doesn't expire during normal operations.

The algorithm is rather well performing in theory and modeled (and tested) using
the bundled Jupyter_ Notebook.


Monitoring
----------

If it recommended to have an external monitoring system check the status of
lovebeat. There is an API endpoint, `/status` for that purpose.

Calling it will result in the following response:

.. code-block:: bash

    $ curl http://localhost:8080/status
    num_ok 4
    num_warning 0
    num_error 2
    has_warning false
    has_error true
    good false

`good` will be **true** only if there are no services in **WARNING** or
**ERROR** state.

By specifying a `?view=name` query parameters, only services that are members
of the provided view will be used.

You can let e.g. nagios_ monitor it. There is a
provided nagios plugin in the contrib/ directory.

Logging
-------

Lovebeat prints its logs to stderr. If you want the logs to be sent to the local
syslog service, add the command line switch ``-syslog``.

You can also increase the verbosity of the logs by adding ``-debug``.

.. _nagios: https://www.nagios.org/
.. _jupyter: http://jupyter.org/
