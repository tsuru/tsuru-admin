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

.. tsuru-command:: bs-info
   :title: Show bs container information

.. tsuru-command:: bs-env-set
   :title: Set environment variables for bs container

.. tsuru-command:: bs-upgrade
   :title: Upgrade bs image

Node management
===============

.. _tsuru_admin_docker_node_add_cmd:

.. tsuru-command:: docker-node-add
   :title: Add a new docker node

.. _tsuru_admin_docker_node_list_cmd:

.. tsuru-command:: docker-node-list
   :title: List docker nodes in cluster

.. tsuru-command:: docker-node-update
   :title: Update a docker node

.. _tsuru_admin_docker_node_remove_cmd:

.. tsuru-command:: docker-node-remove
   :title: Remove a docker node

Machine management
==================

.. _tsuru_admin_machines_list_cmd:

.. tsuru-command:: machine-list
   :title: List IaaS machines

.. _tsuru_admin_machine_destroy_cmd:

.. tsuru-command:: machine-destroy
   :title: Destroy IaaS machine

.. tsuru-command:: machine-template-list
   :title: List machine templates

.. _tsuru_admin_machine_template_add_cmd:

.. tsuru-command:: machine-template-add
   :title: Add machine template

.. tsuru-command:: machine-template-remove
   :title: Remove machine template

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

.. _tsuru_admin_platform_add_cmd:

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

