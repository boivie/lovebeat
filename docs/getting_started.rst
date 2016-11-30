Getting started
===============

Concepts
--------

It's really helpful if you understand the different concepts in lovebeat.

Services send "heartbeats" with regular intervals to lovebeat. If they for some
reason stop sending these heartbeats, lovebeat will react to this and update 
the service state, which in may trigger and alarm, which in turn will trigger 
alerts, such as sending e-mails.

And you can monitor it all in the Lovebeat web UI.

So let's break it down a bit:

Services
--------

This is the name of the "thing" you want to monitor. You can choose anything
as the name, and it typically looks like "myapp.mailers.invoice" with periods
as the delimiter.

As you grow and have a lot of services to monitor, it's good to have some
sort of hierarchy. It's up to you to choose one.

States
------

A service can be in different states. **OK** is the state you want to keep
your services in. You can set a timeout so that the service will change state
into **ERROR** if the service hasn't issued a beat within that period of time.

A service can also be muted. This will move it into the **MUTED** state, and then
it will not trigger any alerts or cause alarms to be in **ERROR**.

Alarms
------

An alarm contains a filtered subset of your services. You specify a matching
pattern and all services whose identifiers match this pattern will be part of
the alarm.

This is an example of an alarm called "backup-jobs" that match all services
starting with "backup."

.. code-block:: ini

    [[alarms]]
    name = "backup-jobs"
    pattern = "backup.*"

Alarms also have states. If all services within the alarm are **OK**, the alarm
will be **OK**. But if any service is in **ERROR** state, the alarm will
transition into the **ERROR** state.

Alarms can be automatically created based on the service names, which is a
powerful feature when your service names have a structure.

Say that you have an application running on three servers (alpha, beta and
delta), and the application provides two heartbeats, ".healthcheck" and
".background-job-1". The complete list of services will thus be:

 * application-name.alpha.healthcheck
 * application-name.alpha.background-job-1
 * application-name.beta.healthcheck
 * application-name.beta.background-job-1
 * application-name.delta.healthcheck
 * application-name.delta.background-job-1

By having an alarm configuration such as:

.. code-block:: ini

    [[alarms]]
    name = "server-$name"
    pattern = "application-name.$name.*"

You will then end up with three alarms, "server-alpha" including the services
"application-name.alpha.healthcheck" and "application-name.alpha.background-job-1"
and similar for "server-beta" and "server-delta".

For more advanced pattern matching, use ``includes`` and ``excludes`` to specify a list
of patterns. If any pattern in ``includes`` match, and no pattern in ``excludes`` match,
the service will be part of the alarm. Example as below:

.. code-block:: ini

    [[alarms]]
    name = "source-$name"
    includes = ["source.$name.*", "old-source.$name.*"]
    excludes = ["source.deprecated.*"]

Alerts
------

When an alarm changes state (to *ERROR* or *OK*), it can trigger alerts
that e.g. sends and e-mail (through SMTP or Mailgun_), posts a Slack_ message
to your team's channel, sends an outgoing webhook or runs a shell script.

Web UI
------

Just point your browser to http://localhost:8080/ to see the current status
of all your services and alarms.

.. _Slack: https://slack.com/
.. _Mailgun: https://mailgun.com/
