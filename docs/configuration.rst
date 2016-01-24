Configuration
=============

You don't need to write a configuration file to get started (just launch the
executable), but some settings need to be specified if you want to use advanced
features such as SMTP mail notifications.

Note that lovebeat by default reads ``/etc/lovebeat.cfg`` but you can override
this by specifying the ``-config <file>`` argument when starting lovebeat. If
no configuration file is specified, sensible defaults are used.

Please see the provided ``lovebeat.cfg`` file where all the settings are
documented as well.

.. code-block:: ini

    ## Every section and key is documented, and the default values are
    ## provided here (commented out).

    ##
    ## The database stores information about all services and views. It's
    ## in one single file and it's safely rewritten on save, which it does
    ## when exiting the program as well as every minute while running. This
    ## can be changed by the 'interval' setting.
    #
    #[database]
    #filename = "lovebeat.db"
    #interval = 60


    ##
    ## UDP listener, in statsd format
    #
    #[udp]
    #listen = ":8127"

    ##
    ## TCP listener, in statsd format
    #
    #[tcp]
    #listen = ":8127"

    ##
    ## TCP listener, for the dashboard and the HTTP API
    #
    #[http]
    #listen = ":8080"

    ##
    ## SMTP settings, for the mail alerter.
    #
    #[mail]
    #server = "localhost:25"
    #from = "lovebeat@example.com"

    ##
    ## Metrics reporting to a statsd proxy, using the UDP protocol.
    ## Note that this one is by default disabled, but can be enabled
    ## by specifying a server address and port, e.g. "localhost:8125"
    #
    #[metrics]
    #server = ""
    #prefix = "lovebeat"

    ##
    ## Configuration of the logfile where events are logged. An empty
    ## or unset path disables the logging.
    #
    #[eventlog]
    #path = "/var/log/lovebeat/events.json"
    #mode = 644
