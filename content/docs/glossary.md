---
title: Urknall Glossary
layout: default
---

# Urknall Glossary
{:.no_toc}

* TOC
{:toc}


## Build

A build renders a [template](#template) to a [target](#target).


## Caching

As the [commands](#command) of a [task](#task) should only be executed once or
if one of the predecessors or itself changed, there needs to be some
information what has already been run. [Templates](#template) and
[tasks](#task) are assigned names that are concatenated for nested templatess.
These names are given with the `AddCommands`, `AddTask` and `AddTemplate`
methods. Each task has a subdirectory in the `/var/lib/urknall` directory. This
subdirectory will contain one file per command that is named after the hash of
the command. This information is then used in subsequent executions to
determine whether a command has already been executed.


## Command

A command is the entity that is actually executed on the [target](#target). The
`Command` interface has two methods: `Shell` and `Logging`. The first will
create the command as executed as string (i.e. something like `echo -n "Hello"
&& echo " World"`). The second is used to create dedicated logging output. This
is useful if a command requires multiple (maybe complicated) steps that somehow
obfuscate the intention of the command. For an example have a look into the
`cmd_file.go` file.


## Package

The `Package` interface is exposed to the user only in the `Render` method of
a [template](#template). It is used to add [tasks](#task) for a
[build](#build), either by directly giving the task (see the `AddTask` method),
adding a list of [commands](#command), or adding another [template](#template).


## Target

This is the target a template is rendered too. We decided to not use the "host"
phrase, as it is to broadly used.


## Task

A task is a list of [commands](#command) that are executed with a
[cache](#caching) in the background, i.e. only those commands are exeucted that
have not yet been run or those that follow on a changed one.


## Template

A template is a struct that implements the `Template` interface, i.e. has a
`Render` method. The struct's fields are used to configure the addition of
[tasks](#task) to a given [package](#package) in the form of [tasks](#task),
[commands](#command) or other [templates](#template).
