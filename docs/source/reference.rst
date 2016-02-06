Reference
~~~~~~~~~

Managing remote tsuru server endpoints
======================================

.. tsuru-command:: target

.. tsuru-command:: target-add
   :title: Add a new target
.. tsuru-command:: target-list
   :title: List existing targets
.. tsuru-command:: target-set
   :title: Set a target as current
.. tsuru-command:: target-remove
   :title: Removes an existing target

Check current version
=====================

.. tsuru-command:: version


Container management
====================

All the **container** commands below only exist when using the docker
provisioner.

.. _tsuru_admin_container_move_cmd:

.. tsuru-command:: container-move
   :title: Moves single container

.. _tsuru_admin_containers_move_cmd:

.. tsuru-command:: containers-move
   :title: Moves all containers from on node

.. _tsuru_admin_containers_rebalance_cmd:

.. tsuru-command:: containers-rebalance
   :title: Rebalance containers in nodes

bs management
=============

.. _tsuru_admin_bs_management:

bs-info
-------

.. highlight:: bash

::

    $ tsuru-admin bs-info

This command displays the current configuration of bs, including environment
variables and image.

bs-env-set
----------

.. highlight:: bash

::

    $ tsuru-admin bs-env-set <NAME=value> [NAME=value]... [-p/--pool poolname]

This command sets environment variables used when starting bs (big sibling)
container.

If the `standard bs image <https://github.com/tsuru/bs>`_ is being used, it's
possible to find which environment variables can be configured in `bs readme
file <https://github.com/tsuru/bs#environment-variables>`_.

bs-upgrade
----------

.. highlight:: bash

::

    $ tsuru-admin bs-upgrade

This command upgrades the bs image. You can check the current image with the
``bs-info`` command.

Node management
===============

.. _tsuru_admin_docker_node_add_cmd:

docker-node-add
---------------

.. highlight:: bash

::

    $ tsuru-admin docker-node-add [param_name=param_value]... [--register]

This command add a node to your docker cluster. By default, this command will
call the configured IaaS to create a new machine. Every param will be sent to
the IaaS implementation.

You should configure in **tsuru.conf** the protocol and port for IaaS be able
to access your node (`you can see it here <config.html#iaas-configuration>`_).

If you want to just register an docker node, you should use the --register
flag with an **address=http://your-docker-node:docker-port**

Parameters have special meaning
+++++++++++++++++++++++++++++++

* ``iaas=<iaas name>`` Which iaas provider should be used, if not set tsuru will use
  the default iaas specified in tsuru.conf file.

* ``template=<template name>`` A machine template with predefined parameters,
  additional parameters will override template ones. See
  :ref:`machine-template-add <tsuru_admin_machine_template_add_cmd>` command.

.. _tsuru_admin_docker_node_list_cmd:

docker-node-list
----------------

.. highlight:: bash

::

    $ tsuru-admin docker-node-list [-f/--filter <metadata>=<value>]

This command list all nodes present in the cluster. It will also show you metadata
associated to each node and the IaaS ID if the node was added using tsuru builtin
IaaS providers.

Using the ``-f/--filter`` flag, the user is able to filter the nodes that
appear in the list based on the key pairs displayed in the metadata column.
Users can also combine filters with multiple listings of ``-f``:

docker-node-update
------------------

.. highlight:: bash

::

    $ tsuru-admin docker-node-update <address> [param_name=param_value...] --disable

This command modifies metadata associated to a docker node. If a parameter is set
to an empty value, it will be removed from the node's metadata.

Using the ``--disable`` flag, the node will be disabled in scheduler. It means
this node will not receive any containers.


.. highlight:: bash

::

    $ tsuru-admin docker-node-list -f pool=mypool -f LastSuccess=2014-10-20T15:28:28-02:00

.. _tsuru_admin_docker_node_remove_cmd:

docker-node-remove
------------------

.. highlight:: bash

::

    $ tsuru-admin docker-node-remove <address> [--destroy] --no-rebalance

This command removes a node from the cluster and rebalance the containers in
node to others nodes in cluster. Optionally it also destroys the created IaaS
machine if the ``--destroy`` flag is passed.

Using the ``--no-rebalance`` flag the node will be removed without rebalance.

.. _tsuru_admin_platform_add_cmd:

Machine management
==================

.. _tsuru_admin_machines_list_cmd:

machine-list
------------

.. highlight:: bash

::

    $ tsuru-admin machine-list

This command will list all machines created using ``docker-node-add`` and a IaaS
provider.

.. _tsuru_admin_machine_destroy_cmd:

machine-destroy
---------------

.. highlight:: bash

::

    $ tsuru-admin machine-destroy <machine id>

This command will destroy a IaaS machine based on its ID.

machine-template-list
---------------------

.. highlight:: bash

::

    $ tsuru-admin machine-template-list

This command will list all templates created using ``machine-template-add``.

.. _tsuru_admin_machine_template_add_cmd:

machine-template-add
--------------------

.. highlight:: bash

::

    $ tsuru-admin machine-template-add <name> <iaas> <param>=<value>...

This command creates a new machine template to be used with ``docker-node-add``
command. This template will contain a list of parameters that will be sent to the
IaaS provider.

machine-template-remove
-----------------------

.. highlight:: bash

::

    $ tsuru-admin machine-template-remove <name>

This command removes a machine template by name.

Pool management
===============

pool-add
---------------

.. highlight:: bash

::

    $ tsuru-admin pool-add <pool>

This command adds a new pool (cluster).

pool-list
----------------

.. highlight:: bash

::

    $ tsuru-admin pool-list

This command list available pools.

pool-remove
------------------

.. highlight:: bash

::

    $ tsuru-admin pool-remove <pool> [-y]

This command removes a pool.

The -y flag assume "yes" as answer to all prompts and run non-interactively.

pool-teams-add
---------------------

.. highlight:: bash

::

    $ tsuru-admin pool-teams-add <pool> <teams>

This command adds one or more teams to a poll. You can add one or more teams at once.

pool-teams-remove
------------------------

.. highlight:: bash

::

    $ tsuru-admin pool-teams-remove <pool> <teams>

This command removes one or more teams from a pool. You can remove one or more teams at once.

Healer
======

docker-healing-list
-------------------

.. highlight:: bash

::

    $ tsuru-admin docker-healing-list [--node] [--container]

This command will list all healing processes started for nodes or containers.

Platform management
===================

.. warning::

    All the **platform** commands below only exist when using the docker
    provisioner.

platform-add
------------

.. highlight:: bash

::

    $ tsuru-admin platform-add <name> [--dockerfile]

This command allow you to add a new platform to your tsuru installation.
It will automatically create and build a whole new platform on tsuru server and
will allow your users to create apps based on that platform.

The --dockerfile flag is an URL to a dockerfile which will create your platform.

.. _tsuru_admin_platform_update_cmd:

platform-update
---------------

.. highlight:: bash

::

    $ tsuru-admin platform-update <name> [-d/--dockerfile]

This command allow you to update a platform in your tsuru installation.
It will automatically rebuild your platform and will flag apps to update
platform on next deploy.

The --dockerfile flag is an URL to a dockerfile which will update your platform.

platform-remove
---------------

.. highlight:: bash

::

    $ tsuru-admin platform-remove <platform name> [-y]

This command allow you to remove a platform. This command will not
remove a platform that is used by an application.

The -y flag assume "yes" as answer to all prompts and run non-interactively.

Plan management
===============

.. _tsuru_admin_plan_create:

plan-create
-----------

::

    $ tsuru-admin plan-create <name> -c/--cpu-share cpushare [-m/--memory memory] [-s/--swap swap] [-d/--default]

This command creates a new plan for being used when creating new apps.

The ``--cpushare`` flag defines a relative amount of cpu share for units created
in apps using this plan. This value is unitless and relative, so specifying the
same value for all plans means all units will equally share processing power.

The ``--memory`` flag defines how much physical memory a unit is able to use, in
bytes.

The ``--swap`` flag defines how much virtual swap memory a unit is able to use, in
bytes.

The ``--default`` flag sets this plan as the default plan. It means this plan will
be used when creating an app without explicitly setting a plan.


plan-remove
-----------

::

    $ tsuru-admin plan-remove <name>

This command removes an existing plan, it will no longer be available for newly
created apps. However, this won't change anything for existing apps that were
created using the removed plan. They will keep using the same value amount of
resources described by the plan.

User management
===============

user-list
---------

::

    $ tsuru-admin user-list

This command list all users in tsuru.

Quota management
================

Quotas are handled per application and user. Every user has a quota number for
applications. For example, users may have a default quota of 2 applications, so
whenever a user tries to create more than two applications, he/she will receive
a quota exceeded error. There are also per applications quota. This one limits
the maximum number of units that an application may have.

**tsuru-admin** can be used to see and change quota data.

app-quota-change
----------------

.. highlight:: bash

::

    $ tsuru-admin app-quota-change <app-name> <new-limit>

Changes the limit of units that an app can have. The new limit must be an
integer, it may also be "unlimited".

user-quota-change
-----------------

.. highlight:: bash

::

    $ tsuru-admin user-quota-change <user-email> <new-limit>

Changes the limit of apps that a user can create. The new limit must be an
integer, it may also be "unlimited".

app-quota-view
--------------

.. highlight:: bash

::

    $ tsuru-admin app-quota-view <app-name>

Displays the current usage and limit of the given app.

user-quota-view
---------------

.. highlight:: bash

::

    $ tsuru-admin user-quota-view <user-email>

Displays the current usage and limit of the user.

Another commands
================

.. _tsuru_admin_app_shell_cmd:


app-shell
---------

.. highlight:: bash

::

    $ tsuru-admin app-shell [container-id] -a myapp

This command opens a remote shell inside container, using the API server as a proxy.
You can access an app container just giving app name. Also, you can access a specific container from this app too.
The user may specify part of the ID of the container. For example:

.. highlight:: bash

::

    $ tsuru app-info -a myapp
    Application: tsuru-dashboard
    Repository: git@54.94.9.232:tsuru-dashboard.git
    Platform: python
    Teams: admin
    Address: tsuru-dashboard.54.94.9.232.xip.io
    Owner: admin@example.com
    Deploys: 1
    Units:
    +------------------------------------------------------------------+---------+
    | Unit                                                             | State   |
    +------------------------------------------------------------------+---------+
    | 39f82550514af3bbbec1fd204eba000546217a2fe6049e80eb28899db0419b2f | started |
    +------------------------------------------------------------------+---------+
    $ tsuru-admin app-shell 39f8 -a myapp
    Welcome to Ubuntu 14.04 LTS (GNU/Linux 3.13.0-24-generic x86_64)
    ubuntu@ip-10-253-6-84:~$

log-remove
----------

.. highlight:: bash

::

    $ tsuru-admin log-remove [--app appname]

This command removes the application log from the tsuru database.

fix-containers
--------------

.. highlight:: bash

::

    $ tsuru-admin fix-containers

In some cases, like when a node is restarted, information about the containers
can be outdated in tsuru database, because docker changes the container
exposed port when the container is restarted.

This command verify if has a container with wrong data stored in the database
and fix this information.

app-unlock
----------

.. highlight:: bash

::

    $ tsuru-admin app-unlock -a <app-name> [-y]

Forces the removal of an app lock.
Use with caution, removing an active lock may cause inconsistencies.

router-list
-----------

.. highlight:: bash

::

    $ tsuru-admin router-list

List all routers available for plan creation.

