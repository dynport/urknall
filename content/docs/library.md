---
title: Library
---

# Urknall Library

The library part of the urknall is where most of the magic happens. For a
detailed information on the API of urknall have look into the [API
documentation](http://godoc.org/github/dynport/urknall). This guide will guide
you through the concepts required for using urknall.


## Targets

The target is the "host" where the commands are executed on. Currently there is
support for remote execution using SSH and running commands locally.


## Templates


## Commands


## Tasks


## Sudo Without Password

Urknall must be able to execute commands like installing packages or creating
users, which require `root` permissions. If you're not provisioning
using the `root` user the `sudo` mechanism is required. As manual entry of
passwords is tedious it is required that the user is allowed sudo without
password. This can be achieved by adding the following setting (make sure you
change the username from 'ubuntu' to whatever suits you):

	echo "ubuntu ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/90-nopassword

Now you should verify that there is no password required on running commands
with `sudo`.



