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

.. tsuru-command:: pool-add
   :title: Add a new pool

.. tsuru-command:: pool-update
   :title: Update pool attributes

.. tsuru-command:: pool-remove
   :title: Remove a pool

.. tsuru-command:: pool-teams-add
   :title: Add team to a pool

.. tsuru-command:: pool-teams-remove
   :title: Remove a team from a pool

Healer
======

.. tsuru-command:: docker-healing-list
   :title: List latest healing events

Platform management
===================

.. warning::

    All the **platform** commands below only exist when using the docker
    provisioner.

.. _tsuru_admin_platform_add_cmd:

.. tsuru-command:: platform-add
   :title: Add a new platform

.. _tsuru_admin_platform_update_cmd:

.. tsuru-command:: platform-update
   :title: Update an existing platform

.. tsuru-command:: platform-remove
   :title: Remove an existing platform


Plan management
===============

.. _tsuru_admin_plan_create:

.. tsuru-command:: plan-create
   :title: Create a new plan

.. tsuru-command:: plan-remove
   :title: Remove an existing plan

.. tsuru-command:: router-list
   :title: List available routers


Quota management
================

Quotas are handled per application and user. Every user has a quota number for
applications. For example, users may have a default quota of 2 applications, so
whenever a user tries to create more than two applications, he/she will receive
a quota exceeded error. There are also per applications quota. This one limits
the maximum number of units that an application may have.

**tsuru-admin** can be used to see and change quota data.


.. tsuru-command:: app-quota-change
   :title: Change application quota

.. tsuru-command:: user-quota-change
   :title: Change user quota

.. tsuru-command:: app-quota-view
   :title: View application quota

.. tsuru-command:: user-quota-view
   :title: View user quota

Other commands
==============

.. tsuru-command:: app-unlock
   :title: Unlock an application
