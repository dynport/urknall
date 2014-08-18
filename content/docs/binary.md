---
title: Urknall Binary
---

# Urknall Binary

The urknall binary is used to manage urknall based provisioning tools.


## Template Management

urknall has gone through quite some development untill we reached a state we
were actually quite happy with. The problem arose from the question, how
templates should be distributed.


### Integrated With The Library

In the first iteration we had the templates integrated with the library, as
separate package. This comes with two downsides. First and foremost any change
to these templates would've affected every user. Every small change to a
template would have affected every user, resulting in the execution of the
template whether desired or not. On the other hand this is not the usage
pattern we had in mind....

Templates provided by urknall should be a jump start to solve a problem, i.e. a
user should be able to use them to get up and running fast. But he must always
be able to adopt them to his special needs. We can't build templates that will
solve every single use case out there. Users know what they need and should be
able to help themselves.


### Integrated With The Binary

In a second iteration we had the templates integrated with the binary. This was
way better than the first approach. Users would import the templates needed
into their project and had all the freedom to change them according to their
needs.

The downside with this approach is that users must update the urknall binary
(and library) every time they want to update the templatess to the latest
version.


### A Binary Accessing Templates On Github

In the last iteration the templates are fetched from github, so that they are
always up to date.


## Project Management

The urknall binary can be used to create a simple basic urknall provisioning
tool. This is to help with the basic steps.

